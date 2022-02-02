package constants

const (
	Hedera          = "Hedera"
	Ethereum        = "Ethereum"
	Polygon         = "Polygon"
	EthereumChainId = 3
	PolygonChainId  = 80001
)

var (
	NetworkIdToName = map[uint64]string{
		0:     Hedera,
		3:     Ethereum,
		80001: Polygon,
	}
)
