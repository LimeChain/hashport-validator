package constants

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/model/pricing"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/shopspring/decimal"
	"math/big"
)

var (
	// Hedera //

	HbarCoinGeckoId              = "hedera-hashgraph"
	HbarCoinMarketCapId          = "4642"
	HbarPriceInUsd               = decimal.NewFromFloat(0.2)
	HbarMinAmountWithFee         = big.NewInt(5000000000)
	HbarMinAmountWithFeeInEVM, _ = big.NewInt(0).SetString("50000000000000000000", 10)

	// Ethereum //

	EthereumCoinGeckoId                 = "ethereum"
	EthereumCoinMarketCapId             = "1027"
	EthereumNativeTokenPriceInUsd       = decimal.NewFromFloat(8000.0)
	EthereumNativeTokenMinAmountWithFee = big.NewInt(1250000000000000)

	UsdPrices = map[uint64]map[string]decimal.Decimal{
		constants.HederaNetworkId: {
			constants.Hbar: HbarPriceInUsd,
		},
		EthereumNetworkId: {
			NetworkEthereumFungibleNativeToken: EthereumNativeTokenPriceInUsd,
		},
	}

	CoinGeckoIds = map[uint64]map[string]string{
		constants.HederaNetworkId: {
			constants.Hbar: HbarCoinGeckoId,
		},
		EthereumNetworkId: {
			NetworkEthereumFungibleNativeToken: EthereumCoinGeckoId,
		},
	}

	CoinMarketCapIds = map[uint64]map[string]string{
		constants.HederaNetworkId: {
			constants.Hbar: HbarCoinMarketCapId,
		},
		EthereumNetworkId: {
			NetworkEthereumFungibleNativeToken: EthereumCoinMarketCapId,
		},
	}

	TokenPriceInfos = map[uint64]map[string]pricing.TokenPriceInfo{
		constants.HederaNetworkId: {
			constants.Hbar: pricing.TokenPriceInfo{
				UsdPrice:         HbarPriceInUsd,
				MinAmountWithFee: HbarMinAmountWithFee,
			},
		},
		EthereumNetworkId: {
			NetworkEthereumFungibleNativeToken: pricing.TokenPriceInfo{
				UsdPrice:         EthereumNativeTokenPriceInUsd,
				MinAmountWithFee: EthereumNativeTokenMinAmountWithFee,
			},
			NetworkEthereumFungibleWrappedTokenForNetworkHedera: pricing.TokenPriceInfo{
				UsdPrice:         HbarPriceInUsd,
				MinAmountWithFee: HbarMinAmountWithFeeInEVM,
			},
		},

		PolygonNetworkId: {
			NetworkPolygonFungibleWrappedTokenForNetworkHedera: pricing.TokenPriceInfo{
				UsdPrice:         HbarPriceInUsd,
				MinAmountWithFee: HbarMinAmountWithFeeInEVM,
			},
			NetworkPolygonFungibleWrappedTokenForNetworkEthereum: pricing.TokenPriceInfo{
				UsdPrice:         EthereumNativeTokenPriceInUsd,
				MinAmountWithFee: EthereumNativeTokenMinAmountWithFee,
			},
		},
	}

	MinAmountsForApi = map[uint64]map[string]string{
		constants.HederaNetworkId: {
			constants.Hbar: HbarMinAmountWithFee.String(),
		},
		EthereumNetworkId: {
			NetworkEthereumFungibleNativeToken:                  EthereumNativeTokenMinAmountWithFee.String(),
			NetworkEthereumFungibleWrappedTokenForNetworkHedera: HbarMinAmountWithFeeInEVM.String(),
		},
		PolygonNetworkId: {
			NetworkPolygonFungibleWrappedTokenForNetworkEthereum: EthereumNativeTokenMinAmountWithFee.String(),
			NetworkPolygonFungibleWrappedTokenForNetworkHedera:   HbarMinAmountWithFeeInEVM.String(),
		},
	}
)
