package sync

type Statuses string

const (
	SCHEDULED_MINT_TYPE     = "mint"
	SCHEDULED_BURN_TYPE     = "burn"
	SCHEDULED_TRANSFER_TYPE = "transfer"
	DONE                    = "DONE"
	FAIL                    = "FAIL"
)
