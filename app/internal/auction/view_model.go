package auction

type PaginatedAuctions struct {
	ViewModels []*AuctionViewModel
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
}

type AuctionViewModel struct {
	Auction *Auction
	ZScore  float64
}

func (av *AuctionViewModel) IsGoodDeal() bool {
	return av.ZScore <= -1
}

func (av *AuctionViewModel) IsBadDeal() bool {
	return av.ZScore >= 1.5
}
