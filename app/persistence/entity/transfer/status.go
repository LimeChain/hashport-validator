package transfer

// Transfer Statuses
const (
	// StatusInitial is the first status on Transfer Record creation
	StatusInitial = "INITIAL"
	// StatusInProgress is a status set once the transfer is accepted and the process
	// of bridging the asset has started
	StatusInProgress = "IN_PROGRESS"
	// StatusCompleted is a status set once the Transfer operation is successfully finished.
	// This is a terminal status
	StatusCompleted = "COMPLETED"
	// StatusRecovered is a status set when a transfer has not been processed yet,
	// but has been found by the recovery service
	StatusRecovered = "RECOVERED"
	// StatusSignatureSubmitted is a SignatureStatus set once the signature is submitted to HCS
	StatusSignatureSubmitted = "SIGNATURE_SUBMITTED"
	// StatusSignatureMined is a SignatureStatus set once the signature submission TX is successfully mined.
	// This is a terminal status
	StatusSignatureMined = "SIGNATURE_MINED"
	// StatusSignatureFailed is a SignatureStatus set if the signature submission TX fails.
	// This is a terminal status
	StatusSignatureFailed = "SIGNATURE_FAILED"
	// StatusScheduledTokenBurnSubmitted is a ScheduledTokenBurn status set if the scheduled token burn TX gets submitted successfully.
	StatusScheduledTokenBurnSubmitted = "SCHEDULED_TOKEN_BURN_SUBMITTED"
	// StatusScheduledTokenBurnFailed is a ScheduledTokenBurn status set if the scheduled token burn TX fails.
	// This is a terminal status
	StatusScheduledTokenBurnFailed = "SCHEDULED_TOKEN_BURN_FAILED"
	// StatusScheduledTokenBurnCompleted is a ScheduledTokenBurn status set if the scheduled token burn TX completes successfully.
	// This is a terminal status
	StatusScheduledTokenBurnCompleted = "SCHEDULED_TOKEN_BURN_COMPLETED"
)
