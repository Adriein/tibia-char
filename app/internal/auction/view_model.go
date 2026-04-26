package auction

import (
	"slices"

	"github.com/adriein/tibia-char/pkg/enums"
)

type PaginatedAuctions struct {
	ViewModels []*AuctionViewModel
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
	Filters    FilterParams
}

type FilterParams struct {
	Flags    []string
	Status   []string
	Vocation []string
}

type AuctionViewModel struct {
	Auction *Auction
}

func (av *AuctionViewModel) IsGoodDeal() bool {
	return slices.ContainsFunc(av.Auction.Flags, func(f *Flag) bool {
		return f.ID == enums.GoodDeal
	})
}

func (av *AuctionViewModel) IsBadDeal() bool {
	return slices.ContainsFunc(av.Auction.Flags, func(f *Flag) bool {
		return f.ID == enums.BadDeal
	})
}
