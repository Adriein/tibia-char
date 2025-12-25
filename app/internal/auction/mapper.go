package auction

import (
	"time"

	"github.com/adriein/tibia-char/internal/auction/model"
	"github.com/adriein/tibia-char/pkg/constants"
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

	return &model.Auction{
		AuctionID:        int64(dto.AuctionId),
		TibiaAuctionLink: dto.Link,
		Img:              dto.ImgUrl,
		FeaturedItems:    dto.FeaturedItems,
		Featured:         dto.Featured,
		CharName:         dto.CharName,
		CharLevel:        dto.CharLevel,
		CharVocation:     vocation,
		CharGender:       gender,
		CharWorld:        world,
		Bid:              dto.Bid,
		AuctionStart:     dto.AuctionStartTime,
		AuctionEnd:       dto.AuctionEndTime,
		Status:           constants.RecordableActive,
		DateAdd:          time.Now(),
		DateUpd:          time.Now(),
	}, nil
}
