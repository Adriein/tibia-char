package scrap

import (
	"time"
)

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

type BazaarAuctionDetailSet map[int]BazaarCharAuctionDetail

func (set BazaarAuctionDetailSet) Get(key int) (BazaarCharAuctionDetail, bool) {
	value, ok := set[key]
	return value, ok
}

func (set BazaarAuctionDetailSet) Set(key int, value BazaarCharAuctionDetail) {
	set[key] = value
}

func (set BazaarAuctionDetailSet) Del(key int) {
	delete(set, key)
}

func (set BazaarAuctionDetailSet) Has(key int) bool {
	_, ok := set[key]
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
