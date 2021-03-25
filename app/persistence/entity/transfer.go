package entity

type Transfer struct {
	TransactionID         string `gorm:"primaryKey; not null; autoIncrement:false"`
	Receiver              string
	SourceAsset           string
	TargetAsset           string
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
