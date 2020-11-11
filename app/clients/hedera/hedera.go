package hedera

import (
	"encoding/json"
	"fmt"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"io/ioutil"
	"log"
	"net/http"
)

type Client struct {
	mirrorAPIAddress string
	mirrorClient     *hederasdk.MirrorClient
	httpClient       *http.Client
}

func (hedera Client) GetMirror() *hederasdk.MirrorClient {
	return hedera.mirrorClient
}

func NewClient(mirrorAPIAddress, mirrorNodeClientAddress string) *Client {
	mirrorClient, e := hederasdk.NewMirrorClient(mirrorNodeClientAddress)
	if e != nil {
		log.Printf("Error: Could not instantiate mirror client - [%s]", e)
	}

	return &Client{
		mirrorAPIAddress: mirrorAPIAddress,
		mirrorClient:     &mirrorClient,
		httpClient:       &http.Client{},
	}
}

func (hedera Client) GetTransactionsByAccountIdAndTimestamp(account hederasdk.AccountID, lastProcessedTimestamp string) (*transaction.Transactions, error) {
	address := fmt.Sprintf("%s%s", hedera.mirrorAPIAddress, "transactions")
	accountLink := fmt.Sprintf("%s?account.id=%s&type=credit&result=success&timestamp=gt:%s&order=asc",
		address,
		account.String(),
		lastProcessedTimestamp)

	response, e := hedera.httpClient.Get(accountLink)
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
