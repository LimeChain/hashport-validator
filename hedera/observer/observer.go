package observer

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

var (
	stoppers = make(map[hederasdk.AccountID]chan bool)
)

func ObserveAccount(account hederasdk.AccountID, client *hederasdk.Client) {
	quitter := make(chan bool)
	go observe(account, client, quitter)
	stoppers[account] = quitter
}

func getTransactionsFor(account hederasdk.AccountID, lastProcessedTimestamp string) (*transaction.Transactions, error) {
	// TODO: GET LAST PROCESSED TX TIMESTAMP
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

func observe(account hederasdk.AccountID, client *hederasdk.Client, quitter <-chan bool) {
	lastObservedTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	for {
		select {
		case <-quitter:
			fmt.Printf("Stopped observing config [%s] successfully\n", account.String())
			return
		default:
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

					info := essential.Essential{
						TxMemo: transaction.MemoBase64,
						Sender: transaction.Transfers[len(transaction.Transfers)-2].Account,
						Amount: transaction.Transfers[len(transaction.Transfers)-1].Amount,
					}

					// TODO: Start Processing TX
					process(info)
				}
				lastObservedTimestamp = transactions.Transactions[len(transactions.Transactions)-1].ConsensusTimestamp
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func Stop(account hederasdk.AccountID) bool {
	if stoppers[account] == nil {
		fmt.Printf("[%s] is not being observed. Should throw error.\n", account.String())
		return false
	}

	stoppers[account] <- true
	fmt.Printf("Stopping observing config [%s]\n", account.String())
	return true
}

func process(essential.Essential) {
	// TODO: Implement processing logic
}
