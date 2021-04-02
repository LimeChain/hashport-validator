package entity

type Transfer struct {
	TransactionID         string `gorm:"primaryKey"`
	Receiver              string
	NativeToken           string
	WrappedToken          string
	Amount                string
	TxReimbursement       string
	GasPrice              string
	Status                string
	SignatureMsgStatus    string
	EthTxMsgStatus        string
	EthTxStatus           string
	EthTxHash             string
	ExecuteEthTransaction bool
	Messages              []Message `gorm:"foreignKey:TransferID"`
}

func (t Transfer) Equals(comparing Transfer) bool {
	return t.TransactionID == comparing.TransactionID &&
		t.Receiver == comparing.Receiver &&
		t.NativeToken == comparing.NativeToken &&
		t.WrappedToken == comparing.WrappedToken &&
		t.Amount == comparing.Amount &&
		t.TxReimbursement == comparing.TxReimbursement &&
		t.GasPrice == comparing.GasPrice &&
		t.Status == comparing.Status &&
		t.SignatureMsgStatus == comparing.SignatureMsgStatus &&
		//t.EthTxMsgStatus == comparing.EthTxMsgStatus && // TODO: Uncomment when ready
		t.EthTxStatus == comparing.EthTxStatus &&
		t.EthTxHash == comparing.EthTxHash &&
		t.ExecuteEthTransaction == comparing.ExecuteEthTransaction
}
