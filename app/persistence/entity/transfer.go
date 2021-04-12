package entity

type Transfer struct {
	TransactionID      string `gorm:"primaryKey"`
	Receiver           string
	NativeToken        string
	WrappedToken       string
	Amount             string
	TxReimbursement    string
	Status             string
	SignatureMsgStatus string
	EthTxMsgStatus     string
	EthTxStatus        string
	EthTxHash          string
	Messages           []Message `gorm:"foreignKey:TransferID"`
}
