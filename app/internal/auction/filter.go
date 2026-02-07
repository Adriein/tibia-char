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

type AuctionFilter struct {
	Limit     int
	Offset    int
	SortBy    SortField
	SortOrder SortOrder
}

func DefaultAuctionFilter() *AuctionFilter {
	return &AuctionFilter{
		Limit:     20,
		Offset:    0,
		SortBy:    SortByEndTime,
		SortOrder: SortOrderAsc,
	}
}
