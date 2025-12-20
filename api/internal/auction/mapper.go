package auction

import (
	"time"
)

type Mapper struct {
	worldRepository WorldRepository
}

func NewMapper(wr WorldRepository) *Mapper {
	return &Mapper{
		worldRepository: wr,
	}
}

func (m *Mapper) ToDomain(dto *AuctionDTO) (*Auction, error) {
	vocation, err := NewVocationFromName(dto.CharVocation)

	if err != nil {
		return nil, err
	}

	gender, err := NewGenderFromName(dto.CharGender)

	if err != nil {
		return nil, err
	}

	world, err := m.worldRepository.GetOrCreate(dto.CharWorld)

	if err != nil {
		return nil, err
	}

	return &Auction{
		Id:               dto.AuctionId,
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
		IsActive:         true,
		DateAdd:          time.Now(),
		DateUpd:          time.Now(),
	}, nil
}
