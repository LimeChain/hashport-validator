package message

type ByTimestamp []TransactionMessage

func (tm ByTimestamp) Len() int {
	return len(tm)
}
func (tm ByTimestamp) Swap(i, j int) {
	tm[i], tm[j] = tm[j], tm[i]
}
func (tm ByTimestamp) Less(i, j int) bool {
	return tm[i].TransactionTimestamp < tm[j].TransactionTimestamp
}
