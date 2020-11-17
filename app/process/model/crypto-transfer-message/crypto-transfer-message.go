package cryptotransfermessage

type (
	CryptoTransferMessage struct {
		EthAddress string
		TxId       string
		Amount     int64
		TxFee      string
		Sender     string
	}
)
