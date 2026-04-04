package auction

import (
	"math"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/adriein/tibia-char/pkg/enums"
	"github.com/rotisserie/eris"
)

type Mapper struct {
	worldRepository WorldRepository
}

func NewMapper(wr WorldRepository) *Mapper {
	return &Mapper{
		worldRepository: wr,
	}
}

func (m *Mapper) FromDTO(dto *AuctionDTO) (*Auction, error) {
	vocation, err := NewVocationFromName(dto.CharVocation)

	if err != nil {
		return nil, err
	}

	gender, err := NewGenderFromName(dto.CharGender)

	if err != nil {
		return nil, err
	}

	world, err := m.worldRepository.GetOrCreate(&World{Name: dto.CharWorld})

	if err != nil {
		return nil, err
	}

	var featuredItems []*FeaturedItem

	for _, item := range dto.FeaturedItems {
		if item.Link == "" {
			continue
		}

		fileName := path.Base(item.Link)
		itemIDString := strings.TrimSuffix(fileName, ".gif")

		itemID, err := strconv.Atoi(itemIDString)

		if err != nil {
			return nil, eris.Wrap(err, "Error parsing itemIDString to int")
		}

		featuredItems = append(featuredItems, &FeaturedItem{AuctionID: dto.AuctionId, ItemID: itemID})
	}

	var imbuements []*Imbuement

	for _, imbuementDTO := range dto.Imbuements {
		imbuement, err := NewImbuementFromName(imbuementDTO.Name)

		if err != nil {
			return nil, eris.Wrap(err, "Error converting ImbuementDTO to Imbuement")
		}

		imbuements = append(imbuements, imbuement)
	}

	var charms []*Charm

	for _, charmDTO := range dto.Charms {
		charm, err := NewCharmFromName(charmDTO.Name)

		if err != nil {
			return nil, eris.Wrap(err, "Error converting CharmDTO to Charm")
		}

		charm.Grade = charmDTO.Grade

		charms = append(charms, charm)
	}

	var quests []*Quest

	for _, questDTO := range dto.Quests {
		quest, err := NewQuestFromName(questDTO.Name)

		if err != nil {
			//TODO: right now we ignore the quest not found because we are not interested in all of them
			continue
		}

		quests = append(quests, quest)
	}

	var outfits []*Outfit

	for _, outfitDTO := range dto.Outfits {
		outfits = append(outfits, &Outfit{Name: outfitDTO.Name, Addons: outfitDTO.Addons})
	}

	var mounts []*Mount

	for _, mountDTO := range dto.Mounts {
		mounts = append(mounts, &Mount{Name: mountDTO.Name})
	}

	eurBidEquivalence := int(math.Round(float64(dto.Bid) * constants.TibiaCoinEuroEquivalence))

	return &Auction{
		AuctionID:        dto.AuctionId,
		TibiaAuctionLink: dto.Link,
		Img:              dto.ImgUrl,
		FeaturedItems:    featuredItems,
		Featured:         dto.Featured,
		CharName:         dto.CharName,
		CharLevel:        dto.CharLevel,
		CharVocation:     vocation,
		CharGender:       gender,
		CharWorld:        world,
		WorldTransfer:    dto.WorldTransfer,
		BossPoints:       dto.BossPoints,
		CharmExpansion:   dto.CharmPoints.Expansion,
		CharmPoints:      dto.CharmPoints.Points,
		Bid:              dto.Bid,
		BidFiat:          eurBidEquivalence,
		BidCurrency:      enums.CurrencyEUR,
		Stage:            dto.Stage,
		AuctionStart:     dto.AuctionStartTime,
		AuctionEnd:       dto.AuctionEndTime,
		Status:           dto.Status,
		Skills: &Skills{
			AuctionID:  dto.AuctionId,
			Axe:        dto.Skills.Axe,
			Club:       dto.Skills.Club,
			Distance:   dto.Skills.Distance,
			Fishing:    dto.Skills.Fishing,
			Fist:       dto.Skills.Fist,
			MagicLevel: dto.Skills.MagicLevel,
			Shielding:  dto.Skills.Shielding,
			Sword:      dto.Skills.Sword,
		},
		Charms:     charms,
		Quests:     quests,
		Outfits:    outfits,
		Imbuements: imbuements,
		DateAdd:    time.Now(),
		DateUpd:    time.Now(),
	}, nil
}
