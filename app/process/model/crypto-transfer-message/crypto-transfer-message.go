package cryptotransfermessage

type (
	CryptoTransferMessage struct {
		EthAddress string
		TxId       string
		Amount     int64
		TxFee      uint64
		Sender     string
	}
)
