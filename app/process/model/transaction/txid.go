package transaction

import (
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	"strings"
)

type TxId struct {
	AccountId string
	Seconds   string
	Nanos     string
}

func FromHederaTransactionID(id *hedera.TransactionID) TxId {
	stringTxId := id.String()
	split := strings.Split(stringTxId, "@")
	accId := split[0]

	split = strings.Split(split[1], ".")

	return TxId{
		AccountId: accId,
		Seconds:   split[0],
		Nanos:     split[1],
	}
}

func (txId *TxId) String() string {
	return fmt.Sprintf("%s-%s-%s", txId.AccountId, txId.Seconds, txId.Nanos)
}

func (txId *TxId) Timestamp() string {
	return fmt.Sprintf("%s.%s", txId.Seconds, txId.Nanos)
}
