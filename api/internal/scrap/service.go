package scrap

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/rotisserie/eris"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) ScrapBazaar(world string) error {
	set := make(map[int]string)

	currentPage := 0

	for {
		links, err := s.scrapPage(currentPage)

		if err != nil {
			return err
		}

		for _, link := range links {
			parsedLink, err := url.Parse(link)

			if err != nil {
				return eris.New(fmt.Sprintf("Error converting link to url: %s", err.Error()))
			}

			auctionIdStr := parsedLink.Query().Get("auctionid")

			auctionId, err := strconv.Atoi(auctionIdStr)

			if err != nil {
				return eris.New(fmt.Sprintf("Error converting to integer: %s", err.Error()))
			}

			set.Get()
		}

	}

	return nil
}

func (s *Service) scrapPage(page int) ([]string, error) {
	var result []string

	c := colly.NewCollector(
		colly.AllowedDomains("www.tibia.com"),
	)

	c.OnHTML("div[class=AuctionLinks]", func(e *colly.HTMLElement) {
		e.ForEach("a[href]", func(_ int, e *colly.HTMLElement) {
			charDetailLink := e.Attr("href")

			result = append(result, charDetailLink)
		})
	})

	c.Visit(fmt.Sprintf("https://www.tibia.com/charactertrade/?subtopic=currentcharactertrades&currentpage=%d", page))

	return result, nil
}

func (s *Service) getTotalCurrentAuctions() (int, error) {
	var errors []error
	var totalCurrentAuctions int = 0

	c := colly.NewCollector(
		colly.AllowedDomains("www.tibia.com"),
	)

	c.OnHTML("td[class=PageNavigation]", func(e *colly.HTMLElement) {
		htmlExtractedText := e.Text

		parts := strings.Split(htmlExtractedText, ": ")

		if len(parts) < 2 {
			err := eris.New(fmt.Sprintf("String format is unexpected: %s", htmlExtractedText))
			errors = append(errors, err)

			return
		}

		numberStr := parts[1]

		cleanStr := strings.ReplaceAll(numberStr, ",", "")

		resultInt, err := strconv.Atoi(cleanStr)

		if err != nil {
			err := eris.New(fmt.Sprintf("Error converting to integer: %s", err.Error()))
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

		return totalCurrentAuctions, eris.New(fmt.Sprintf("Error getting total current auctions: %s", b.String()))
	}

	return totalCurrentAuctions, nil
}
