package hedera

import (
	"sync"
)

// AwaitMultipleMinedScheduledTransactions is meant to be used when you need to wait all the scheduled transactions to be mined.
func AwaitMultipleMinedScheduledTransactions(
	wg *sync.WaitGroup,
	outTransactionsResults []*bool,
	sourceChainId int64,
	targetChainId int64,
	asset string,
	transferID string,
	callback func(sourceChainId int64, targetChainId int64, nativeAsset string, transferID string, isTransferSuccessful bool)) {

	wg.Wait()

	isTransferSuccessful := true
	for _, result := range outTransactionsResults {
		if result != nil && *result == false {
			isTransferSuccessful = false
			break
		}
	}

	callback(sourceChainId, targetChainId, asset, transferID, isTransferSuccessful)
}
