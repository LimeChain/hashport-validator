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
	timestampOne := timestamp.NewTimestamp(tm[i].TransactionTimestampWhole, tm[i].TransactionTimestampDec)
	timestampTwo := timestamp.NewTimestamp(tm[j].TransactionTimestampWhole, tm[j].TransactionTimestampDec)
	return timestampOne.ToString() < timestampTwo.ToString()
}
