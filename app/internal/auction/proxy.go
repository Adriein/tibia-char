package auction

import (
	"fmt"

	"github.com/adriein/tibia-char/pkg/helper/collections"
)

type ProxyWorkload map[string][]collections.KeyValue[int, string]

type ProxyManager struct {
	ProxyAddress []string
}

func NewProxyManager() *ProxyManager {
	proxyURL1 := fmt.Sprintf("http://%s:%s@%s:%d", "ibnnxcva", "fl36k8kwqjcg", "142.111.48.253", 7030)
	proxyURL2 := fmt.Sprintf("http://%s:%s@%s:%d", "ibnnxcva", "fl36k8kwqjcg", "23.95.150.145", 6114)
	return &ProxyManager{ProxyAddress: []string{proxyURL1, proxyURL2}}
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
