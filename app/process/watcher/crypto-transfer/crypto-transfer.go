package cryptotransfer

import (
	"encoding/base64"
	"errors"
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	cryptotransfermessage "github.com/limechain/hedera-eth-bridge-validator/app/process/model/crypto-transfer-message"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"regexp"
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
	startTimestamp   string
	started          bool
}

func NewCryptoTransferWatcher(client *hederaClient.HederaClient, accountID hedera.AccountID, pollingInterval time.Duration, repository repositories.StatusRepository, maxRetries int, startTimestamp string) *CryptoTransferWatcher {
	return &CryptoTransferWatcher{
		client:           client,
		accountID:        accountID,
		typeMessage:      "HCS_CRYPTO_TRANSFER",
		pollingInterval:  pollingInterval,
		statusRepository: repository,
		maxRetries:       maxRetries,
		startTimestamp:   startTimestamp,
		started:          false,
	}
}

func (ctw CryptoTransferWatcher) Watch(q *queue.Queue) {
	go ctw.beginWatching(q)
}

func (ctw CryptoTransferWatcher) getTimestamp(q *queue.Queue) string {
	accountAddress := ctw.accountID.String()
	milestoneTimestamp := ctw.startTimestamp
	var err error

	if !ctw.started {
		if milestoneTimestamp != "" {
			return milestoneTimestamp
		}

		log.Warnf("[%s] Starting Timestamp was empty, proceeding to get [timestamp] from database.\n", accountAddress)
		milestoneTimestamp, err := ctw.statusRepository.GetLastFetchedTimestamp(accountAddress)
		if milestoneTimestamp != "" {
			return milestoneTimestamp
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Fatal(err)
		}

		log.Warnf("[%s] Database Timestamp was empty, proceeding with [timestamp] from current moment.\n", accountAddress)
		milestoneTimestamp = strconv.FormatInt(time.Now().Unix(), 10)
		e := ctw.statusRepository.CreateTimestamp(accountAddress, milestoneTimestamp)
		if e != nil {
			log.Fatal(e)
		}
		return milestoneTimestamp
	}

	milestoneTimestamp, err = ctw.statusRepository.GetLastFetchedTimestamp(accountAddress)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Warnf("[%s] Database Timestamp was empty. Restarting.\n", accountAddress)
		ctw.started = false
		ctw.restart(q)
	}

	return milestoneTimestamp
}

func (ctw CryptoTransferWatcher) beginWatching(q *queue.Queue) {
	if !ctw.client.AccountExists(ctw.accountID) {
		log.Errorf("Error incoming: Could not start monitoring account [%s] - Account not found.\n", ctw.accountID.String())
		return
	}
	log.Infof("Starting Crypto Transfer Watcher for account [%s]\n", ctw.accountID)

	milestoneTimestamp := ctw.getTimestamp(q)
	if milestoneTimestamp == "" {
		log.Fatalf("Could not start Crypto Transfer Watcher for account [%s] - Could not generate a milestone timestamp.\n", ctw.accountID)
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

				var sender string
				var amount int64
				for _, tr := range tx.Transfers {
					if tr.Amount < 0 {
						sender = tr.Account
					} else if tr.Account == ctw.accountID.String() {
						amount = tr.Amount
					}
				}

				decodedMemo, e := base64.StdEncoding.DecodeString(tx.MemoBase64)
				if e != nil || len(decodedMemo) < 20 {
					log.Errorf("[%s] Crypto Transfer Watcher: Could not verify transaction memo - Error: [%s]\n", ctw.accountID.String(), e)
					continue
				}

				// TODO: Should verify memo.
				ethAddress := string(decodedMemo[:20])
				feeString := string(decodedMemo[20:])

				re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
				if !re.MatchString(string(ethAddress)) {
					log.Errorf("[%s] Crypto Transfer Watcher: Could not verify Ethereum Address\n\t- [%s]", ctw.accountID.String(), ethAddress)
					continue
				}

				fee, e := strconv.ParseInt(feeString, 10, 64)
				if e != nil {
					log.Errorf("[%s] Crypto Transfer Watcher: Could not verify transaction fee\n\t- [%s]", ctw.accountID.String(), feeString)
					continue
				}

				information := cryptotransfermessage.CryptoTransferMessage{
					EthAddress: ethAddress,
					TxId:       tx.TransactionID,
					TxFee:      uint64(fee),
					Sender:     sender,
					Amount:     amount,
				}
				publisher.Publish(information, ctw.typeMessage, ctw.accountID, q)
			}
			milestoneTimestamp = transactions.Transactions[len(transactions.Transactions)-1].ConsensusTimestamp
		}

		err := ctw.statusRepository.UpdateLastFetchedTimestamp(ctw.accountID.String(), milestoneTimestamp)
		if err != nil {
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
		log.Infof("Crypto Transfer Watcher - Account [%s] - Trying to reconnect\n", ctw.accountID)
		go ctw.Watch(q)
		return
	}
	log.Errorf("Crypto Transfer Watcher - Account [%s] - Crypto Transfer Watcher failed: [Too many retries]\n", ctw.accountID)
}
