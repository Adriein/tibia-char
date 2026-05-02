package auction

import (
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adriein/tibia-char/pkg/constants"
	"github.com/adriein/tibia-char/pkg/enums"
	"github.com/adriein/tibia-char/pkg/helper"
	"github.com/rotisserie/eris"
)

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
	"curse":            {ID: 26, Type: "Major", Name: "Curse"},
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

type Quest struct {
	ID   int
	Name string
}

var questMap = map[string]*Quest{
	"the postman missions":                {ID: 1, Name: "The Postman Missions"},
	"the djinn war (blue)":                {ID: 2, Name: "The Djinn War (blue)"},
	"the djinn war (green)":               {ID: 3, Name: "The Djinn War (green)"},
	"the travelling trader (rashid)":      {ID: 4, Name: "The Travelling Trader (Rashid)"},
	"the thieves guild":                   {ID: 5, Name: "The Thieves Guild"},
	"shadows of yalahar":                  {ID: 6, Name: "Shadows of Yalahar"},
	"the pits of inferno":                 {ID: 7, Name: "The Pits of Inferno"},
	"the inquisition":                     {ID: 8, Name: "The Inquisition"},
	"barbarian test":                      {ID: 9, Name: "Barbarian Test"},
	"lion's rock":                         {ID: 10, Name: "Lion's Rock"},
	"the shattered isles":                 {ID: 11, Name: "The Shattered Isles"},
	"the ice islands":                     {ID: 12, Name: "The Ice Islands"},
	"twenty miles beneath the sea":        {ID: 13, Name: "Twenty Miles Beneath the Sea"},
	"the explorer society":                {ID: 14, Name: "The Explorer Society"},
	"blood brothers":                      {ID: 15, Name: "Blood Brothers"},
	"the new frontier":                    {ID: 16, Name: "The New Frontier"},
	"wrath of the emperor":                {ID: 17, Name: "Wrath of the Emperor"},
	"the ape city":                        {ID: 18, Name: "The Ape City"},
	"rathleton (citzen)":                  {ID: 19, Name: "Rathleton (Citzen)"},
	"dark trails":                         {ID: 20, Name: "Dark Trails"},
	"asura palace":                        {ID: 21, Name: "Asura Palace"},
	"the dream courts":                    {ID: 22, Name: "The Dream Courts"},
	"the secret library":                  {ID: 23, Name: "The Secret Library"},
	"soul war":                            {ID: 24, Name: "Soul War"},
	"primal ordeal":                       {ID: 25, Name: "Primal Ordeal"},
	"rotten blood":                        {ID: 26, Name: "Rotten Blood"},
	"hero of rathleton":                   {ID: 27, Name: "Hero of Rathleton"},
	"cults of tibia":                      {ID: 28, Name: "Cults of Tibia"},
	"the curse spreads":                   {ID: 29, Name: "The Curse Spreads"},
	"grimvale":                            {ID: 30, Name: "Grimvale"},
	"bigfoot's burden (rank iv)":          {ID: 31, Name: "Bigfoot's Burden (Rank IV)"},
	"bigfoot's burden (free boss access)": {ID: 32, Name: "Bigfoot's Burden (Free boss access)"},
	"kilmaresh":                           {ID: 33, Name: "Kilmaresh"},
	"heart of destruction":                {ID: 34, Name: "Heart of Destruction"},
	"feaster of souls":                    {ID: 35, Name: "Feaster of Souls"},
	"dangerous depths (warzone 4)":        {ID: 36, Name: "Dangerous Depths (Warzone 4)"},
	"dangerous depths (warzone 5)":        {ID: 37, Name: "Dangerous Depths (Warzone 5)"},
	"dangerous depths (warzone 6)":        {ID: 38, Name: "Dangerous Depths (Warzone 6)"},
	"ferumbras' ascendant":                {ID: 39, Name: "Ferumbras' Ascendant"},
	"the order of the cobra":              {ID: 40, Name: "The Order of the Cobra"},
	"the order of the lion":               {ID: 41, Name: "The Order of the Lion"},
	"the order of the falcon":             {ID: 42, Name: "The Order of the Falcon"},
}

func NewQuestFromName(name string) (*Quest, error) {
	lowerCaseQuest := strings.ToLower(name)

	sanitizedQuest := strings.ReplaceAll(lowerCaseQuest, "'", "")

	if quest, ok := questMap[sanitizedQuest]; ok {
		return quest, nil
	}

	return nil, eris.Errorf("Quest %s not registered", name)
}

type BidRegistry struct {
	Amount  int
	DateAdd time.Time
}

type Outfit struct {
	Name   string
	Addons int
}

func (o *Outfit) isRare(isMale bool) bool {
	if o.Name == string(enums.Golden) && o.Addons == 2 {
		return true
	}

	if o.Name == string(enums.Royal) && o.Addons == 2 {
		return true
	}

	if isMale && o.Name == string(enums.FeruMale) && o.Addons == 2 {
		return true
	}

	if !isMale && o.Name == string(enums.FeruFemale) && o.Addons == 2 {
		return true
	}

	return false
}

type Mount struct {
	Name string
}

func (m *Mount) isRare() bool {
	return false
}

type Flag struct {
	ID enums.Flag
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
	Quests           []*Quest
	Outfits          []*Outfit
	Mounts           []*Mount
	WorldTransfer    bool
	BossPoints       int
	CharmExpansion   bool
	CharmPoints      int
	TaskExpansion    bool
	Bid              int
	BidFiat          int
	BidCurrency      enums.Currency
	BidRegistry      []*BidRegistry
	Stage            enums.AuctionStage
	AuctionStart     time.Time
	AuctionEnd       time.Time
	Status           enums.AuctionRecordableStatus
	Flags            []*Flag
	DateAdd          time.Time
	DateUpd          time.Time
}

func (a *Auction) IsTrending() bool {
	var increasedTimes int

	for i, registry := range a.BidRegistry {
		if len(a.BidRegistry) <= i+1 {
			break
		}

		next := a.BidRegistry[i+1]

		if registry.Amount < next.Amount {
			increasedTimes++
		}
	}

	return increasedTimes >= 4
}

func (a *Auction) SubsetKey() string {
	var voc string

	switch a.CharVocation.Id {
	case constants.VocationKnight:
		voc = "K"
	case constants.VocationPaladin:
		voc = "P"
	case constants.VocationMonk:
		voc = "M"
	case constants.VocationDruid:
		voc = "D"
	case constants.VocationSorcerer:
		voc = "S"
	default:
		voc = "N"
	}

	lvl := a.getRangeLabel(a.CharLevel, []int{8, 100, 300, 600, 1000, 2000})

	var skillVal int

	switch a.CharVocation.Id {
	case constants.VocationKnight:
		skillVal = max(a.Skills.Axe, a.Skills.Sword, a.Skills.Club)
	case constants.VocationPaladin:
		skillVal = a.Skills.Distance
	case constants.VocationMonk:
		skillVal = a.Skills.Fist
	default:
		skillVal = a.Skills.MagicLevel
	}

	skill := a.getRangeLabel(skillVal, []int{10, 100, 110, 120, 130})

	worldType := a.CharWorld.BattleEye.String()

	hasROutfitFlag := slices.ContainsFunc(a.Flags, func(f *Flag) bool {
		return f.ID == enums.ROutfit
	})

	if hasROutfitFlag {
		return voc + "_" + lvl + "_" + skill + "_" + worldType + "_" + "routfit"
	}

	return voc + "_" + lvl + "_" + skill + "_" + worldType
}

func (a *Auction) getRangeLabel(val int, thresholds []int) string {
	for i := 0; i < len(thresholds)-1; i++ {
		if val >= thresholds[i] && val < thresholds[i+1] {
			return strconv.Itoa(thresholds[i]) + "_" + strconv.Itoa(thresholds[i+1]-1)
		}
	}

	return "+" + strconv.Itoa(thresholds[len(thresholds)-1])
}

func (a *Auction) CalculateFlags(stats *AggAuctionStats) {
	ZScore := helper.SafeDivision(float64(a.Bid-int(stats.Median)), stats.StdDeviation)

	if ZScore <= -1 {
		a.Flags = append(a.Flags, &Flag{ID: enums.GoodDeal})
	}

	if ZScore >= 1.5 {
		a.Flags = append(a.Flags, &Flag{ID: enums.BadDeal})
	}

	isMale := a.CharGender.Id == constants.GenderMale

	for _, outfit := range a.Outfits {
		if !outfit.isRare(isMale) {
			continue
		}

		a.Flags = append(a.Flags, &Flag{ID: enums.ROutfit})
		break
	}

	for _, mount := range a.Mounts {
		if !mount.isRare() {
			continue
		}

		a.Flags = append(a.Flags, &Flag{ID: enums.RMount})
		break
	}

	if a.IsTrending() {
		a.Flags = append(a.Flags, &Flag{ID: enums.Hot})
	}
}

func (a *Auction) IsEqual(other *Auction) bool {
	if other == nil {
		return false
	}

	return a.AuctionEnd.Equal(other.AuctionEnd) &&
		a.Bid == other.Bid &&
		a.TibiaAuctionLink == other.TibiaAuctionLink &&
		a.Stage == other.Stage
}

func (a *Auction) ShouldBeArchived(other *Auction) bool {
	now := time.Now().UTC()
	return other.Status == enums.RecordableActive && now.After(a.AuctionEnd)
}

type AuctionStats struct {
	Median       float64
	StdDeviation float64
}

type AggAuctionStats struct {
	SubsetKey    string
	MinPrice     int
	MaxPrice     int
	Median       float64
	Mean         float64
	StdDeviation float64
	Mode         int
	SampleSize   int
}

type StdDeviationSubsets = map[string]*AuctionStats

type PriceSubsets = map[string][]int

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

func (s *AuctionLinkSet) IsEmpty() bool {
	return len(s.Data) == 0
}

func (s *AuctionLinkSet) AllowLongTail() bool {
	return len(s.Data) < constants.MaxLongTailRefreshAllowance
}
