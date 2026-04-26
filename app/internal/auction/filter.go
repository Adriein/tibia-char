package auction

import (
	"fmt"
	"slices"
	"strings"

	"github.com/adriein/tibia-char/pkg/enums"
	"github.com/adriein/tibia-char/pkg/helper/numbers"
)

type UrlQueryParamsDto struct {
	Page  int      `form:"pag"`
	Qty   int      `form:"qty"`
	Flags []string `form:"flg"`
}

const QueryFilter = "filter"

type SortField string

const (
	SortByEndTime  SortField = "end_time"
	SortByGoodDeal SortField = "good_deal"
	SortByFlags    SortField = "flags"
)

func (sf SortField) sql() string {
	switch sf {
	case SortByEndTime:
		return "a.ta_auction_end"
	case SortByGoodDeal:
		return "a.ta_current_bid"
	default:
		return "a.ta_auction_end"
	}
}

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

func (so SortOrder) sql() string {
	switch so {
	case SortOrderAsc:
		return "ASC"
	case SortOrderDesc:
		return "DESC"
	default:
		return "ASC"
	}
}

type PaginationParam string

const (
	Page PaginationParam = "page"
	Qty  PaginationParam = "qty"
)

type FilterPagination struct {
	Limit int
	Page  int
}

type FilterParam string

const (
	GoodDealFilter      FilterParam = "gdeal"
	RareOutfitFilter    FilterParam = "routfit"
	AuctionStatusFilter FilterParam = "status"
)

type AuctionFilter struct {
	Pagination *FilterPagination
	GoodDeal   bool
	RareOutfit bool
	SortBy     SortField
	SortOrder  SortOrder
}

func FilterFromQueryParams(dto *UrlQueryParamsDto) (*AuctionFilter, error) {
	pagination, err := buildPagination(dto)

	if err != nil {
		return nil, err
	}

	filter := AuctionFilter{
		Pagination: pagination,
		SortBy:     SortByEndTime,
		SortOrder:  SortOrderAsc,
	}

	goodDealFilter := slices.Contains(dto.Flags, string(GoodDealFilter))

	if goodDealFilter {
		filter = AuctionFilter{
			Pagination: pagination,
			GoodDeal:   goodDealFilter,
			SortBy:     SortByEndTime,
			SortOrder:  SortOrderDesc,
		}
	}

	return &filter, nil
}

func buildPagination(dto *UrlQueryParamsDto) (*FilterPagination, error) {
	pagination := &FilterPagination{
		Limit: numbers.DefaultInt(dto.Qty, 20),
		Page:  numbers.DefaultInt(dto.Page, 1) - 1,
	}

	return pagination, nil
}

func (f *AuctionFilter) ToSQL() (string, string, []any, error) {
	var whereClauses []string
	var orderParts []string
	var args []any

	argIndex := 1

	whereClauses = append(whereClauses, "tar.tar_status = 'active'")

	if f.GoodDeal {
		gdWhere := fmt.Sprintf("EXISTS (SELECT 1 FROM tc_auction_flags WHERE taf_auction_id = a.ta_auction_id AND taf_flag_id = $%d)", argIndex)
		whereClauses = append(whereClauses, gdWhere)
		args = append(args, enums.GoodDeal)
		argIndex++
	}

	whereClause := strings.Join(whereClauses, " AND ")

	orderParts = append(orderParts, f.SortBy.sql()+" "+f.SortOrder.sql())

	orderClause := " ORDER BY " + strings.Join(orderParts, ", ")

	return whereClause, orderClause, args, nil
}
