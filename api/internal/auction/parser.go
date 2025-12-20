package auction

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/adriein/tibia-char/pkg/vendor"
	"github.com/gocolly/colly/v2"
	"github.com/rotisserie/eris"
)

const (
	BaseAuctionListURL = "https://www.tibia.com/charactertrade/?subtopic=currentcharactertrades"
	AuctionIDParam     = "auctionid"
)

type AuctionListHtmlParser struct {
	worldRepository WorldRepository
	tibiaAPI        *vendor.TibiaApi
	collector       *colly.Collector
}

func NewAuctionListHtmlParser(api *vendor.TibiaApi, wr WorldRepository, c *colly.Collector) *AuctionListHtmlParser {
	return &AuctionListHtmlParser{tibiaAPI: api, worldRepository: wr, collector: c}
}

func (p *AuctionListHtmlParser) GetLinks() (AuctionLinkSet, error) {
	set := make(AuctionLinkSet)

	worlds, err := p.tibiaAPI.GetWorlds()

	if err != nil {
		return set, eris.Wrap(err, "Failed to fetch worlds from Tibia API")
	}

	for _, world := range worlds {
		_, err := p.worldRepository.GetOrCreate(world)

		if err != nil {
			return set, err
		}

		if err := p.scrapeWorld(world, set); err != nil {
			return set, eris.Wrapf(err, "Failed to scrape world %s", world)
		}
	}

	return set, nil
}

func (p *AuctionListHtmlParser) GetTotalCurrentAuctions() (int, error) {
	var errors []error
	var totalCurrentAuctions int = 0

	p.collector.OnHTML("td[class=PageNavigation]", func(e *colly.HTMLElement) {
		htmlExtractedText := e.Text

		parts := strings.Split(htmlExtractedText, ": ")

		if len(parts) < 2 {
			err := eris.Errorf("String format is unexpected: %s", htmlExtractedText)
			errors = append(errors, err)

			return
		}

		numberStr := parts[1]

		cleanStr := strings.ReplaceAll(numberStr, ",", "")

		resultInt, err := strconv.Atoi(cleanStr)

		if err != nil {
			err := eris.Errorf("Error converting to integer: %s", err.Error())
			errors = append(errors, err)

			return
		}

		totalCurrentAuctions = resultInt
	})

	p.collector.Visit(BaseAuctionListURL)

	if len(errors) > 0 {
		var b strings.Builder

		for _, err := range errors {
			b.WriteString(fmt.Sprintln(err.Error()))
		}

		return totalCurrentAuctions, eris.Errorf("Error getting total current auctions: %s", b.String())
	}

	return totalCurrentAuctions, nil
}

func (p *AuctionListHtmlParser) scrapeWorld(world string, set AuctionLinkSet) error {
	for page := 1; ; page++ {
		links, err := p.scrapeAuctionListPage(world, page)

		if err != nil {
			return eris.Wrapf(err, "Failed to scrape page %d", page)
		}

		if len(links) == 0 {
			break
		}

		newLinksAdded := 0

		for _, link := range links {
			auctionID, err := p.extractAuctionID(link)

			if err != nil {
				return err
			}

			if set.Has(auctionID) {
				continue
			}

			set.Set(auctionID, link)

			newLinksAdded++
		}

		if newLinksAdded == 0 {
			break
		}
	}

	return nil
}

func (p *AuctionListHtmlParser) scrapeAuctionListPage(world string, page int) ([]string, error) {
	var result []string
	var scrapeErr error

	p.collector.OnHTML("div[class=AuctionLinks]", func(e *colly.HTMLElement) {
		e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
			if href := el.Attr("href"); href != "" {
				result = append(result, href)
			}
		})
	})

	p.collector.OnError(func(r *colly.Response, err error) {
		scrapeErr = eris.Wrapf(err, "scraping error for world %s page %d: status %d", world, page, r.StatusCode)
	})

	targetURL := fmt.Sprintf("%s&filter_world=%s&currentpage=%d", BaseAuctionListURL, world, page)

	if err := p.collector.Visit(targetURL); err != nil {
		return nil, eris.Wrapf(err, "failed to visit %s", targetURL)
	}

	if scrapeErr != nil {
		return nil, scrapeErr
	}

	return result, nil
}

func (p *AuctionListHtmlParser) extractAuctionID(link string) (int, error) {
	parsedLink, err := url.Parse(link)

	if err != nil {
		return 0, eris.Wrapf(err, "Failed to parse URL: %s", link)
	}

	auctionIDStr := parsedLink.Query().Get(AuctionIDParam)

	auctionID, err := strconv.Atoi(auctionIDStr)

	if err != nil {
		return 0, eris.Wrapf(err, "Failed to convert auction ID '%s' to int", auctionIDStr)
	}

	return auctionID, nil
}

type AuctionHtmlParser struct {
	collector *colly.Collector
}

func NewAuctionHtmlParser(c *colly.Collector) *AuctionHtmlParser {
	return &AuctionHtmlParser{
		collector: c,
	}
}

func (p *AuctionHtmlParser) Parse(auctionId int, link string) (*AuctionDTO, error) {
	dto := AuctionDTO{
		AuctionId: auctionId,
		Link:      link,
	}

	var parseErrors []error

	p.collector.OnError(func(r *colly.Response, err error) {
		parseErrors = append(parseErrors, eris.Errorf("Failed to fetch %s: status %d", r.Request.URL, r.StatusCode))
	})

	p.collector.OnHTML("div[class=Auction]", func(e *colly.HTMLElement) {
		e.ForEach("div[class]", func(_ int, ch *colly.HTMLElement) {
			switch ch.Attr("class") {
			case "AuctionHeader":
				if err := p.parseAuctionHeader(e, &dto); err != nil {
					parseErrors = append(parseErrors, err)
				}

			case "AuctionBody":
				if err := p.parseAuctionBody(e, &dto); err != nil {
					parseErrors = append(parseErrors, err)
				}
			}
		})
	})

	if err := p.collector.Visit(link); err != nil {
		parseErrors = append(parseErrors, eris.Wrap(err, "Visit failed"))
	}

	if len(parseErrors) > 0 {
		return &dto, eris.Errorf("Parsing failed with %d errors: %v", len(parseErrors), parseErrors)
	}

	return &dto, nil
}

func (p *AuctionHtmlParser) parseAuctionHeader(e *colly.HTMLElement, dto *AuctionDTO) error {
	dto.CharName = e.ChildText("div[class=AuctionCharacterName]")
	dto.CharWorld = e.ChildText("a[href]")

	auctionHeader := e.Text

	level, err := p.extractLevel(auctionHeader)

	if err != nil {
		return eris.Wrap(err, "Failed to extract level")
	}

	dto.CharLevel = level

	dto.CharVocation = p.extractVocation(auctionHeader)
	dto.CharGender = p.extractGender(auctionHeader)

	return nil
}

func (p *AuctionHtmlParser) parseAuctionBody(e *colly.HTMLElement, dto *AuctionDTO) error {
	var errors []error

	e.ForEachWithBreak("div", func(_ int, ch *colly.HTMLElement) bool {
		classes := strings.Split(ch.Attr("class"), " ")

		section := classes[len(classes)-1]

		switch section {
		case "AuctionOutfit":
			dto.ImgUrl = ch.ChildAttr("img[class=AuctionOutfitImage]", "src")

		case "AuctionItemsViewBox":
			ch.ForEach("div[title]", func(_ int, itemViewBoxCh *colly.HTMLElement) {
				imgTitle := itemViewBoxCh.Attr("title")
				imgLink := itemViewBoxCh.ChildAttr("img", "src")

				dto.FeaturedItems = append(dto.FeaturedItems, ImgDisplay{Name: imgTitle, Link: imgLink})
			})

		case "ShortAuctionData":
			ch.ForEachWithBreak("div", func(_ int, sAuctionDataCh *colly.HTMLElement) bool {
				section := sAuctionDataCh.Attr("class")

				switch section {
				case "ShortAuctionDataValue":
					rawDate := sAuctionDataCh.Text

					normDate := strings.ReplaceAll(rawDate, "\u00a0", " ")

					dateCET, err := time.Parse("Jan 02 2006, 15:04 MST", normDate)

					if err != nil {
						errors = append(errors, eris.Errorf("Error parsing auction date: %s", err.Error()))

						return false
					}

					dateUTC := dateCET.In(time.UTC)

					if dto.AuctionStartTime.IsZero() {
						dto.AuctionStartTime = dateUTC

						break
					}

					dto.AuctionEndTime = dateUTC

				case "ShortAuctionDataBidRow":
					selector := sAuctionDataCh.DOM.Children()

					rawBid := strings.ReplaceAll(selector.Find("b").Text(), ",", "")

					bid, err := strconv.Atoi(rawBid)

					if err != nil {
						errors = append(errors, eris.Errorf("Error converting bid to int: %s", err.Error()))

						return false
					}

					dto.Bid = bid

					return false
				}

				return true
			})

		case "SpecialCharacterFeatures":
			ch.ForEach("div", func(_ int, spcfCh *colly.HTMLElement) {
				dto.Featured = append(dto.Featured, spcfCh.Text)
			})
		}

		return true
	})

	if len(errors) > 0 {
		return eris.Errorf("Auction body parsing errors: %v", errors)
	}

	return nil
}

func (p *AuctionHtmlParser) extractLevel(auctionHeader string) (int, error) {
	headerParts := strings.Split(auctionHeader, "|")

	levelStringHeader := headerParts[0]

	levelStringParts := strings.Split(levelStringHeader, ":")

	return strconv.Atoi(strings.TrimSpace(levelStringParts[1]))
}

func (p *AuctionHtmlParser) extractVocation(auctionHeader string) string {
	headerParts := strings.Split(auctionHeader, "|")

	levelStringHeader := headerParts[1]

	levelStringParts := strings.Split(levelStringHeader, ":")

	return strings.TrimSpace(levelStringParts[1])
}

func (p *AuctionHtmlParser) extractGender(auctionHeader string) string {
	headerParts := strings.Split(auctionHeader, "|")

	return strings.TrimSpace(headerParts[2])
}
