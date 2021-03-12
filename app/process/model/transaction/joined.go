package transaction

type JoinedTxnMessage struct {
	TransactionId string
	EthAddress    string
	Amount        string
	Fee           string
	Signature     string
	GasPriceGwei  string
	Asset         string
}

type CTMKey struct {
	TransactionId string
	EthAddress    string
	Amount        string
	Fee           string
	GasPriceGwei  string
	Asset         string
}
