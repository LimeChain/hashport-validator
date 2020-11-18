package hedera

import (
	"encoding/json"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	hcstopicmessage "github.com/limechain/hedera-eth-bridge-validator/app/process/model/hcs-topic-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type HederaMirrorClient struct {
	mirrorAPIAddress string
	mirrorClient     *hedera.MirrorClient
	httpClient       *http.Client
}

func NewHederaMirrorClient(mirrorNodeAPIAddress, mirrorNodeClientAddress string) *HederaMirrorClient {
	mirrorClient, e := hedera.NewMirrorClient(mirrorNodeClientAddress)
	if e != nil {
		log.Fatal(e)
	}

	return &HederaMirrorClient{
		mirrorAPIAddress: mirrorNodeAPIAddress,
		mirrorClient:     &mirrorClient,
		httpClient:       &http.Client{},
	}
}

func (c HederaMirrorClient) GetMirrorClient() *hedera.MirrorClient {
	return c.mirrorClient
}

func (c HederaMirrorClient) GetAccountTransactionsAfterDate(accountId hedera.AccountID, milestoneTimestamp string) (*transaction.HederaTransactions, error) {
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
	err := json.Unmarshal(bodyBytes, &transactions)
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

func (c HederaMirrorClient) GetUnprocessedMessagesAfterTimestamp(topicID hedera.ConsensusTopicID, timestamp string) (*hcstopicmessage.HCSMessages, error) {
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
