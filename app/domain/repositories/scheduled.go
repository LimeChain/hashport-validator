package repositories

type ScheduledRepository interface {
	Create(amount int64, nonce, recipient, bridgeThresholdAccountID, payerAccountID string) error
	UpdateStatusSubmitted(nonce, scheduleID, submissionTxId string) error
	UpdateStatusCompleted(txId string) error
	UpdateStatusFailed(txId string) error
}
