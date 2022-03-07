package mirror_node

import "github.com/shopspring/decimal"

type UpdatedFileRateData struct {
	CurrentRate decimal.Decimal
	NextRate    decimal.Decimal
}

type TransactionsResponse struct {
	Transactions []Transaction     `json:"transactions"`
	Links        map[string]string `json:"links"`
}

type Transaction struct {
	Bytes                    interface{}   `json:"bytes"`
	ChargedTxFee             int           `json:"charged_tx_fee"`
	ConsensusTimestamp       string        `json:"consensus_timestamp"`
	EntityId                 string        `json:"entity_id"`
	MaxFee                   string        `json:"max_fee"`
	MemoBase64               string        `json:"memo_base64"`
	Name                     string        `json:"name"`
	Node                     string        `json:"node"`
	Nonce                    int           `json:"nonce"`
	ParentConsensusTimestamp string        `json:"parent_consensus_timestamp"`
	Result                   string        `json:"result"`
	Scheduled                bool          `json:"scheduled"`
	TransactionHash          string        `json:"transaction_hash"`
	TransactionId            string        `json:"transaction_id"`
	Transfers                []interface{} `json:"transfers"`
	ValidDurationSeconds     string        `json:"valid_duration_seconds"`
	ValidStartTimestamp      string        `json:"valid_start_timestamp"`
}
