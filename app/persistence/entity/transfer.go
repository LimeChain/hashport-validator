package entity

type Transfer struct {
	TransactionID      string `gorm:"primaryKey"`
	Receiver           string
	NativeToken        string
	WrappedToken       string
	Amount             string
	Status             string
	SignatureMsgStatus string
	Messages           []Message `gorm:"foreignKey:TransferID"`
}
