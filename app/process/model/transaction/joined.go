package transaction

// TODO not necessary anymore - remove
type JoinedTxnMessage struct {
	TransactionId string
	EthAddress    string
	Amount        string
	Fee           string
	Signature     string
	GasPriceGwei  string
	Asset         string
}

// TODO not necessary anymore - remove
type CTMKey struct {
	TransactionId string
	EthAddress    string
	Amount        string
	Fee           string
	GasPriceGwei  string
	Asset         string
}
