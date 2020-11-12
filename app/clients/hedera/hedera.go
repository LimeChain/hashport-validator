package hedera

import (
	"encoding/json"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"io/ioutil"
	"net/http"
)

type HederaClient struct {
	mirrorAPIAddress string
	mirrorClient     *hedera.MirrorClient
	httpClient       *http.Client
}

func (c HederaClient) GetMirrorClient() *hedera.MirrorClient {
	return c.mirrorClient
}

func NewHederaClient(mirrorNodeAPIAddress, mirrorNodeClientAddress string) *HederaClient {
	mirrorClient, e := hedera.NewMirrorClient(mirrorNodeClientAddress)
	if e != nil {
		panic(e)
	}

	return &HederaClient{
		mirrorAPIAddress: mirrorNodeAPIAddress,
		mirrorClient:     &mirrorClient,
		httpClient:       &http.Client{},
	}
}

func (c HederaClient) GetAccountTransactionsAfterDate(accountId hedera.AccountID, milestoneTimestamp string) (*transaction.HederaTransactions, error) {
	mirrorNodeApiTransactionAddress := fmt.Sprintf("%s%s", c.mirrorAPIAddress, "transactions")
	transactionsDownloadQuery := fmt.Sprintf("%s?account.id=%s&type=credit&result=success&timestamp=gt:%s&order=asc",
		mirrorNodeApiTransactionAddress,
		accountId.String(),
		milestoneTimestamp)

	response, e := c.httpClient.Get(transactionsDownloadQuery)
	if e != nil {
		return nil, e
	}

	defer response.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(response.Body)

	var transactions *transaction.HederaTransactions
	failed := json.Unmarshal(bodyBytes, &transactions)
	if failed != nil {
		return nil, failed
	}

	return transactions, nil
}
