package cryptotransfermessage

type (
	CryptoTransferMessage struct {
		TxMemo string
		Sender string
		Amount int64
	}
)
