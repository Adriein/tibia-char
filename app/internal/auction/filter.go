package auction

type SortField string

const (
	SortByEndTime SortField = "end_time"
)

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

type FilterPagination struct {
	Limit     int
	Page      int
	SortBy    SortField
	SortOrder SortOrder
}

type AuctionFilter struct {
	Pagination *FilterPagination
	IsGoodDeal bool
}

func DefaultAuctionFilter() *AuctionFilter {
	return &AuctionFilter{
		Pagination: &FilterPagination{
			Limit:     20,
			Page:      0,
			SortBy:    SortByEndTime,
			SortOrder: SortOrderAsc,
		},
	}
}
