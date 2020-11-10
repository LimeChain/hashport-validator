package crypto_tranfer

import (
	"Event-Listener/hedera/config"
	"Event-Listener/hedera/model/essential"
	"Event-Listener/hedera/model/transaction"
	"encoding/json"
	"fmt"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	"io/ioutil"
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
			fmt.Printf("Suddenly stopped monitoring config [%s]\n", account.String())
			fmt.Println(e)
			return
		}

		if len(transactions.Transactions) > 0 {
			fmt.Printf("After [%s] - Account [%s] - Transactions Length: [%d]\n", lastObservedTimestamp, account.String(), len(transactions.Transactions))
			for _, transaction := range transactions.Transactions {
				fmt.Printf("[%s] - New transaction on config [%s] - Tx Hash: [%s]\n",
					transaction.ConsensusTimestamp,
					account.String(),
					transaction.TransactionHash)

				_ = essential.Essential{
					TxMemo: transaction.MemoBase64,
					Sender: transaction.Transfers[len(transaction.Transfers)-2].Account,
					Amount: transaction.Transfers[len(transaction.Transfers)-1].Amount,
				}

				// TODO: Send TX for processing
				// TODO: push info object to SDK queue
			}
			lastObservedTimestamp = transactions.Transactions[len(transactions.Transactions)-1].ConsensusTimestamp
		}
		time.Sleep(5 * time.Second)
	}
}
