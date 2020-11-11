package crypto_transfer

import (
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	http "github.com/limechain/hedera-eth-bridge-validator/app/clients/http"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/crypto-transfer-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/proceed"
	"github.com/limechain/hedera-watcher-sdk/queue"
	"log"
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

func beginWatching(client *http.Client, account hederasdk.AccountID, typeMessage string, q *queue.Queue) {
	lastObservedTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	for {
		transactions, e := client.GetTransactionsByAccountIdAndTimestamp(account, lastObservedTimestamp)
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
