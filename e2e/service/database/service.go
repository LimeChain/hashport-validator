package database

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repository"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
	entity_transfer "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/transfer"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/e2e/model"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	transactions repository.Transfer
	messages     repository.Message
	logger       *log.Entry
}

func NewService(dbConfig config.Db) *Service {
	db := persistence.RunDb(dbConfig)
	return &Service{
		transactions: transfer.NewRepository(db),
		messages:     message.NewRepository(db),
		logger:       config.GetLoggerFor("DB Validation Service"),
	}
}

func (s *Service) TransactionRecordExists(transactionID, receiverAddress, nativeToken, wrappedToken, amount, txReimbursement, gasCost, ethTxHash string, executeEthTransaction bool) (bool, *entity.Transfer, error) {
	bigGasPriceGwei, err := helper.ToBigInt(gasCost)
	if err != nil {
		s.logger.Fatalf("Could not parse GasPriceGwei [%s] to Big Integer for TX ID [%s]. Error: [%s].", gasCost, transactionID, err)
	}
	bigGasPriceWei := ethereum.GweiToWei(bigGasPriceGwei).String()

	expectedTransferRecord := &entity.Transfer{
		TransactionID:      transactionID,
		Receiver:           receiverAddress,
		NativeToken:        nativeToken,
		WrappedToken:       wrappedToken,
		Amount:             amount,
		TxReimbursement:    txReimbursement,
		GasPrice:           bigGasPriceWei,
		Status:             entity_transfer.StatusCompleted,
		SignatureMsgStatus: entity_transfer.StatusSignatureMined,
		//EthTxMsgStatus:        entity_transfer.StatusEthTxMsgMined, // TODO: Uncomment when ready
		EthTxStatus:           entity_transfer.StatusEthTxMined,
		EthTxHash:             ethTxHash,
		ExecuteEthTransaction: true,
	}

	actualDbTx, err := s.transactions.GetByTransactionId(expectedTransferRecord.TransactionID)
	if err != nil {
		return false, nil, err
	}
	return expectedTransferRecord.Equals(*actualDbTx), expectedTransferRecord, nil
}

func (s *Service) SignatureMessagesExist(record *entity.Transfer, signatureDuplets []model.SigDuplet) (bool, []entity.Message, error) {
	var expectedMessageRecords []entity.Message

	for _, duplet := range signatureDuplets {
		tm := entity.Message{
			TransferID: record.TransactionID,
			Transfer:   *record,
			Hash:       record.EthTxHash,
			Signature:  duplet.Signature,
			//Signer:               duplet.Signer, // TODO: Decode Signature, EncodeMessageBytes, VerifySignature to get Signer
			TransactionTimestamp: duplet.ConsensusTimestamp,
		}
		expectedMessageRecords = append(expectedMessageRecords, tm)
	}

	messages, err := s.messages.Get(record.TransactionID)
	if err != nil {
		return false, nil, err
	}

	for _, m := range expectedMessageRecords {
		if !contains(m, messages) {
			return false, expectedMessageRecords, nil
		}
	}
	return len(messages) == len(expectedMessageRecords), expectedMessageRecords, nil
}

func contains(m entity.Message, array []entity.Message) bool {
	for _, a := range array {
		if a.Equals(m) {
			return true
		}
	}
	return false
}
