package constants

const (
	Hedera   = "Hedera"
	Ethereum = "Ethereum"
	Polygon  = "Polygon"
)

var (
	NetworkIdToName = map[uint64]string{
		0:     Hedera,
		3:     Ethereum,
		80001: Polygon,
	}
)
