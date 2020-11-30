package hedera

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	hcstopicmessage "github.com/limechain/hedera-eth-bridge-validator/app/process/model/hcs-topic-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"io/ioutil"
	"net/http"
)

type HederaMirrorClient struct {
	mirrorAPIAddress string
	httpClient       *http.Client
}

func NewHederaMirrorClient(mirrorNodeAPIAddress string) *HederaMirrorClient {
	return &HederaMirrorClient{
		mirrorAPIAddress: mirrorNodeAPIAddress,
		httpClient:       &http.Client{},
	}
}

func (c HederaMirrorClient) GetSuccessfulAccountCreditTransactionsAfterDate(accountId hedera.AccountID, milestoneTimestamp string) (*transaction.HederaTransactions, error) {
	transactionsDownloadQuery := fmt.Sprintf("?account.id=%s&type=credit&result=success&timestamp=gt:%s&order=asc",
		accountId.String(),
		milestoneTimestamp)
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

func (c HederaMirrorClient) GetAccountTransaction(transactionID string) (*transaction.HederaTransactions, error) {
	transactionsDownloadQuery := fmt.Sprintf("/%s",
		transactionID)
	return c.getTransactionsByQuery(transactionsDownloadQuery)
}

func (c HederaMirrorClient) GetStateProof(transactionID string) ([]byte, error) {
	query := fmt.Sprintf("%s%s%s", c.mirrorAPIAddress, "transactions",
		fmt.Sprintf("/%s/stateproof", transactionID))

	response, e := c.get(query)
	if e != nil {
		return nil, e
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("State Proof HTTP GET for TransactionID [%s] ended with Status Code [%d].", transactionID, response.StatusCode))
	}

	return readResponseBody(response)
}

func (c HederaMirrorClient) get(query string) (*http.Response, error) {
	return c.httpClient.Get(query)
}

func (c HederaMirrorClient) getTransactionsByQuery(query string) (*transaction.HederaTransactions, error) {
	transactionsQuery := fmt.Sprintf("%s%s%s", c.mirrorAPIAddress, "transactions", query)

	response, e := c.get(transactionsQuery)
	if e != nil {
		return nil, e
	}

	bodyBytes, err := readResponseBody(response)

	var transactions *transaction.HederaTransactions
	err = json.Unmarshal(bodyBytes, &transactions)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func (c HederaMirrorClient) AccountExists(accountID hedera.AccountID) bool {
	mirrorNodeApiTransactionAddress := fmt.Sprintf("%s%s", c.mirrorAPIAddress, "accounts")
	accountQuery := fmt.Sprintf("%s/%s",
		mirrorNodeApiTransactionAddress,
		accountID.String())
	response, e := c.httpClient.Get(accountQuery)
	if e != nil {
		return false
	}

	if response.StatusCode != 200 {
		return false
	}

	return true
}

func (c HederaMirrorClient) GetUnprocessedMessagesAfterTimestamp(topicID hedera.TopicID, timestamp string) (*hcstopicmessage.HCSMessages, error) {
	mirrorNodeApiTopicAddress := fmt.Sprintf("%s%s", c.mirrorAPIAddress, "topics")
	unprocessedMessagesQuery := fmt.Sprintf("%s/%s/messages?timestamp=gt:%s",
		mirrorNodeApiTopicAddress,
		topicID.String(),
		timestamp,
	)
	response, e := c.httpClient.Get(unprocessedMessagesQuery)
	if e != nil {
		return nil, e
	}

	defer response.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(response.Body)

	var messages *hcstopicmessage.HCSMessages
	err := json.Unmarshal(bodyBytes, &messages)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func readResponseBody(response *http.Response) ([]byte, error) {
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
}
