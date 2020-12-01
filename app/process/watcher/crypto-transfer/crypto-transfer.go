package cryptotransfer

import (
	"encoding/base64"
	"errors"
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/process"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	protomsg "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-state-proof-verifier-go/stateproof"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"regexp"
	"time"
)

type CryptoTransferWatcher struct {
	client           *hederaClient.HederaMirrorClient
	accountID        hedera.AccountID
	typeMessage      string
	pollingInterval  time.Duration
	statusRepository repositories.StatusRepository
	maxRetries       int
	startTimestamp   int64
	started          bool
}

func NewCryptoTransferWatcher(
	client *hederaClient.HederaMirrorClient,
	accountID hedera.AccountID,
	pollingInterval time.Duration,
	repository repositories.StatusRepository,
	maxRetries int,
	startTimestamp int64,
) *CryptoTransferWatcher {
	return &CryptoTransferWatcher{
		client:           client,
		accountID:        accountID,
		typeMessage:      process.CryptoTransferMessageType,
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

func (ctw CryptoTransferWatcher) getTimestamp(q *queue.Queue) int64 {
	accountAddress := ctw.accountID.String()
	milestoneTimestamp := ctw.startTimestamp
	var err error

	if !ctw.started {
		if milestoneTimestamp > 0 {
			return milestoneTimestamp
		}

		log.Warnf("[%s] Starting Timestamp was empty, proceeding to get [timestamp] from database.", accountAddress)
		milestoneTimestamp, err = ctw.statusRepository.GetLastFetchedTimestamp(accountAddress)
		if err == nil && milestoneTimestamp > 0 {
			return milestoneTimestamp
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Fatal(err)
		}

		log.Warnf("[%s] Database Timestamp was empty, proceeding with [timestamp] from current moment.", accountAddress)
		milestoneTimestamp = time.Now().UnixNano()
		e := ctw.statusRepository.CreateTimestamp(accountAddress, milestoneTimestamp)
		if e != nil {
			log.Fatal(e)
		}
		return milestoneTimestamp
	}

	milestoneTimestamp, err = ctw.statusRepository.GetLastFetchedTimestamp(accountAddress)
	if err != nil {
		log.Warnf("[%s] Database Timestamp was empty. Restarting. Error - [%s]", accountAddress, err)
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
	if milestoneTimestamp == 0 {
		log.Fatalf("Could not start Crypto Transfer Watcher for account [%s] - Could not generate a milestone timestamp.\n", ctw.accountID)
	}

	log.Infof("Started Crypto Transfer Watcher for account [%s]\n", ctw.accountID)
	for {
		transactions, e := ctw.client.GetSuccessfulAccountCreditTransactionsAfterDate(ctw.accountID, milestoneTimestamp)
		if e != nil {
			log.Errorf("Error incoming: Suddenly stopped monitoring account [%s] - [%s]", ctw.accountID.String(), e)
			ctw.restart(q)
			return
		}

		if len(transactions.Transactions) > 0 {
			for _, tx := range transactions.Transactions {
				log.Infof("[%s] - New transaction on account [%s] - Tx Hash: [%s]",
					tx.ConsensusTimestamp,
					ctw.accountID.String(),
					tx.TransactionHash)

				var amount int64
				for _, tr := range tx.Transfers {
					if tr.Account == ctw.accountID.String() {
						amount = tr.Amount
					}
				}

				decodedMemo, e := base64.StdEncoding.DecodeString(tx.MemoBase64)
				if e != nil || len(decodedMemo) < 42 {
					log.Errorf("[%s] Crypto Transfer Watcher: Could not verify transaction memo - Error: [%s]\n", ctw.accountID.String(), e)
					continue
				}

				// TODO: Should verify memo.
				ethAddress := decodedMemo[:42]
				feeString := string(decodedMemo[42:])

				re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
				if !re.MatchString(string(ethAddress)) {
					log.Errorf("[%s] Crypto Transfer Watcher: Could not verify Ethereum Address\n\t- [%s]", ctw.accountID.String(), ethAddress)
					continue
				}

				_, e = helper.ToBigInt(feeString)
				if e != nil {
					log.Errorf("[%s] Crypto Transfer Watcher: Could not verify transaction fee\n\t- [%s]", ctw.accountID.String(), feeString)
					continue
				}

				stateProof, e := ctw.client.GetStateProof(tx.TransactionID)
				if e != nil {
					log.Errorf("[%s] Crypto Transfer Watcher: Could not GET state proof, TransactionID [%s]. Error [%s]", ctw.accountID.String(), tx.TransactionID, e)
					continue
				}

				verified, e := stateproof.Verify(tx.TransactionID, stateProof)
				if e != nil {
					log.Errorf("[%s] Crypto Transfer Watcher: Error while trying to verify state proof for TransactionID [%s]. Error [%s]", ctw.accountID.String(), tx.TransactionID, e)
					continue
				}

				if !verified {
					log.Errorf("[%s] Crypto Transfer Watcher: Failed to verify state proof for TransactionID [%s]", ctw.accountID.String(), tx.TransactionID)
					continue
				}

				information := &protomsg.CryptoTransferMessage{
					TransactionId: tx.TransactionID,
					EthAddress:    string(ethAddress),
					Amount:        uint64(amount),
					Fee:           feeString,
				}
				publisher.Publish(information, ctw.typeMessage, ctw.accountID, q)
			}
			var err error
			milestoneTimestamp, err = timestamp.FromString(transactions.Transactions[len(transactions.Transactions)-1].ConsensusTimestamp)
			if err != nil {
				log.Errorf("[%s] Crypto Transfer Watcher: [%s]", ctw.accountID.String(), err)
				continue
			}
		}

		err := ctw.statusRepository.UpdateLastFetchedTimestamp(ctw.accountID.String(), milestoneTimestamp)
		if err != nil {
			log.Errorf("Error incoming: Suddenly stopped monitoring account [%s] - [%s]", ctw.accountID.String(), e)
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
