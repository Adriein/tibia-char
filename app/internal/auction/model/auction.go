package model

import (
	"strings"
	"sync"
	"time"

	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/adriein/tibia-char/pkg/enums"
	"github.com/rotisserie/eris"
)

type AuctionDTO struct {
	AuctionId        int
	Link             string
	ImgUrl           string
	FeaturedItems    []*ImgDisplay
	Featured         []string
	CharName         string
	CharLevel        int
	CharVocation     string
	CharGender       string
	CharWorld        string
	Skills           *SkillsDTO
	WorldTransfer    bool
	Bid              int
	AuctionStartTime time.Time
	AuctionEndTime   time.Time
}

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

type Vocation struct {
	Id   int
	Name string
}

var vocationMap = map[string]*Vocation{
	constants.Knight:   {Id: constants.VocationKnight, Name: constants.Knight},
	constants.Paladin:  {Id: constants.VocationPaladin, Name: constants.Paladin},
	constants.Sorcerer: {Id: constants.VocationSorcerer, Name: constants.Sorcerer},
	constants.Druid:    {Id: constants.VocationDruid, Name: constants.Druid},
	constants.Monk:     {Id: constants.VocationMonk, Name: constants.Monk},
}

var genderMap = map[string]*Gender{
	constants.Male:   {Id: constants.GenderMale, Name: constants.Male},
	constants.Female: {Id: constants.GenderFemale, Name: constants.Female},
}

func NewVocationFromName(name string) (*Vocation, error) {
	promotionRemovedString := strings.NewReplacer(
		"Elite ", "",
		"Royal ", "",
		"Master ", "",
		"Elder ", "",
		"Exalted ", "",
	).Replace(name)

	lowerCaseVocation := strings.ToLower(promotionRemovedString)

	if vocation, ok := vocationMap[lowerCaseVocation]; ok {
		return vocation, nil
	}

	return nil, eris.Errorf("Vocation %s not registered", name)
}

type Gender struct {
	Id   int
	Name string
}

func NewGenderFromName(name string) (*Gender, error) {
	lowerCaseGender := strings.ToLower(name)

	if gender, ok := genderMap[lowerCaseGender]; ok {
		return gender, nil
	}

	return nil, eris.Errorf("Gender %s not registered", name)
}

type World struct {
	Id        int
	Name      string
	Location  string
	BattleEye enums.BattleEye
	Pvp       string
}

type Skills struct {
	ID         int64
	AuctionID  int64
	Axe        int
	Club       int
	Distance   int
	Fishing    int
	Fist       int
	MagicLevel int
	Shielding  int
	Sword      int
}

type FeaturedItem struct {
	ID        int64
	AuctionID int64
	ItemID    int
}

type Auction struct {
	ID               int64
	AuctionID        int64
	TibiaAuctionLink string
	Img              string
	FeaturedItems    []*FeaturedItem
	Featured         []string
	CharName         string
	CharLevel        int
	CharVocation     *Vocation
	CharGender       *Gender
	CharWorld        *World
	Skills           *Skills
	WorldTransfer    bool
	Bid              int
	AuctionStart     time.Time
	AuctionEnd       time.Time
	Status           enums.AuctionRecordableStatus
	DateAdd          time.Time
	DateUpd          time.Time
}

type AuctionLinkSet struct {
	sync.RWMutex
	Data map[int]string
}

func NewAuctionLinkSet() *AuctionLinkSet {
	return &AuctionLinkSet{
		Data: make(map[int]string),
	}
}

func (s *AuctionLinkSet) Get(key int) (string, bool) {
	s.RLock()
	defer s.RUnlock()

	value, ok := s.Data[key]

	return value, ok
}

func (s *AuctionLinkSet) Set(key int, value string) {
	s.Lock()
	defer s.Unlock()

	s.Data[key] = value
}

func (s *AuctionLinkSet) Has(key int) bool {
	s.RLock()
	defer s.RUnlock()

	_, ok := s.Data[key]

	return ok
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
