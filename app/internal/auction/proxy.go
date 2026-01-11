package auction

import "github.com/adriein/tibia-char/pkg/helper/collections"

type ProxyWorkload map[string][]collections.KeyValue[int, string]

type ProxyManager struct {
	ProxyAddress []string
}

func NewProxyManager() *ProxyManager {
	return &ProxyManager{ProxyAddress: []string{"fakeProxy1", "fakeProxy2"}}
}

func (pm *ProxyManager) BalanceLoad(auctions map[int]string) ProxyWorkload {
	if len(pm.ProxyAddress) == 0 {
		return ProxyWorkload{}
	}

	chunks := collections.ChunkMap(auctions, len(pm.ProxyAddress))
	result := make(ProxyWorkload, len(pm.ProxyAddress))

	for i, chunk := range chunks {
		proxyIdx := i % len(pm.ProxyAddress)
		proxyAddr := pm.ProxyAddress[proxyIdx]

		result[proxyAddr] = append(result[proxyAddr], chunk...)
	}

	return result
}
