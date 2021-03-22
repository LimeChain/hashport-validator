package transaction

type Data struct {
	Recipient    string   `json:"recipient"`
	Amount       string   `json:"amount"`
	ERC20Address string   `json:"erc20Address"`
	Fee          string   `json:"fee"`
	GasPrice     string   `json:"gasPrice"`
	Signatures   []string `json:"signatures"`
	Majority     bool     `json:"majority"`
}
