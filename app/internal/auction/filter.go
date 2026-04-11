package auction

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/adriein/tibia-char/pkg/enums"
	"github.com/rotisserie/eris"
)

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
	OnlyGoodDeal FilterParam = "gdeal"
)

type AuctionFilter struct {
	Pagination   *FilterPagination
	OnlyGoodDeal bool
	SortBy       SortField
	SortOrder    SortOrder
}

func FilterFromQueryParams(qDict map[string]string) (*AuctionFilter, error) {
	pagination, err := buildPagination(qDict)

	if err != nil {
		return nil, err
	}

	onlyGoodDeal, err := getBool(qDict, string(OnlyGoodDeal), false)

	if err != nil {
		return nil, err
	}

	if onlyGoodDeal {
		return &AuctionFilter{
			Pagination:   pagination,
			OnlyGoodDeal: onlyGoodDeal,
			SortBy:       SortByGoodDeal,
			SortOrder:    SortOrderDesc,
		}, nil
	}

	return &AuctionFilter{
		Pagination:   pagination,
		OnlyGoodDeal: onlyGoodDeal,
		SortBy:       SortByEndTime,
		SortOrder:    SortOrderAsc,
	}, nil
}

func getInt(m map[string]string, key string, fallback int) (int, error) {
	val, ok := m[key]

	if !ok {
		return fallback, nil
	}

	i, err := strconv.Atoi(val)

	if err != nil {
		return fallback, eris.Wrap(err, "cannot convert string to int")
	}

	return i, nil
}

func getBool(m map[string]string, key string, fallback bool) (bool, error) {
	val, ok := m[key]

	if !ok {
		return fallback, nil
	}

	i, err := strconv.ParseBool(val)

	if err != nil {
		return fallback, eris.Wrap(err, "cannot convert string to bool")
	}

	return i, nil
}

func buildPagination(qDict map[string]string) (*FilterPagination, error) {
	page, err := getInt(qDict, string(Page), 1)

	if err != nil {
		return nil, err
	}

	qty, err := getInt(qDict, string(Qty), 20)

	if err != nil {
		return nil, err
	}

	pagination := &FilterPagination{
		Limit: qty,
		Page:  page - 1,
	}

	return pagination, nil
}

func (f *AuctionFilter) ToSQL() (string, string, []any, error) {
	var whereClauses []string
	var orderParts []string
	var args []any

	argIndex := 1

	whereClauses = append(whereClauses, "tar.tar_status = 'active'")

	if f.OnlyGoodDeal {
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
