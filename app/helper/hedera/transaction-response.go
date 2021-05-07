package hedera

import "github.com/hashgraph/hedera-sdk-go/v2"

type WrappedTransactionResponse struct {
	r *hedera.TransactionResponse
}

func NewWrappedTransactionResponse(response hedera.TransactionResponse) *WrappedTransactionResponse {
	return &WrappedTransactionResponse{r: &response}
}

func (tr *WrappedTransactionResponse) GetReceipt(client *hedera.Client) (hedera.TransactionReceipt, error) {
	return tr.r.GetReceipt(client)
}

func (tr *WrappedTransactionResponse) GetRecord(client *hedera.Client) (hedera.TransactionRecord, error) {
	return tr.r.GetRecord(client)
}

func (tr *WrappedTransactionResponse) GetTransactionID() hedera.TransactionID {
	return tr.r.TransactionID
}
