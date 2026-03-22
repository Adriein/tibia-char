package auction

import (
	"time"

	"github.com/adriein/tibia-char/pkg/enums"
)

type SkillsDTO struct {
	Axe        int
	Club       int
	Distance   int
	Fishing    int
	Fist       int
	MagicLevel int
	Shielding  int
	Sword      int
}

type CharmPointsDTO struct {
	Expansion bool
	Points    int
}

type CharmDTO struct {
	Type  string
	Name  string
	Grade int
}

type ImbuementDTO struct {
	Name string
}

type QuestDTO struct {
	Name string
}

type ImgDisplayDTO struct {
	Link string
	Name string
}

type AuctionDTO struct {
	AuctionId        int
	Link             string
	ImgUrl           string
	FeaturedItems    []*ImgDisplayDTO
	Featured         []string
	CharName         string
	CharLevel        int
	CharVocation     string
	CharGender       string
	CharWorld        string
	Skills           *SkillsDTO
	CharmPoints      *CharmPointsDTO
	Charms           []*CharmDTO
	Imbuements       []*ImbuementDTO
	Quests           []*QuestDTO
	WorldTransfer    bool
	BossPoints       int
	Bid              int
	Stage            enums.AuctionStage
	Status           enums.AuctionRecordableStatus
	AuctionStartTime time.Time
	AuctionEndTime   time.Time
}
