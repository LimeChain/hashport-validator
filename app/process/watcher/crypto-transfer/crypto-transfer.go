package cryptotransfer

import (
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	cryptotransfermessage "github.com/limechain/hedera-eth-bridge-validator/app/process/model/crypto-transfer-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"time"
)

type CryptoTransferWatcher struct {
	client           *hederaClient.HederaClient
	accountID        hedera.AccountID
	typeMessage      string
	pollingInterval  time.Duration
	statusRepository repositories.StatusRepository
}

func NewCryptoTransferWatcher(client *hederaClient.HederaClient, accountID hedera.AccountID, pollingInterval time.Duration, repository repositories.StatusRepository) *CryptoTransferWatcher {
	return &CryptoTransferWatcher{
		client:           client,
		accountID:        accountID,
		typeMessage:      "HCS_CRYPTO_TRANSFER",
		pollingInterval:  pollingInterval,
		statusRepository: repository,
	}
}

func (ctw CryptoTransferWatcher) Watch(queue *queue.Queue) {
	go ctw.beginWatching(ctw.accountID, ctw.typeMessage, queue)
}

func (ctw CryptoTransferWatcher) beginWatching(accountID hedera.AccountID, typeMessage string, q *queue.Queue) {
	milestoneTimestamp := ctw.statusRepository.GetLastFetchedTimestamp(accountID)
	for {
		transactions, e := ctw.client.GetAccountTransactionsAfterDate(accountID, milestoneTimestamp)
		if e != nil {
			log.Errorf("Suddenly stopped monitoring account [%s]\n", accountID.String())
			log.Errorln(e)
			return
		}

		if len(transactions.Transactions) > 0 {
			log.Infof("After [%s] - Account [%s] - Transactions Length: [%d]\n", milestoneTimestamp, accountID.String(), len(transactions.Transactions))
			for _, tx := range transactions.Transactions {
				log.Infof("[%s] - New transaction on account [%s] - Tx Hash: [%s]\n",
					tx.ConsensusTimestamp,
					accountID.String(),
					tx.TransactionHash)

				information := cryptotransfermessage.CryptoTransferMessage{
					TxMemo: tx.MemoBase64,
					Sender: tx.Transfers[len(tx.Transfers)-2].Account,
					Amount: tx.Transfers[len(tx.Transfers)-1].Amount,
				}
				publisher.Publish(information, typeMessage, accountID, q)
			}
			milestoneTimestamp = transactions.Transactions[len(transactions.Transactions)-1].ConsensusTimestamp
		}

		failure := ctw.statusRepository.UpdateLastFetchedTimestamp(accountID, milestoneTimestamp)
		if failure != nil {
			log.Errorf("Suddenly stopped monitoring account [%s]\n", accountID.String())
			log.Errorln(e)
			return
		}
		time.Sleep(ctw.pollingInterval * time.Second)
	}
}
