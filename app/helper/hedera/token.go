package hedera

import "github.com/hashgraph/hedera-sdk-go/v2"

func IsTokenID(tokenID string) bool {
	_, err := hedera.TokenIDFromString(tokenID)
	if err != nil {
		return false
	}

	return true
}
