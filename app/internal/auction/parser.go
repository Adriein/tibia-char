package auction

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/adriein/tibia-char/pkg/enums"
	"github.com/gocolly/colly/v2"
	"github.com/rotisserie/eris"
)

const (
	BaseAuctionListURL        = "https://www.tibia.com/charactertrade/?subtopic=currentcharactertrades"
	AuctionIDParam            = "auctionid"
	TibiaDotComForbiddenError = "Forbidden"
)

var RateLimitError = eris.New("Error rate limit reached")

/*
================================================================================
TOTAL AUCTIONS NUMBER PARSER
================================================================================
*/

type AuctionNumberHtmlParser struct {
	collector *colly.Collector
}

func NewAuctionNumberHtmlParser(c *colly.Collector) *AuctionNumberHtmlParser {
	return &AuctionNumberHtmlParser{collector: c}
}

func (p *AuctionNumberHtmlParser) Scrap() (int, error) {
	var errors []error
	var totalCurrentAuctions int = 0

	c := p.collector

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

	c.Visit(BaseAuctionListURL)

	if len(errors) > 0 {
		var b strings.Builder

		for _, err := range errors {
			b.WriteString(fmt.Sprintln(err.Error()))
		}

		return totalCurrentAuctions, eris.Errorf("Error getting total current auctions: %s", b.String())
	}

	return totalCurrentAuctions, nil
}

/*
================================================================================
LINK PARSER
================================================================================
*/

type AuctionListHtmlParser struct {
	collector *colly.Collector
}

func NewAuctionListHtmlParser(c *colly.Collector) *AuctionListHtmlParser {
	return &AuctionListHtmlParser{collector: c}
}

func (p *AuctionListHtmlParser) Scrap(world string, set *AuctionLinkSet) error {
	for page := 1; ; page++ {
		randDelay := time.Duration(2+rand.Intn(5)) * time.Second

		time.Sleep(randDelay)

		links, err := p.scrapAuctionListPage(world, page)

		if err != nil {
			return eris.Wrapf(err, "Failed to scrape page %d for world %s", page, world)
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

func (p *AuctionListHtmlParser) scrapAuctionListPage(world string, page int) ([]string, error) {
	var result []string
	var scrapeErr error

	c := p.collector

	c.OnHTML("div[class=AuctionLinks]", func(e *colly.HTMLElement) {
		e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
			if href := el.Attr("href"); href != "" {
				result = append(result, href)
			}
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		scrapeErr = eris.Wrapf(err, "scraping error for world %s page %d: status %d", world, page, r.StatusCode)
	})

	targetURL := fmt.Sprintf("%s&filter_world=%s&currentpage=%d", BaseAuctionListURL, world, page)

	if err := c.Visit(targetURL); err != nil {
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

/*
================================================================================
DETAIL PARSER
================================================================================
*/

type AuctionHtmlParser struct {
	collector *colly.Collector
}

func NewAuctionHtmlParser(c *colly.Collector) *AuctionHtmlParser {
	return &AuctionHtmlParser{
		collector: c,
	}
}

func (p *AuctionHtmlParser) Parse(ctx context.Context, auctionId int, link string) (*AuctionDTO, error) {
	dto := AuctionDTO{
		AuctionId:   auctionId,
		Link:        fmt.Sprintf("https://www.tibia.com/charactertrade/?subtopic=currentcharactertrades&page=details&auctionid=%d", auctionId),
		Skills:      &SkillsDTO{},
		CharmPoints: &CharmPointsDTO{},
	}

	var parseErr error

	c := p.collector

	proxyAddr := ctx.Value(constants.ProxyAddr).(string)

	if proxyAddr != constants.LocalProxy {
		c.SetProxy(proxyAddr)
	}

	c.OnError(func(r *colly.Response, err error) {
		if r.StatusCode == http.StatusForbidden {
			parseErr = eris.Wrapf(RateLimitError, "Failed to fetch %s: status %d", r.Request.URL, r.StatusCode)

			return
		}

		parseErr = eris.Wrapf(err, "Failed to fetch %s: status %d", r.Request.URL, r.StatusCode)
	})

	c.OnHTML("div[class=Auction]", func(e *colly.HTMLElement) {
		e.ForEachWithBreak("div[class]", func(_ int, ch *colly.HTMLElement) bool {
			switch ch.Attr("class") {
			case "AuctionHeader":
				if err := p.parseAuctionHeader(e, &dto); err != nil {
					parseErr = eris.Wrap(err, "Error parsing auction header")

					return false
				}

			case "AuctionBody":
				if err := p.parseAuctionBody(e, &dto); err != nil {
					parseErr = eris.Wrap(err, "Error parsing auction body")

					return false
				}
			}

			return true
		})
	})

	c.OnHTML("table[id=CharacterDetails]", func(e *colly.HTMLElement) {
		if err := p.parseAuctionGeneral(e, &dto); err != nil {
			parseErr = eris.Wrap(err, "Error parsing auction general")
		}
	})

	c.OnHTML("div[id=Imbuements]", func(e *colly.HTMLElement) {
		if err := p.parseAuctionImbuements(e, &dto); err != nil {
			parseErr = eris.Wrap(err, "Error parsing imbuements")
		}
	})

	c.OnHTML("div[id=Charms]", func(e *colly.HTMLElement) {
		if err := p.parseAuctionCharms(e, &dto); err != nil {
			parseErr = eris.Wrap(err, "Error parsing charms")
		}
	})

	c.OnHTML("div[id=CompletedQuestLines]", func(e *colly.HTMLElement) {
		if err := p.parseAuctionQuests(e, &dto); err != nil {
			parseErr = eris.Wrap(err, "Error parsing quests")
		}
	})

	if err := c.Visit(link); err != nil {
		if err.Error() == TibiaDotComForbiddenError {
			return nil, eris.Wrapf(RateLimitError, "Error visiting link")
		}

		return nil, eris.Wrapf(err, "Error visiting link")
	}

	if parseErr != nil {
		return nil, parseErr
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
	var bodyErr error

	e.ForEachWithBreak("div", func(_ int, ch *colly.HTMLElement) bool {
		classes := strings.Split(ch.Attr("class"), " ")

		section := classes[len(classes)-1]

		switch section {
		case "AuctionOutfit":
			dto.ImgUrl = ch.ChildAttr("img[class=AuctionOutfitImage]", "src")

		case "AuctionItemsViewBox":
			ch.ForEach("div[title]", func(_ int, itemViewBoxCh *colly.HTMLElement) {
				//TODO parse also the item tiers that has an extra element
				imgTitle := itemViewBoxCh.Attr("title")
				imgLink := itemViewBoxCh.ChildAttr("img", "src")

				dto.FeaturedItems = append(dto.FeaturedItems, &ImgDisplay{Name: imgTitle, Link: imgLink})
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
						bodyErr = eris.Wrap(err, "Error parsing auction date")

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

					stage := selector.Filter("div[class=ShortAuctionDataLabel]").Text()

					switch stage {
					case "Minimum Bid:":
						dto.Stage = enums.StageInitial
					case "Current Bid:":
						dto.Stage = enums.StageCurrent
					case "Winning Bid:":
						dto.Stage = enums.StageWinning
					}

					rawBid := strings.ReplaceAll(selector.Find("b").Text(), ",", "")

					bid, err := strconv.Atoi(rawBid)

					if err != nil {
						bodyErr = eris.Wrap(err, "Error converting bid to int")

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

	if bodyErr != nil {
		return eris.Wrap(bodyErr, "Error parsing auction body")
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

func (p *AuctionHtmlParser) parseAuctionGeneral(e *colly.HTMLElement, dto *AuctionDTO) error {
	var auctionGeneralErr error

	e.ForEachWithBreak("span[class=LabelV],td[class=PercentageColumn]", func(_ int, el *colly.HTMLElement) bool {
		if err := p.extractSkills(el, dto); err != nil {
			auctionGeneralErr = err

			return false
		}

		p.extractWorldTransfer(el, dto)

		if err := p.extractCharmPoints(el, dto); err != nil {
			auctionGeneralErr = err

			return false
		}

		if err := p.extractBossPoints(el, dto); err != nil {
			auctionGeneralErr = err

			return false
		}

		return true
	})

	return auctionGeneralErr
}

func (p *AuctionHtmlParser) extractWorldTransfer(e *colly.HTMLElement, dto *AuctionDTO) {
	if e.Text == "Regular World Transfer:" {
		divSibling := e.DOM.Siblings().Contents()

		worldTransferAllowance := divSibling.Text()

		switch worldTransferAllowance {
		case "can be purchased and used immediately":
			dto.WorldTransfer = true
		default:
			dto.WorldTransfer = false
		}
	}
}

func (p *AuctionHtmlParser) extractSkills(e *colly.HTMLElement, dto *AuctionDTO) error {
	skillComponents := e.DOM.Siblings().AndSelf().Nodes

	if len(skillComponents) == 3 {
		skillTagColumn := skillComponents[0]
		skillValueColumn := skillComponents[1]

		class := skillTagColumn.Attr[0]

		if class.Val == "LabelColumn" {
			skillName := skillTagColumn.FirstChild.FirstChild.Data

			skill, err := strconv.Atoi(skillValueColumn.FirstChild.Data)

			if err != nil {
				return eris.Wrapf(err, "Error converting skill to number")
			}

			switch skillName {
			case "Axe Fighting":
				dto.Skills.Axe = skill
			case "Club Fighting":
				dto.Skills.Club = skill
			case "Distance Fighting":
				dto.Skills.Distance = skill
			case "Fishing":
				dto.Skills.Fishing = skill
			case "Fist Fighting":
				dto.Skills.Fist = skill
			case "Magic Level":
				dto.Skills.MagicLevel = skill
			case "Shielding":
				dto.Skills.Shielding = skill
			case "Sword Fighting":
				dto.Skills.Sword = skill
			}
		}
	}

	return nil
}

func (p *AuctionHtmlParser) extractCharmPoints(e *colly.HTMLElement, dto *AuctionDTO) error {
	text := e.Text

	switch text {
	case "Charm Expansion:":
		charmExpansionText := e.DOM.Siblings().Children().Get(0).Attr[0].Val

		dto.CharmPoints.Expansion = strings.Contains(charmExpansionText, "icon_yes.png")

		return nil
	case "Available Charm Points:",
		"Spent Charm Points:",
		"Available Minor Charm Echoes:",
		"Spent Minor Charm Echoes:":
		pointString := e.DOM.Siblings().Eq(0).Text()

		points, err := strconv.Atoi(strings.ReplaceAll(pointString, ",", ""))

		if err != nil {
			return err
		}

		dto.CharmPoints.Points += points

		return nil
	}

	return nil
}

func (p *AuctionHtmlParser) extractBossPoints(e *colly.HTMLElement, dto *AuctionDTO) error {
	text := e.Text

	if text == "Boss Points:" {
		bossPointsString := e.DOM.Siblings().Eq(0).Text()

		sanitizedBossPointsString := strings.ReplaceAll(bossPointsString, ",", "")

		bossPoints, err := strconv.Atoi(sanitizedBossPointsString)

		if err != nil {
			return eris.Wrap(err, "Error converting boss points to int")
		}

		dto.BossPoints = bossPoints
	}

	return nil
}

func (p *AuctionHtmlParser) parseAuctionImbuements(e *colly.HTMLElement, dto *AuctionDTO) error {
	e.DOM.Find("tr[class=LabelH]").Siblings().EachWithBreak(func(i int, s *goquery.Selection) bool {
		if s.HasClass("IndicateMoreEntries") {
			return false
		}

		dto.Imbuements = append(dto.Imbuements, &ImbuementDTO{Name: s.Children().Eq(0).Text()})

		return true
	})

	return nil
}

func (p *AuctionHtmlParser) parseAuctionCharms(e *colly.HTMLElement, dto *AuctionDTO) error {
	var charmErr error
	relevantHTML := e.DOM.Find("tr[class=LabelH]")

	sanitizedTextContent := strings.ReplaceAll(relevantHTML.Children().Text(), " ", "")

	if strings.Contains(sanitizedTextContent, "CostsTypeCharmNameGrade") {
		relevantHTML.Siblings().EachWithBreak(func(i int, s *goquery.Selection) bool {
			if s.HasClass("IndicateMoreEntries") {
				return true
			}

			children := s.Children()

			charmType := children.Eq(1).Text()
			charmName := children.Eq(2).Text()
			charmGradeString := children.Eq(3).Text()

			charmGrade, err := strconv.Atoi(charmGradeString)

			if err != nil {
				charmErr = eris.Wrap(err, "Error parsing charm grade to int")

				return false
			}

			dto.Charms = append(dto.Charms, &CharmDTO{Type: charmType, Name: charmName, Grade: charmGrade})

			return true
		})
	}

	if charmErr != nil {
		return charmErr
	}

	return nil
}

func (p *AuctionHtmlParser) parseAuctionQuests(e *colly.HTMLElement, dto *AuctionDTO) error {
	relevantHTML := e.DOM.Find("tr[class=LabelH]")

	if strings.Contains(relevantHTML.Children().Text(), "Quest Line Name") {
		relevantHTML.Siblings().EachWithBreak(func(i int, s *goquery.Selection) bool {
			quest := s.Children().Text()

			dto.Quests = append(dto.Quests, &QuestDTO{Name: quest})

			return true
		})
	}

	return nil
}

type ParserFactory interface {
	CreateAuctionNumberParser(collector *CollyScrapper) *AuctionNumberHtmlParser
	CreateAuctionListParser(collector *CollyScrapper) *AuctionListHtmlParser
	CreateAuctionParser(collector *CollyScrapper) *AuctionHtmlParser
}

type HtmlParserFactory struct{}

func (f *HtmlParserFactory) CreateAuctionNumberParser(collector *CollyScrapper) *AuctionNumberHtmlParser {
	return NewAuctionNumberHtmlParser(collector.Collector)
}

func (f *HtmlParserFactory) CreateAuctionListParser(collector *CollyScrapper) *AuctionListHtmlParser {
	return NewAuctionListHtmlParser(collector.Collector)
}

func (f *HtmlParserFactory) CreateAuctionParser(collector *CollyScrapper) *AuctionHtmlParser {
	return NewAuctionHtmlParser(collector.Collector)
}
