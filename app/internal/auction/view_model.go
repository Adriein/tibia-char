package auction

import (
	"net/url"
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

func (fp *FilterParams) ToQueryParams() string {
	v := url.Values{}

	for _, flag := range fp.Flags {
		v.Add("flg", flag)
	}

	for _, status := range fp.Status {
		v.Add("sts", status)
	}

	for _, voc := range fp.Vocation {
		v.Add("voc", voc)
	}

	return v.Encode()
}

type AuctionViewModel struct {
	Auction         *Auction
	AucEndFormatted string
	LastUpdated     string
	TimeLeft        string
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
