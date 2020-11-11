package transaction

type (
	Transaction struct {
		ConsensusTimestamp   string `json:"consensus_timestamp"`
		TransactionHash      string `json:"transaction_hash"`
		ValidStartTimestamp  string `json:"valid_start_timestamp"`
		ChargedTxFee         int    `json:"charged_tx_fee"`
		MemoBase64           string `json:"memo_base64"`
		Result               string `json:"result"`
		Name                 string `json:"name"`
		MaxFee               string `json:"max_fee"`
		ValidDurationSeconds string `json:"valid_duration_seconds"`
		Node                 string `json:"node"`
		TransactionID        string `json:"transaction_id"`
		Transfers            []Transfer
	}
	Transfer struct {
		Account string `json:"config"`
		Amount  int64  `json:"amount"`
	}
	Transactions struct {
		Transactions []Transaction
	}
)
