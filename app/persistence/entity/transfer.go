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
		t.SourceAsset == comparing.SourceAsset &&
		t.TargetAsset == comparing.TargetAsset &&
		t.Amount == comparing.Amount &&
		t.TxReimbursement == comparing.TxReimbursement &&
		t.GasPrice == comparing.GasPrice &&
		t.Status == comparing.Status &&
		t.SignatureMsgStatus == comparing.SignatureMsgStatus &&
		t.EthTxMsgStatus == comparing.EthTxMsgStatus &&
		t.EthTxStatus == comparing.EthTxStatus &&
		t.EthTxHash == comparing.EthTxHash &&
		t.ExecuteEthTransaction == comparing.ExecuteEthTransaction
}
