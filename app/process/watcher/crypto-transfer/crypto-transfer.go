package cryptotransfer

import (
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	cryptotransfermessage "github.com/limechain/hedera-eth-bridge-validator/app/process/model/crypto-transfer-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type CryptoTransferWatcher struct {
	client           *hederaClient.HederaClient
	accountID        hedera.AccountID
	typeMessage      string
	pollingInterval  time.Duration
	statusRepository repositories.StatusRepository
	maxRetries       int
}

func NewCryptoTransferWatcher(client *hederaClient.HederaClient, accountID hedera.AccountID, pollingInterval time.Duration, repository repositories.StatusRepository, maxRetries int) *CryptoTransferWatcher {
	return &CryptoTransferWatcher{
		client:           client,
		accountID:        accountID,
		typeMessage:      "HCS_CRYPTO_TRANSFER",
		pollingInterval:  pollingInterval,
		statusRepository: repository,
		maxRetries:       maxRetries,
	}
}

func (ctw CryptoTransferWatcher) Watch(q *queue.Queue) {
	go ctw.beginWatching(q)
}

func (ctw CryptoTransferWatcher) beginWatching(q *queue.Queue) {
	if !ctw.client.AccountExists(ctw.accountID) {
		log.Errorf("Error incoming: Could not start monitoring account [%s]\n", ctw.accountID.String())
		return
	}
	log.Infof("Starting Crypto Transfer Watcher for account [%s]\n", ctw.accountID)

	milestoneTimestamp, err := ctw.statusRepository.GetLastFetchedTimestamp(ctw.accountID)
	if err != nil {
		log.Errorf("Could not get last fetched timestamp for account [%s]\n", ctw.accountID.String())
		now := time.Now()
		milestoneTimestamp = strconv.FormatInt(now.Unix(), 10)
		log.Infof("Proceeding to monitor from current moment [%s]\n", now.String())
		e := ctw.statusRepository.CreateTimestamp(ctw.accountID, milestoneTimestamp)
		if e != nil {
			log.Errorf("Error incoming: Could not start monitoring account [%s]\n", ctw.accountID.String())
			log.Errorln(err)
			ctw.restart(q)
			return
		}
	}

	log.Infof("Started Crypto Transfer Watcher for account [%s]\n", ctw.accountID)
	for {
		transactions, e := ctw.client.GetAccountTransactionsAfterDate(ctw.accountID, milestoneTimestamp)
		if e != nil {
			log.Errorf("Error incoming: Suddenly stopped monitoring account [%s]\n", ctw.accountID.String())
			log.Errorln(e)
			ctw.restart(q)
			return
		}

		if len(transactions.Transactions) > 0 {
			for _, tx := range transactions.Transactions {
				log.Infof("[%s] - New transaction on account [%s] - Tx Hash: [%s]\n",
					tx.ConsensusTimestamp,
					ctw.accountID.String(),
					tx.TransactionHash)

				information := cryptotransfermessage.CryptoTransferMessage{
					TxMemo: tx.MemoBase64,
					Sender: tx.Transfers[len(tx.Transfers)-2].Account,
					Amount: tx.Transfers[len(tx.Transfers)-1].Amount,
				}
				publisher.Publish(information, ctw.typeMessage, ctw.accountID, q)
			}
			milestoneTimestamp = transactions.Transactions[len(transactions.Transactions)-1].ConsensusTimestamp
		}

		failure := ctw.statusRepository.UpdateLastFetchedTimestamp(ctw.accountID, milestoneTimestamp)
		if failure != nil {
			log.Errorf("Error incoming: Suddenly stopped monitoring account [%s]\n", ctw.accountID.String())
			log.Errorln(e)
			return
		}
		time.Sleep(ctw.pollingInterval * time.Second)
	}
}

func (ctw CryptoTransferWatcher) restart(q *queue.Queue) {
	if ctw.maxRetries > 0 {
		ctw.maxRetries--
		log.Printf("Crypto Transfer Watcher - Account [%s] - Trying to reconnect\n", ctw.accountID)
		go ctw.Watch(q)
		return
	}
	log.Errorf("Crypto Transfer Watcher - Account [%s] - Crypto Transfer Watcher failed: [Too many retries]\n", ctw.accountID)
}
