package crypto_transfer

import (
	"Event-Listener/app/process/model/essential"
	"Event-Listener/app/process/model/transaction"
	"Event-Listener/config"
	"encoding/json"
	"fmt"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type CryptoTransferWatcher struct {
	Account hederasdk.AccountID
}

func (ctw CryptoTransferWatcher) Watch( /*TODO: add SDK queue as parameter*/ ) {
	go beginWatching(ctw.Account /*TODO: add SDK queue as parameter*/)
}

func getTransactionsFor(account hederasdk.AccountID, lastProcessedTimestamp string) (*transaction.Transactions, error) {
	// TODO: Get last processed Tx timestamp
	address := fmt.Sprintf("%s%s", config.MirrorNodeAPIAddress, "transactions")
	accountLink := fmt.Sprintf("%s?account.id=%s&type=credit&result=success&timestamp=gt:%s&order=asc",
		address,
		account.String(),
		lastProcessedTimestamp)

	response, e := http.Get(accountLink)
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

func beginWatching(account hederasdk.AccountID /*TODO: add SDK queue as parameter*/) {
	lastObservedTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	for {
		transactions, e := getTransactionsFor(account, lastObservedTimestamp)
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

				_ = essential.Essential{
					TxMemo: tx.MemoBase64,
					Sender: tx.Transfers[len(tx.Transfers)-2].Account,
					Amount: tx.Transfers[len(tx.Transfers)-1].Amount,
				}

				// TODO: Send TX for processing
				// TODO: push info object to SDK queue
			}
			lastObservedTimestamp = transactions.Transactions[len(transactions.Transactions)-1].ConsensusTimestamp
		}
		time.Sleep(5 * time.Second)
	}
}
