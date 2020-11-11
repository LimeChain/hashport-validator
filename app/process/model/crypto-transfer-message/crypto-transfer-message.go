package crypto_transfer_message

type (
	CryptoTransferMessage struct {
		TxMemo string
		Sender string
		Amount int64
	}
)
