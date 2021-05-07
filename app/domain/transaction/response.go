package transaction

import "github.com/hashgraph/hedera-sdk-go/v2"

type Response interface {
	GetReceipt(client *hedera.Client) (hedera.TransactionReceipt, error)
	GetRecord(client *hedera.Client) (hedera.TransactionRecord, error)
	GetTransactionID() hedera.TransactionID
}
