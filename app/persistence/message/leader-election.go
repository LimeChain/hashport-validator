package message

import "github.com/limechain/hedera-eth-bridge-validator/app/process/model/timestamp"

type ByTimestamp []TransactionMessage

func (tm ByTimestamp) Len() int {
	return len(tm)
}
func (tm ByTimestamp) Swap(i, j int) {
	tm[i], tm[j] = tm[j], tm[i]
}
func (tm ByTimestamp) Less(i, j int) bool {
	timestampOne := timestamp.NewTimestamp(tm[i].TransactionTimestampSeconds, tm[i].TransactionTimestampNanoseconds)
	timestampTwo := timestamp.NewTimestamp(tm[j].TransactionTimestampSeconds, tm[j].TransactionTimestampNanoseconds)
	return timestampOne.ToString() < timestampTwo.ToString()
}
