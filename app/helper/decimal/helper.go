package decimal

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/model/asset"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/shopspring/decimal"
	"math/big"
)

// ToLowestDenomination decimal amount to the lowest denomination
func ToLowestDenomination(amount decimal.Decimal, decimals uint8) *big.Int {

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result := amount.Mul(mul)

	toSmallestDenomination := new(big.Int)
	toSmallestDenomination.SetString(result.String(), 10)

	return toSmallestDenomination
}

func GetAmountInUsd(price decimal.Decimal, amount *big.Int, assetsService service.Assets, nativeAsset *asset.NativeAsset) *big.Int {
	var decimals uint8
	nativeFungibleAssetInfo, exist := assetsService.GetFungibleAssetInfo(nativeAsset.ChainId, nativeAsset.Asset)
	if exist {
		decimals = nativeFungibleAssetInfo.Decimals
	} else {
		if nativeAsset.ChainId == constants.HederaNetworkId {
			decimals = constants.HederaDefaultDecimals
		} else {
			decimals = constants.EvmDefaultDecimals
		}
	}

	priceToSmallest := ToLowestDenomination(price, decimals)
	amountInUsd := big.NewInt(0).Mul(amount, priceToSmallest)
	return amountInUsd
}
