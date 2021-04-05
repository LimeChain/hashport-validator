package transfer

// Transfer Statuses
const (
	// StatusInitial is the first status on Transfer Record creation
	StatusInitial = "INITIAL"
	// StatusInsufficientFee is a status set once transfer is made but the provided TX
	// reimbursement is not enough for validators to process it. This is a terminal status
	StatusInsufficientFee = "INSUFFICIENT_FEE"
	// StatusInProgress is a status set once the transfer is accepted and the process
	// of bridging the asset has started
	StatusInProgress = "IN_PROGRESS"
	// StatusCompleted is a status set once the Transfer operation is successfully finished.
	// This is a terminal status
	StatusCompleted = "COMPLETED"
	// StatusRecovered is a status set when a transfer has not been processed yet,
	// but has been found by the recovery service
	StatusRecovered = "RECOVERED"
	// StatusFailed is a status set when an ethereum transaction is reverted
	StatusFailed = "FAILED"

	// StatusSignatureSubmitted is a SignatureStatus set once the signature is submitted to HCS
	StatusSignatureSubmitted = "SIGNATURE_SUBMITTED"
	// StatusSignatureMined is a SignatureStatus set once the signature submission TX is successfully mined.
	// This is a terminal status
	StatusSignatureMined = "SIGNATURE_MINED"
	// StatusSignatureFailed is a SignatureStatus set if the signature submission TX fails.
	// This is a terminal status
	StatusSignatureFailed = "SIGNATURE_FAILED"

	// StatusEthTxSubmitted is a EthTxStatus set once the Ethereum transaction is submitted to the Ethereum network
	StatusEthTxSubmitted = "ETH_TX_SUBMITTED"
	// StatusEthTxMined is a EthTxStatus set once the Ethereum transaction is successfully mined.
	// This is a terminal status
	StatusEthTxMined = "ETH_TX_MINED"
	// StatusEthTxReverted is a EthTxStatus set if the Ethereum transaction reverts.
	// This is a terminal status
	StatusEthTxReverted = "ETH_TX_REVERTED"

	// StatusEthTxMsgSubmitted is a EthTxMsgStatus set once the `Ethereum TX Hash` is submitted to HCS
	StatusEthTxMsgSubmitted = "ETH_TX_MSG_SUBMITTED"
	// StatusEthTxMsgMined is a EthTxMsgStatus set once the `Ethereum TX Hash` HCS message is mined.
	// This is a terminal status
	StatusEthTxMsgMined = "ETH_TX_MSG_MINED"
	// StatusEthTxMsgFailed is a EthTxMsgStatus set once the `Ethereum TX Hash` HCS message fails
	// This is a terminal status
	StatusEthTxMsgFailed = "ETH_TX_MSG_FAILED"
)
