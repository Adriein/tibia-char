package scrap

import (
	"fmt"
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
	totalActiveAuctions, err := s.getTotalCurrentAuctions()

	if err != nil {
		fmt.Println("error")
	}

	fmt.Println(totalActiveAuctions)

	c := colly.NewCollector(
		colly.AllowedDomains("www.tibia.com"),
	)

	c.OnHTML("div[class=AuctionLinks]", func(e *colly.HTMLElement) {
		e.ForEach("a[href]", func(_ int, e *colly.HTMLElement) {
			charDetailLink := e.Attr("href")

			fmt.Println(charDetailLink)
		})
	})

	c.Visit("https://www.tibia.com/charactertrade/?subtopic=currentcharactertrades&currentpage=0")

	return nil
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
