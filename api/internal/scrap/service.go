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
	"github.com/gocolly/colly/v2/debug"
	"github.com/rotisserie/eris"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) ScrapBazaar() error {
	auctionLinkSet, err := s.getCurrentAuctionLinks()

	for auctionId, link := range auctionLinkSet {
		s.getCharAuctionDetails(auctionId, link)
	}

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) scrapAuctionListPage(world string, page int) ([]string, error) {
	var result []string

	c := colly.NewCollector(
		colly.AllowedDomains(constants.TibiaOfficialWebsite),
		colly.Debugger(&debug.LogDebugger{}),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  constants.TibiaOfficialWebsite,
		RandomDelay: 5 * time.Second,
	})

	c.OnHTML("div[class=AuctionLinks]", func(e *colly.HTMLElement) {
		e.ForEach("a[href]", func(_ int, e *colly.HTMLElement) {
			charDetailLink := e.Attr("href")

			result = append(result, charDetailLink)
		})
	})

	c.Visit(fmt.Sprintf("https://www.tibia.com/charactertrade/?subtopic=currentcharactertrades&filter_world=%s&currentpage=%d", world, page))

	return result, nil
}

func (s *Service) getTotalCurrentAuctions() (int, error) {
	var errors []error
	var totalCurrentAuctions int = 0

	c := colly.NewCollector(
		colly.AllowedDomains(constants.TibiaOfficialWebsite),
	)

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

	/*worlds, err := vendor.NewTibiaApi().GetWorlds()

	if err != nil {
		return set, err
	}*/

	worlds := []string{"Secura"}

	currentAuctions, err := s.getTotalCurrentAuctions()

	if err != nil {
		return set, err
	}

	for _, world := range worlds {
		for currentPage := 1; ; currentPage++ {
			links, err := s.scrapAuctionListPage(world, currentPage)

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
		colly.Debugger(&debug.LogDebugger{}),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  constants.TibiaOfficialWebsite,
		Parallelism: 5,
		RandomDelay: 5 * time.Second,
	})

	set := NewBazaarAuctionDetailSet()

	c.OnHTML("div[class=AuctionHeader]", func(e *colly.HTMLElement) {
		charName := e.ChildText("div[class=AuctionCharacterName]")

		world := e.ChildText("a[href]")

		auctionHeader := e.Text

		level, err := s.extractLevel(auctionHeader)

		if err != nil {
			errors = append(errors, eris.Errorf("Error extracting character level: %s", err.Error()))

			return
		}

		vocation := s.extractVocation(auctionHeader)

		gender := s.extractGender(auctionHeader)

		charDetails, ok := set.Get(auctionId)

		if !ok {
			charDetails := BazaarCharAuctionDetail{
				AuctionHeader: AuctionHeader{
					Name:     charName,
					World:    world,
					Level:    level,
					Vocation: vocation,
					Gender:   gender,
				},
			}

			set.Set(auctionId, charDetails)

			return
		}

		charDetails.AuctionHeader.Name = charName
		charDetails.AuctionHeader.World = world
		charDetails.AuctionHeader.Level = level
		charDetails.AuctionHeader.Vocation = vocation
		charDetails.AuctionHeader.Gender = gender

		set.Set(auctionId, charDetails)
	})

	c.OnHTML("div[class=AuctionBody]", func(e *colly.HTMLElement) {
		var displayImg string
		var specialItems []ImgDisplay

		e.ForEach("div", func(_ int, ch *colly.HTMLElement) {
			classes := strings.Split(ch.Attr("class"), " ")

			section := classes[len(classes)-1]

			switch section {
			case "AuctionOutfit":
				displayImg = ch.ChildAttr("img[class=AuctionOutfitImage]", "src")
			case "AuctionItemsViewBox":
				ch.ForEach("div[title]", func(_ int, ivbCh *colly.HTMLElement) {
					imgTitle := ivbCh.Attr("title")
					imgLink := ivbCh.ChildAttr("img", "src")

					specialItems = append(specialItems, ImgDisplay{Name: imgTitle, Link: imgLink})
				})
			}
		})

		charDetails, ok := set.Get(auctionId)

		if !ok {
			charDetails := BazaarCharAuctionDetail{
				AuctionHeader: AuctionHeader{
					Img: displayImg,
				},
			}

			set.Set(auctionId, charDetails)

			return
		}

		charDetails.AuctionHeader.Img = displayImg

		set.Set(auctionId, charDetails)
	})

	c.Visit(link)

	c.Wait()

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
