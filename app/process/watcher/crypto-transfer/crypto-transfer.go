package crypto_transfer

import (
	"encoding/json"
	"fmt"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/crypto-transfer-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/proceed"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-watcher-sdk/queue"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type CryptoTransferWatcher struct {
	client      *http.Client
	accountID   hederasdk.AccountID
	typeMessage string
}

func (ctw CryptoTransferWatcher) Watch(queue *queue.Queue) {
	go beginWatching(ctw.client, ctw.accountID, ctw.typeMessage, queue)
}

func NewCryptoTransferWatcher(client *http.Client, accountID hederasdk.AccountID) *CryptoTransferWatcher {
	return &CryptoTransferWatcher{
		client:      client,
		accountID:   accountID,
		typeMessage: "HCS_CRYPTO_TRANSFER",
	}
}

func getTransactionsFor(client *http.Client, account hederasdk.AccountID, lastProcessedTimestamp string) (*transaction.Transactions, error) {
	// TODO: Get last processed Tx timestamp
	address := fmt.Sprintf("%s%s", config.LoadConfig().Hedera.MirrorNode.ApiAddress, "transactions")
	accountLink := fmt.Sprintf("%s?account.id=%s&type=credit&result=success&timestamp=gt:%s&order=asc",
		address,
		account.String(),
		lastProcessedTimestamp)

	response, e := client.Get(accountLink)
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

func beginWatching(client *http.Client, account hederasdk.AccountID, typeMessage string, q *queue.Queue) {
	lastObservedTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	for {
		transactions, e := getTransactionsFor(client, account, lastObservedTimestamp)
		if e != nil {
			log.Printf("Suddenly stopped monitoring config [%s]\n", account.String())
			log.Println(e)
			return
		}

		if len(transactions.Transactions) > 0 {
			log.Printf("After [%s] - Account [%s] - Transactions Length: [%d]\n", lastObservedTimestamp, account.String(), len(transactions.Transactions))
			for _, tx := range transactions.Transactions {
				log.Printf("[%s] - New transaction on config [%s] - Tx Hash: [%s]\n",
					tx.ConsensusTimestamp,
					account.String(),
					tx.TransactionHash)

				information := crypto_transfer_message.CryptoTransferMessage{
					TxMemo: tx.MemoBase64,
					Sender: tx.Transfers[len(tx.Transfers)-2].Account,
					Amount: tx.Transfers[len(tx.Transfers)-1].Amount,
				}
				proceed.Proceed(information, typeMessage, account, q)
			}
			lastObservedTimestamp = transactions.Transactions[len(transactions.Transactions)-1].ConsensusTimestamp
		}
		time.Sleep(5 * time.Second)
	}
}
