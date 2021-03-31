package entity

type Message struct {
	TransferID           string
	Transfer             Transfer `gorm:"foreignKey:TransferID;references:TransactionID;"`
	Hash                 string
	Signature            string `gorm:"unique"`
	Signer               string
	TransactionTimestamp int64
}
