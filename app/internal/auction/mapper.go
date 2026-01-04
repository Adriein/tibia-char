package auction

import (
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/adriein/tibia-char/internal/auction/model"
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

func (m *Mapper) FromDTO(dto *model.AuctionDTO) (*model.Auction, error) {
	vocation, err := model.NewVocationFromName(dto.CharVocation)

	if err != nil {
		return nil, err
	}

	gender, err := model.NewGenderFromName(dto.CharGender)

	if err != nil {
		return nil, err
	}

	world, err := m.worldRepository.GetOrCreate(&model.World{Name: dto.CharWorld})

	if err != nil {
		return nil, err
	}

	var featuredItems []*model.FeaturedItem

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

		featuredItems = append(featuredItems, &model.FeaturedItem{AuctionID: dto.AuctionId, ItemID: itemID})
	}

	var imbuements []*model.Imbuement

	for _, imbuementDTO := range dto.Imbuements {
		imbuement, err := model.NewImbuementFromName(imbuementDTO.Name)

		if err != nil {
			return nil, eris.Wrap(err, "Error converting ImbuementDTO to Imbuement")
		}

		imbuements = append(imbuements, imbuement)
	}

	var charms []*model.Charm

	for _, charmDTO := range dto.Charms {
		charm, err := model.NewCharmFromName(charmDTO.Name)

		if err != nil {
			return nil, eris.Wrap(err, "Error converting CharmDTO to Charm")
		}

		charm.Grade = charmDTO.Grade

		charms = append(charms, charm)
	}

	var quests []*model.Quest

	for _, questDTO := range dto.Quests {
		quest, err := model.NewQuestFromName(questDTO.Name)

		if err != nil {
			//TODO: right now we ignore the quest not found because we are not interested in all of them
			continue
		}

		quests = append(quests, quest)
	}

	return &model.Auction{
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
		Stage:            dto.Stage,
		AuctionStart:     dto.AuctionStartTime,
		AuctionEnd:       dto.AuctionEndTime,
		Status:           enums.RecordableActive,
		Skills: &model.Skills{
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
		Imbuements: imbuements,
		DateAdd:    time.Now(),
		DateUpd:    time.Now(),
	}, nil
}
