package scrap

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
