package auction

import (
	"strconv"

	"github.com/rotisserie/eris"
)

const QueryFilter = "f"

type SortField string

const (
	SortByEndTime SortField = "end_time"
	SortByFlags   SortField = "flags"
)

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

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
			SortBy:       SortByFlags,
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
