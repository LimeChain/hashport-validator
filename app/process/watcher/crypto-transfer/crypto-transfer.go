package cryptotransfer

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
	"github.com/limechain/hedera-eth-bridge-validator/app/process"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/publisher"
	"github.com/limechain/hedera-eth-bridge-validator/config"
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
	logger           *log.Entry
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
		logger:           config.GetLoggerFor(fmt.Sprintf("Account [%s] Transfer Watcher", accountID.String())),
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

		ctw.logger.Warnf("[%s] Starting Timestamp was empty, proceeding to get [timestamp] from database.", accountAddress)
		milestoneTimestamp, err = ctw.statusRepository.GetLastFetchedTimestamp(accountAddress)
		if err == nil && milestoneTimestamp > 0 {
			return milestoneTimestamp
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			ctw.logger.Fatal(err)
		}

		ctw.logger.Warnf("[%s] Database Timestamp was empty, proceeding with [timestamp] from current moment.", accountAddress)
		milestoneTimestamp = time.Now().UnixNano()
		e := ctw.statusRepository.CreateTimestamp(accountAddress, milestoneTimestamp)
		if e != nil {
			ctw.logger.Fatal(e)
		}
		return milestoneTimestamp
	}

	milestoneTimestamp, err = ctw.statusRepository.GetLastFetchedTimestamp(accountAddress)
	if err != nil {
		ctw.logger.Warnf("[%s] Database Timestamp was empty. Restarting. Error - [%s]", accountAddress, err)
		ctw.started = false
		ctw.restart(q)
	}

	return milestoneTimestamp
}

func (ctw CryptoTransferWatcher) beginWatching(q *queue.Queue) {
	if !ctw.client.AccountExists(ctw.accountID) {
		ctw.logger.Errorf("Error incoming: Could not start monitoring account - Account not found.")
		return
	}
	ctw.logger.Info("Starting watcher")

	milestoneTimestamp := ctw.getTimestamp(q)
	if milestoneTimestamp == 0 {
		ctw.logger.Fatalf("Could not start watcher - Could not generate a milestone timestamp.")
	}

	ctw.logger.Infof("Started watcher")
	for {
		transactions, e := ctw.client.GetSuccessfulAccountCreditTransactionsAfterDate(ctw.accountID, milestoneTimestamp)
		if e != nil {
			ctw.logger.Errorf("Error incoming: Suddenly stopped monitoring account - [%s]", e)
			ctw.restart(q)
			return
		}

		if len(transactions.Transactions) > 0 {
			for _, tx := range transactions.Transactions {
				ctw.logger.Infof("[%s] - New transaction on account [%s] - Tx Hash: [%s]",
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
					ctw.logger.Errorf("Could not verify transaction memo - Error: [%s]", e)
					continue
				}

				// TODO: Should verify memo.
				ethAddress := decodedMemo[:42]
				feeString := string(decodedMemo[42:])

				re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
				if !re.MatchString(string(ethAddress)) {
					ctw.logger.Errorf("Could not verify Ethereum Address - [%s]", ethAddress)
					continue
				}

				_, e = helper.ToBigInt(feeString)
				if e != nil {
					ctw.logger.Errorf("Could not verify transaction fee - [%s]", feeString)
					continue
				}

				stateProof, e := ctw.client.GetStateProof(tx.TransactionID)
				if e != nil {
					ctw.logger.Errorf("Could not GET state proof, TransactionID [%s]. Error [%s]", tx.TransactionID, e)
					continue
				}

				verified, e := stateproof.Verify(tx.TransactionID, stateProof)
				if e != nil {
					ctw.logger.Errorf("Error while trying to verify state proof for TransactionID [%s]. Error [%s]", tx.TransactionID, e)
					continue
				}

				if !verified {
					ctw.logger.Errorf("Failed to verify state proof for TransactionID [%s]", tx.TransactionID)
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
				ctw.logger.Errorf("[%s]", ctw.accountID.String(), err)
				continue
			}
		}

		err := ctw.statusRepository.UpdateLastFetchedTimestamp(ctw.accountID.String(), milestoneTimestamp)
		if err != nil {
			ctw.logger.Errorf("Error incoming: Suddenly stopped monitoring account - [%s]", e)
			return
		}
		time.Sleep(ctw.pollingInterval * time.Second)
	}
}

func (ctw CryptoTransferWatcher) restart(q *queue.Queue) {
	if ctw.maxRetries > 0 {
		ctw.maxRetries--
		ctw.logger.Infof("Watcher is trying to reconnect")
		go ctw.Watch(q)
		return
	}
	ctw.logger.Errorf("Watcher failed: [Too many retries]")
}
