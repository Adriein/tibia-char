package scrap

import (
	"fmt"
	"log"

	"github.com/gocolly/colly/v2"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) ScrapBazaar(world string) error {
	log.Print("collecting...")
	c := colly.NewCollector(
	//colly.AllowedDomains("tibia.com"),
	)

	c.OnHTML("div[class=Auction]", func(e *colly.HTMLElement) {
		e.ForEach("div", func(_ int, e *colly.HTMLElement) {
			fmt.Printf("%+v", e)
		})
	})

	c.Visit("https://www.tibia.com/charactertrade/?subtopic=currentcharactertrades")

	return nil
}
