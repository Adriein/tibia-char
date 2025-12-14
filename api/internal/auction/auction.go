package auction

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adriein/tibia-char/pkg/helper/array"
	"github.com/gocolly/colly/v2"
	"github.com/rotisserie/eris"
)

type Vocation struct {
	Id   string
	Name string
}

type Gender struct {
	Id   string
	Name string
}

type World struct {
	Id   string
	Name string
}

type Auctions struct {
	Data []Auction
}

func (a *Auctions) Scrap(ar AuctionRepository, vr VocationRepository, gr GenderRepository, wr WorldRepository, ls *CollyScrapper, ds *CollyScrapper, logger *log.Logger) error {
	currentAuctions, err := a.getTotalCurrentAuctions(ls)

	if err != nil {
		return err
	}

	auctionLinkSet, err := a.getCurrentAuctionLinks(ls)

	logger.Printf("Current auctions %d - Scrapped Auctions %d", currentAuctions, len(auctionLinkSet))

	if err != nil {
		return err
	}

	c := ds.Collector

	const MaxConcurrency = 5

	links := array.Chunk(auctionLinkSet.Values(), MaxConcurrency)

	var wg sync.WaitGroup

	maxWorkers := make(chan struct{}, MaxConcurrency)

	for i, chunk := range links {
		if i != 0 {
			randDelay := time.Duration(1+rand.Intn(5)) * time.Second

			time.Sleep(randDelay)
		}

		for _, link := range chunk {
			maxWorkers <- struct{}{}

			wg.Add(1)

			go func(url string) {
				defer wg.Done()

				defer func() { <-maxWorkers }()

				auctionId, err := a.extractAutctionId(url)

				if err != nil {
					logger.Printf("Error extracting auctionId from link: %s\n", err.Error())

					return
				}

				auction := Auction{}

				err = auction.ScrapAuction(vr, gr, wr, c, url)

				if err != nil {
					logger.Printf("Error fetching details for auction %d: %v\n", auctionId, err)

					return
				}

				ar.Save(&auction)

			}(link)
		}

		wg.Wait()
	}

	return nil
}

func (a *Auctions) getCurrentAuctionLinks(ls *CollyScrapper) (BazaarAuctionLinkSet, error) {
	set := make(BazaarAuctionLinkSet)

	c := ls.Collector

	/*worlds, err := vendor.NewTibiaApi().GetWorlds()

	if err != nil {
		return set, err
	}*/

	worlds := []string{"Calmera"}

	for _, world := range worlds {
		for currentPage := 1; ; currentPage++ {
			links, err := a.scrapAuctionListPage(c, world, currentPage)

			if err != nil {
				return set, err
			}

			if len(links) == 0 {
				break
			}

			newLinksAdded := 0

			for _, link := range links {
				auctionId, err := a.extractAutctionId(link)

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

	return set, nil
}

func (a *Auctions) scrapAuctionListPage(c *colly.Collector, world string, page int) ([]string, error) {
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

func (a *Auctions) getTotalCurrentAuctions(ls *CollyScrapper) (int, error) {
	var errors []error
	var totalCurrentAuctions int = 0

	c := ls.Collector

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

func (a *Auctions) extractAutctionId(link string) (int, error) {
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

type Auction struct {
	Id               int
	TibiaAuctionId   int
	TibiaAuctionLink string
	Img              string
	FeaturedItems    []ImgDisplay
	Featured         []string
	CharName         string
	CharLevel        int
	CharVocation     *Vocation
	CharGender       *Gender
	CharWorld        *World
	Bid              int
	AuctionStart     time.Time
	AuctionEnd       time.Time
	IsActive         bool
	DateAdd          time.Time
	DateUpd          time.Time
}

func (a *Auction) ScrapAuction(vr VocationRepository, gr GenderRepository, wr WorldRepository, c *colly.Collector, link string) error {
	var errors []error

	c.OnError(func(r *colly.Response, err error) {
		errors = append(errors, eris.Errorf("Url: %s failed with status code: %d", r.Request.URL, r.StatusCode))
	})

	c.OnHTML("div[class=Auction]", func(e *colly.HTMLElement) {
		e.ForEachWithBreak("div[class]", func(_ int, ch *colly.HTMLElement) bool {
			class := ch.Attr("class")

			switch class {
			case "AuctionHeader":
				a.CharName = e.ChildText("div[class=AuctionCharacterName]")

				world, err := wr.GetByName(e.ChildText("a[href]"))

				if err != nil {
					errors = append(errors, err)

					return false
				}

				a.CharWorld = world

				auctionHeader := e.Text

				level, err := a.extractLevel(auctionHeader)

				a.CharLevel = level

				if err != nil {
					errors = append(errors, eris.Errorf("Error extracting character level: %s", err.Error()))

					return false
				}

				vocation, err := vr.GetByName(a.extractVocation(auctionHeader))

				if err != nil {
					errors = append(errors, err)

					return false
				}

				a.CharVocation = vocation

				gender, err := gr.GetByName(a.extractGender(auctionHeader))

				if err != nil {
					errors = append(errors, err)

					return false
				}

				a.CharGender = gender
			case "AuctionBody":
				e.ForEachWithBreak("div", func(_ int, ch *colly.HTMLElement) bool {
					classes := strings.Split(ch.Attr("class"), " ")

					section := classes[len(classes)-1]

					switch section {
					case "AuctionOutfit":
						a.Img = ch.ChildAttr("img[class=AuctionOutfitImage]", "src")

					case "AuctionItemsViewBox":
						ch.ForEach("div[title]", func(_ int, itemViewBoxCh *colly.HTMLElement) {
							imgTitle := itemViewBoxCh.Attr("title")
							imgLink := itemViewBoxCh.ChildAttr("img", "src")

							a.FeaturedItems = append(a.FeaturedItems, ImgDisplay{Name: imgTitle, Link: imgLink})
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

								if a.AuctionStart.IsZero() {
									a.AuctionStart = dateUTC

									break
								}

								a.AuctionEnd = dateUTC

							case "ShortAuctionDataBidRow":
								selector := sAuctionDataCh.DOM.Children()

								rawBid := strings.ReplaceAll(selector.Find("b").Text(), ",", "")

								bid, err := strconv.Atoi(rawBid)

								if err != nil {
									errors = append(errors, eris.Errorf("Error converting bid to int: %s", err.Error()))

									return false
								}

								a.Bid = bid

								return false
							}

							return true
						})

					case "SpecialCharacterFeatures":
						ch.ForEach("div", func(_ int, spcfCh *colly.HTMLElement) {
							a.Featured = append(a.Featured, spcfCh.Text)
						})
					}

					return true
				})
			}

			return true
		})
	})

	c.Visit(link)

	if len(errors) != 0 {
		//TODO: define what to do with those errors array
		return eris.Errorf("%d Errors happened on Character Detail extraction", len(errors))
	}

	return nil
}

func (a *Auction) extractLevel(auctionHeader string) (int, error) {
	headerParts := strings.Split(auctionHeader, "|")

	levelStringHeader := headerParts[0]

	levelStringParts := strings.Split(levelStringHeader, ":")

	return strconv.Atoi(strings.TrimSpace(levelStringParts[1]))
}

func (a *Auction) extractVocation(auctionHeader string) string {
	headerParts := strings.Split(auctionHeader, "|")

	levelStringHeader := headerParts[1]

	levelStringParts := strings.Split(levelStringHeader, ":")

	return strings.TrimSpace(levelStringParts[1])
}

func (a *Auction) extractGender(auctionHeader string) string {
	headerParts := strings.Split(auctionHeader, "|")

	return strings.TrimSpace(headerParts[2])
}

type BazaarAuctionLinkSet map[int]string

func (set BazaarAuctionLinkSet) Get(key int) (string, bool) {
	value, ok := set[key]

	return value, ok
}

func (set BazaarAuctionLinkSet) Set(key int, value string) {
	set[key] = value
}

func (set BazaarAuctionLinkSet) Del(key int) {
	delete(set, key)
}

func (set BazaarAuctionLinkSet) Has(key int) bool {
	_, ok := set[key]

	return ok
}

func (set BazaarAuctionLinkSet) Values() []string {
	values := make([]string, 0, len(set))

	for _, v := range set {
		values = append(values, v)
	}

	return values
}

type ImgDisplay struct {
	Link string
	Name string
}

type AuctionHeader struct {
	Img             string
	Name            string
	Level           int
	Vocation        string
	Gender          string
	World           string
	SpecialItems    []ImgDisplay
	SpecialFeatures []string
	Bid             int
	AuctionStart    string
	AuctionEnd      string
}

type BazaarCharAuctionDetail struct {
	AuctionHeader AuctionHeader
	General       struct {
		Mounts               int
		Outfits              int
		CreationDate         time.Time
		Gold                 int
		RegularWorldTransfer string
		Skills               struct {
			AxeFighting      int
			ClubFighting     int
			DistanceFighting int
			Fishing          int
			FistFighting     int
			MagicLevel       int
			Shielding        int
			SwordFighting    int
		}
		Charms struct {
			CharmExpansion            string
			AvailableCharmPoints      int
			SpentCharmPoints          int
			AvailableMinorCharmEchoes int
			SpentMinorCharmEchoes     int
		}
		HuntingTasks struct {
			TaskPoints                   int
			PermanentWeeklyTaskExpansion string
			PermanentPreySlots           int
			PreyWildcards                int
		}
		Hirelings struct {
			Amount  int
			Jobs    int
			Outfits int
		}
		ExaltedDust             string
		AnimusMasteriesUnlocked int
		BossPoints              int
		BonusPromotionPoints    int
	}
	ItemSummary []struct {
		Img    string
		Amount int
		Name   string
	}
	StoreItemSummary []struct {
		Img    string
		Amount int
		Name   string
	}
	Mounts       []ImgDisplay
	StoreMounts  []ImgDisplay
	Outfits      []ImgDisplay
	StoreOutfits []ImgDisplay
	Imbuements   []string
	Charms       []struct {
		Cost  int
		Type  string
		Name  string
		Grade int
	}
	Quests   []string
	Bestiary []struct {
		Step    int
		Kills   int
		Name    string
		Mastery bool
	}
	Bosstiary []struct {
		Step  int
		Kills int
		Name  string
	}
	BountyTalisman struct {
		Points int
		Bounty []struct {
			Name  string
			Level int
			Value float64
		}
	}
	RevealedGems []struct {
		Gem  string
		Mod1 ImgDisplay
		Mod2 ImgDisplay
		Mod3 ImgDisplay
	}
	FragmentProgress []struct {
		Grade      string
		SupremeMod string
	}
	Proficiencies []struct {
		Weapon        string
		Level         string
		TotalProgress int
		Mastery       bool
	}
}
