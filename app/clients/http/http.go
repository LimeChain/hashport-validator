package http

import (
	"encoding/json"
	"fmt"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"io/ioutil"
	"net/http"
)

type Client struct {
	client            *http.Client
	mirrorNodeAddress string
}

func NewClient(mirrorNodeAddress string) *Client {
	return &Client{
		client:            &http.Client{},
		mirrorNodeAddress: mirrorNodeAddress,
	}
}

func (http Client) GetTransactionsByAccountIdAndTimestamp(account hederasdk.AccountID, lastProcessedTimestamp string) (*transaction.Transactions, error) {
	address := fmt.Sprintf("%s%s", http.mirrorNodeAddress, "transactions")
	accountLink := fmt.Sprintf("%s?account.id=%s&type=credit&result=success&timestamp=gt:%s&order=asc",
		address,
		account.String(),
		lastProcessedTimestamp)

	response, e := http.client.Get(accountLink)
	if e != nil {
		return nil, e
	}

	defer response.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(response.Body)

	var transactions *transaction.Transactions
	failed := json.Unmarshal(bodyBytes, &transactions)
	if failed != nil {
		return nil, failed
	}

	return transactions, nil
}
