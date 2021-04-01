package entity

type Message struct {
	TransferID           string
	Transfer             Transfer `gorm:"foreignKey:TransferID;references:TransactionID;"`
	Hash                 string
	Signature            string `gorm:"unique"`
	Signer               string
	TransactionTimestamp int64
}

func (m Message) Equals(comparing Message) bool {
	return m.TransferID == comparing.TransferID &&
		m.TransactionTimestamp == comparing.TransactionTimestamp &&
		m.Signature == comparing.Signature &&
		m.Hash == comparing.Hash &&
		m.Transfer.Equals(comparing.Transfer) &&
		m.Signer == comparing.Signer
}
