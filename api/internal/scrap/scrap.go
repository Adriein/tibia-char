package scrap

type BazaarAuctionLinkSet map[int]string

func (v BazaarAuctionLinkSet) Get(key string) string {
	vs := v[key]
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

// Set sets the key to value. It replaces any existing
// values.
func (v BazaarAuctionLinkSet) Set(key, value string) {
	v[key] = []string{value}
}

// Add adds the value to key. It appends to any existing
// values associated with key.
func (v BazaarAuctionLinkSet) Add(key, value string) {
	v[key] = append(v[key], value)
}

// Del deletes the values associated with key.
func (v BazaarAuctionLinkSet) Del(key string) {
	delete(v, key)
}

// Has checks whether a given key is set.
func (v BazaarAuctionLinkSet) Has(key string) bool {
	_, ok := v[key]
	return ok
}
