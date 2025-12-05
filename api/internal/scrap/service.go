package scrap

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/adriein/tibia-char/pkg/vendor"
	"github.com/gocolly/colly/v2"
	"github.com/rotisserie/eris"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) ScrapBazaar() error {
	set := make(BazaarAuctionLinkSet)

	worlds, err := vendor.NewTibiaApi().GetWorlds()

	if err != nil {
		return err
	}

	for _, world := range worlds {
		for currentPage := 1; ; currentPage++ {
			links, err := s.scrapPage(world, currentPage)

			if err != nil {
				return err
			}

			newLinksAdded := 0

			for _, link := range links {
				parsedLink, err := url.Parse(link)

				if err != nil {
					return eris.New(fmt.Sprintf("Error converting link %s to url: %s", parsedLink, err.Error()))
				}

				auctionIdStr := parsedLink.Query().Get("auctionid")

				auctionId, err := strconv.Atoi(auctionIdStr)

				if err != nil {
					return eris.New(fmt.Sprintf("Error converting auction ID '%s' to int: %s", auctionIdStr, err.Error()))
				}

				if set.Has(auctionId) {
					if newLinksAdded == 0 {
						return nil
					}

					continue
				}

				set.Set(auctionId, link)
				newLinksAdded++
			}
		}
	}

	return nil
}

func (s *Service) scrapPage(world string, page int) ([]string, error) {
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

	c.Visit(fmt.Sprintf("https://www.tibia.com/charactertrade/?subtopic=currentcharactertrades&filter_world=%s&currentpage=%d", world, page))

	return result, nil
}
