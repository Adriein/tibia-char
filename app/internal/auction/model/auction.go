package model

import (
	"strings"
	"sync"
	"time"

	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/adriein/tibia-char/pkg/enums"
	"github.com/rotisserie/eris"
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
	CharmPoints      *CharmPointsDTO
	Charms           []*CharmDTO
	Imbuements       []*ImbuementDTO
	WorldTransfer    bool
	BossPoints       int
	Bid              int
	AuctionStartTime time.Time
	AuctionEndTime   time.Time
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
	constants.None:     {Id: constants.VocationNone, Name: constants.None},
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
	AuctionID  int
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
	AuctionID int
	ItemID    int
}

type Charm struct {
	ID    int
	Type  string
	Name  string
	Grade int
}

var charmMap = map[string]*Charm{
	"adrenaline burst": {ID: 1, Type: "Minor", Name: "Adrenaline Burst"},
	"bless":            {ID: 2, Type: "Minor", Name: "Bless"},
	"carnage":          {ID: 3, Type: "Major", Name: "Carnage"},
	"cleanse":          {ID: 4, Type: "Minor", Name: "Cleanse"},
	"cripple":          {ID: 5, Type: "Minor", Name: "Cripple"},
	"curse (charm)":    {ID: 6, Type: "Major", Name: "Curse (Charm)"},
	"divine wrath":     {ID: 7, Type: "Major", Name: "Divine Wrath"},
	"dodge":            {ID: 8, Type: "Major", Name: "Dodge"},
	"enflame":          {ID: 9, Type: "Major", Name: "Enflame"},
	"fatal hold":       {ID: 10, Type: "Minor", Name: "Fatal Hold"},
	"freeze":           {ID: 11, Type: "Major", Name: "Freeze"},
	"gut":              {ID: 12, Type: "Minor", Name: "Gut"},
	"low blow":         {ID: 13, Type: "Major", Name: "Low Blow"},
	"numb":             {ID: 14, Type: "Minor", Name: "Numb"},
	"overflux":         {ID: 15, Type: "Major", Name: "Overflux"},
	"overpower":        {ID: 16, Type: "Major", Name: "Overpower"},
	"parry":            {ID: 17, Type: "Major", Name: "Parry"},
	"poison":           {ID: 18, Type: "Major", Name: "Poison"},
	"savage blow":      {ID: 19, Type: "Major", Name: "Savage Blow"},
	"scavenge":         {ID: 20, Type: "Minor", Name: "Scavenge"},
	"vampiric embrace": {ID: 21, Type: "Minor", Name: "Vampiric Embrace"},
	"void inversion":   {ID: 22, Type: "Minor", Name: "Void Inversion"},
	"voids call":       {ID: 23, Type: "Minor", Name: "Voids Call"},
	"wound":            {ID: 24, Type: "Major", Name: "Wound"},
	"zap":              {ID: 25, Type: "Major", Name: "Zap"},
}

func NewCharmFromName(name string) (*Charm, error) {
	lowerCaseCharm := strings.ToLower(name)

	sanitizedCharm := strings.ReplaceAll(lowerCaseCharm, "'", "")

	if charm, ok := charmMap[sanitizedCharm]; ok {
		return charm, nil
	}

	return nil, eris.Errorf("Charm %s not registered", name)
}

type Imbuement struct {
	ID   int
	Name string
}

var imbuementMap = map[string]*Imbuement{
	"powerful bash":           {ID: 1, Name: "Powerful Bash"},
	"powerful blockade":       {ID: 2, Name: "Powerful Blockade"},
	"powerful chop":           {ID: 3, Name: "Powerful Chop"},
	"powerful cloud fabric":   {ID: 4, Name: "Powerful Cloud Fabric"},
	"powerful demon presence": {ID: 5, Name: "Powerful Demon Presence"},
	"powerful dragon hide":    {ID: 6, Name: "Powerful Dragon Hide"},
	"powerful electrify":      {ID: 7, Name: "Powerful Electrify"},
	"powerful epiphany":       {ID: 8, Name: "Powerful Epiphany"},
	"powerful featherweight":  {ID: 9, Name: "Powerful Featherweight"},
	"powerful frost":          {ID: 10, Name: "Powerful Frost"},
	"powerful lich shroud":    {ID: 11, Name: "Powerful Lich Shroud"},
	"powerful precision":      {ID: 12, Name: "Powerful Precision"},
	"powerful punch":          {ID: 13, Name: "Powerful Punch"},
	"powerful quara scale":    {ID: 14, Name: "Powerful Quara Scale"},
	"powerful reap":           {ID: 15, Name: "Powerful Reap"},
	"powerful scorch":         {ID: 16, Name: "Powerful Scorch"},
	"powerful slash":          {ID: 17, Name: "Powerful Slash"},
	"powerful snake skin":     {ID: 18, Name: "Powerful Snake Skin"},
	"powerful strike":         {ID: 19, Name: "Powerful Strike"},
	"powerful swiftness":      {ID: 20, Name: "Powerful Swiftness"},
	"powerful vampirism":      {ID: 21, Name: "Powerful Vampirism"},
	"powerful venom":          {ID: 22, Name: "Powerful Venom"},
	"powerful vibrancy":       {ID: 23, Name: "Powerful Vibrancy"},
	"powerful void":           {ID: 24, Name: "Powerful Void"},
}

func NewImbuementFromName(name string) (*Imbuement, error) {
	lowerCaseImbuement := strings.ToLower(name)

	if imbuement, ok := imbuementMap[lowerCaseImbuement]; ok {
		return imbuement, nil
	}

	return nil, eris.Errorf("Imbuement %s not registered", name)
}

type Auction struct {
	ID               int64
	AuctionID        int
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
	Charms           []*Charm
	Imbuements       []*Imbuement
	WorldTransfer    bool
	BossPoints       int
	CharmExpansion   bool
	CharmPoints      int
	TaskExpansion    bool
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
