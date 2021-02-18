package recovery

import (
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	cryptotransfer "github.com/limechain/hedera-eth-bridge-validator/app/process/watcher/crypto-transfer"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	"time"
)

type RecoveryService struct {
	transactionRepository repositories.TransactionRepository
	statusRepository      repositories.StatusRepository
	mirrorClient          hedera.HederaMirrorClient
	nodeClient            hedera.HederaNodeClient
	accountID             hederasdk.AccountID
	topicID               hederasdk.TopicID
	configTimestamp       int64
}

func NewRecoveryService(
	transactionRepository repositories.TransactionRepository,
	statusRepository repositories.StatusRepository,
	mirrorClient hedera.HederaMirrorClient,
	nodeClient hedera.HederaNodeClient,
	accountID hederasdk.AccountID,
	topicID hederasdk.TopicID,
) *RecoveryService {
	return &RecoveryService{
		transactionRepository: transactionRepository,
		statusRepository:      statusRepository,
		mirrorClient:          mirrorClient,
		nodeClient:            nodeClient,
		accountID:             accountID,
		topicID:               topicID,
	}
}

func (rs *RecoveryService) Recover() error {
	now, err := rs.cryptoTransferRecovery()
	if err != nil {
		return err
	}

	_, err = rs.consensusMessageRecovery(now)
	if err != nil {
		return err
	}

	return nil
}

func (rs *RecoveryService) cryptoTransferRecovery() (int64, error) {
	now := time.Now().UnixNano()
	result, err := rs.mirrorClient.GetSuccessfulAccountCreditTransactionsAfterDate(rs.accountID, rs.getStartTimestamp(), now)
	if err != nil {
		// TODO: Log error properly
		return 0, err
	}

	for _, transaction := range result.Transactions {
		memoInfo, err := cryptotransfer.DecodeMemo(transaction.MemoBase64)
		if err != nil {
			// TODO: Log error properly
			return 0, err
		}

		err = rs.transactionRepository.Skip(&proto.CryptoTransferMessage{
			TransactionId: transaction.TransactionID,
			EthAddress:    memoInfo.EthAddress,
			Amount:        uint64(cryptotransfer.ExtractAmount(transaction, rs.accountID)),
			Fee:           memoInfo.FeeString,
		})

		if err != nil {
			// TODO: Log error properly
			return 0, err
		}
	}

	return now, nil
}

func (rs *RecoveryService) consensusMessageRecovery(now int64) (interface{}, error) {
	_, _ = hederasdk.NewTopicMessageQuery().
		SetStartTime(time.Unix(0, rs.getStartTimestamp())).
		SetEndTime(time.Unix(0, now)).
		SetTopicID(rs.topicID).
		Subscribe(
			rs.nodeClient.GetClient(),
			func(response hederasdk.TopicMessage) {
				// TODO: Process Topic Message properly
				// 1. Parse Message Type
				// 2. Query TX Status
				// 2.1 If Pending in Mempool -> ETH_TX_SUBMITTED
				// 2.2 If Reverted -> ETH_TX_REVERTED
				// 2.3 If Mined / Successful -> COMPLETED
				// 3. Return and start watchers / handlers from now (int64)
			},
		)

	return nil, nil
}

func (rs *RecoveryService) getStartTimestamp() int64 {
	timestamp, err := rs.statusRepository.GetLastFetchedTimestamp(rs.accountID.String())
	if err == nil {
		return timestamp
	}

	if rs.configTimestamp > 0 {
		return rs.configTimestamp
	}

	return time.Now().UnixNano()
}
