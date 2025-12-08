package scrap

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/gocolly/colly/v2"
	"github.com/rotisserie/eris"
)

type Service struct {
	logger *log.Logger
}

func NewService(logger *log.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

func (s *Service) ScrapBazaar() error {
	s.logger.Println("Start Scrap Bazaar")

	now := time.Now()

	auctionLinkSet, err := s.getCurrentAuctionLinks()

	for auctionId, link := range auctionLinkSet {
		s.getCharAuctionDetails(auctionId, link)
	}

	if err != nil {
		return err
	}

	s.logger.Printf("Finished Scrapping in %s", time.Since(now))

	return nil
}

func (s *Service) scrapAuctionListPage(c *colly.Collector, world string, page int) ([]string, error) {
	var result []string

	c.OnHTML("div[class=AuctionLinks]", func(e *colly.HTMLElement) {
		e.ForEach("a[href]", func(_ int, e *colly.HTMLElement) {
			charDetailLink := e.Attr("href")

			result = append(result, charDetailLink)
		})
	})

	c.Visit(fmt.Sprintf("https://www.tibia.com/charactertrade/?subtopic=currentcharactertrades&filter_world=%s&currentpage=%d", world, page))

	return result, nil
}

func (s *Service) getTotalCurrentAuctions(c *colly.Collector) (int, error) {
	var errors []error
	var totalCurrentAuctions int = 0

	c.OnHTML("td[class=PageNavigation]", func(e *colly.HTMLElement) {
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

	c.Visit("https://www.tibia.com/charactertrade/?subtopic=currentcharactertrades")

	if len(errors) > 0 {
		var b strings.Builder

		for _, err := range errors {
			b.WriteString(fmt.Sprintln(err.Error()))
		}

		return totalCurrentAuctions, eris.Errorf("Error getting total current auctions: %s", b.String())
	}

	return totalCurrentAuctions, nil
}

func (s *Service) extractAutctionId(link string) (int, error) {
	parsedLink, err := url.Parse(link)

	if err != nil {
		return 0, eris.New(fmt.Sprintf("Error converting link %s to url: %s", parsedLink, err.Error()))
	}

	auctionIdStr := parsedLink.Query().Get("auctionid")

	auctionId, err := strconv.Atoi(auctionIdStr)

	if err != nil {
		return 0, eris.New(fmt.Sprintf("Error converting auction ID '%s' to int: %s", auctionIdStr, err.Error()))
	}

	return auctionId, nil
}

func (s *Service) getCurrentAuctionLinks() (BazaarAuctionLinkSet, error) {
	set := make(BazaarAuctionLinkSet)

	c := colly.NewCollector(
		colly.AllowedDomains(constants.TibiaOfficialWebsite),
		colly.Debugger(&TibiaCharCollyLogDebugger{Prefix: "[CollectAuctionLinks] "}),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  constants.TibiaOfficialWebsite,
		RandomDelay: 5 * time.Second,
	})

	/*worlds, err := vendor.NewTibiaApi().GetWorlds()

	if err != nil {
		return set, err
	}*/

	worlds := []string{"Calmera"}

	currentAuctions, err := s.getTotalCurrentAuctions(c)

	if err != nil {
		return set, err
	}

	for _, world := range worlds {
		for currentPage := 1; ; currentPage++ {
			links, err := s.scrapAuctionListPage(c, world, currentPage)

			if err != nil {
				return set, err
			}

			if len(links) == 0 {
				break
			}

			newLinksAdded := 0

			for _, link := range links {
				auctionId, err := s.extractAutctionId(link)

				if err != nil {
					return set, err
				}

				if set.Has(auctionId) {
					continue
				}

				set.Set(auctionId, link)
				newLinksAdded++
			}

			if newLinksAdded == 0 {
				break
			}
		}
	}

	log.Default().Printf("Current auctions %d - Scrapped Auctions %d", currentAuctions, len(set))

	return set, nil
}

func (s *Service) getCharAuctionDetails(auctionId int, link string) error {
	var errors []error

	c := colly.NewCollector(
		colly.AllowedDomains(constants.TibiaOfficialWebsite),
		colly.Debugger(&TibiaCharCollyLogDebugger{Prefix: "[CollectAuctionDetails] "}),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  constants.TibiaOfficialWebsite,
		RandomDelay: 5 * time.Second,
	})

	set := NewBazaarAuctionDetailSet()

	c.OnHTML("div[class=Auction]", func(e *colly.HTMLElement) {
		var header AuctionHeader

		e.ForEachWithBreak("div[class]", func(_ int, ch *colly.HTMLElement) bool {
			class := ch.Attr("class")

			switch class {
			case "AuctionHeader":
				header.Name = e.ChildText("div[class=AuctionCharacterName]")

				header.World = e.ChildText("a[href]")

				auctionHeader := e.Text

				level, err := s.extractLevel(auctionHeader)

				header.Level = level

				if err != nil {
					errors = append(errors, eris.Errorf("Error extracting character level: %s", err.Error()))

					return false
				}

				header.Vocation = s.extractVocation(auctionHeader)

				header.Gender = s.extractGender(auctionHeader)
			case "AuctionBody":
				e.ForEachWithBreak("div", func(_ int, ch *colly.HTMLElement) bool {
					classes := strings.Split(ch.Attr("class"), " ")

					section := classes[len(classes)-1]

					switch section {
					case "AuctionOutfit":
						header.Img = ch.ChildAttr("img[class=AuctionOutfitImage]", "src")

					case "AuctionItemsViewBox":
						ch.ForEach("div[title]", func(_ int, itemViewBoxCh *colly.HTMLElement) {
							imgTitle := itemViewBoxCh.Attr("title")
							imgLink := itemViewBoxCh.ChildAttr("img", "src")

							header.SpecialItems = append(header.SpecialItems, ImgDisplay{Name: imgTitle, Link: imgLink})
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

								dateTimeUTC := dateUTC.Format(time.DateTime)

								if len(header.AuctionStart) == 0 {
									header.AuctionStart = dateTimeUTC

									break
								}

								header.AuctionEnd = dateTimeUTC

							case "ShortAuctionDataBidRow":
								selector := sAuctionDataCh.DOM.Children()

								rawBid := selector.Find("b").Text()

								bid, err := strconv.Atoi(rawBid)

								if err != nil {
									errors = append(errors, eris.Errorf("Error converting bid to int: %s", err.Error()))

									return false
								}

								header.Bid = bid
							}

							return true
						})

					case "SpecialCharacterFeatures":
						ch.ForEach("div", func(_ int, spcfCh *colly.HTMLElement) {
							header.SpecialFeatures = append(header.SpecialFeatures, spcfCh.Text)
						})
					}

					return true
				})
			}

			return true
		})

		charDetails, ok := set.Get(auctionId)

		if !ok {
			charDetails := BazaarCharAuctionDetail{
				AuctionHeader: header,
			}

			set.Set(auctionId, charDetails)

			return
		}

		charDetails.AuctionHeader = header

		set.Set(auctionId, charDetails)
	})

	c.Visit(link)

	return nil
}

func (s *Service) extractLevel(auctionHeader string) (int, error) {
	headerParts := strings.Split(auctionHeader, "|")

	levelStringHeader := headerParts[0]

	levelStringParts := strings.Split(levelStringHeader, ":")

	return strconv.Atoi(strings.TrimSpace(levelStringParts[1]))
}

func (s *Service) extractVocation(auctionHeader string) string {
	headerParts := strings.Split(auctionHeader, "|")

	levelStringHeader := headerParts[1]

	levelStringParts := strings.Split(levelStringHeader, ":")

	return strings.TrimSpace(levelStringParts[1])
}

func (s *Service) extractGender(auctionHeader string) string {
	headerParts := strings.Split(auctionHeader, "|")

	return strings.TrimSpace(headerParts[2])
}
