package repositories

type Scheduled interface {
	Create(amount int64, nonce, recipient, bridgeThresholdAccountID, payerAccountID string) error
	UpdateStatusSubmitted(nonce, scheduleID, submissionTxId string) error
	UpdateStatusCompleted(txId string) error
	UpdateStatusFailed(txId string) error
}
