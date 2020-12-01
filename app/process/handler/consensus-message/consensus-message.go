package consensusmessage

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/hashgraph/hedera-sdk-go"
	hederaClient "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	ethhelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strings"
	"time"
)

type ConsensusMessageHandler struct {
	repository            repositories.MessageRepository
	operatorsEthAddresses []string
	operatorAddress       string
	deadline              int
	hederaNodeClient      *hederaClient.HederaNodeClient
	topicID               hedera.TopicID
}

func (cmh ConsensusMessageHandler) Recover(queue *queue.Queue) {
	log.Println("Recovery method not implemented yet.")
}

func NewConsensusMessageHandler(r repositories.MessageRepository, hederaNodeClient *hederaClient.HederaNodeClient) *ConsensusMessageHandler {
	topicID, err := hedera.TopicIDFromString(config.LoadConfig().Hedera.Handler.ConsensusMessage.TopicId)
	if err != nil {
		log.Fatal("Invalid topic id: [%v]", config.LoadConfig().Hedera.Handler.ConsensusMessage.TopicId)
	}

	return &ConsensusMessageHandler{
		repository:            r,
		operatorsEthAddresses: config.LoadConfig().Hedera.Handler.ConsensusMessage.Addresses,
		operatorAddress:       eth.PrivateToPublicKeyToAddress(config.LoadConfig().Hedera.Client.Operator.EthPrivateKey).String(),
		deadline:              config.LoadConfig().Hedera.Handler.ConsensusMessage.SendDeadline,
		hederaNodeClient:      hederaNodeClient,
		topicID:               topicID,
	}
}

func (cmh ConsensusMessageHandler) Handle(payload []byte) {
	go cmh.errorHandler(payload)
}

func (cmh ConsensusMessageHandler) errorHandler(payload []byte) {
	err := cmh.handlePayload(payload)
	if err != nil {
		log.Errorf("Error - could not handle payload: [%s]", err)
		return
	}
}

func (cmh ConsensusMessageHandler) handlePayload(payload []byte) error {
	m := &validatorproto.TopicSignatureMessage{}
	err := proto.Unmarshal(payload, m)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to unmarshal topic signature message. - [%s]", err))
	}

	ctm := &validatorproto.CryptoTransferMessage{
		TransactionId: m.TransactionId,
		EthAddress:    m.EthAddress,
		Amount:        m.Amount,
		Fee:           m.Fee,
	}

	log.Infof("New Consensus Message for processing Transaction ID [%s] was received\n", m.TransactionId)

	decodedSig, ethSig, err := ethhelper.DecodeSignature(m.GetSignature())
	if err != nil {
		return errors.New(fmt.Sprintf("[%s] - Failed to decode signature. - [%s]", m.TransactionId, err))
	}

	encodedData, err := ethhelper.EncodeData(ctm)
	if err != nil {
		log.Errorf("Failed to encode data for TransactionID [%s]. Error [%s].", ctm.TransactionId, err)
	}

	hash := crypto.Keccak256(encodedData)
	hexHash := hex.EncodeToString(hash)

	key, err := crypto.Ecrecover(hash, decodedSig)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s] - Failed to recover public key. Hash - [%s] - [%s]", m.TransactionId, hexHash, err))
	}

	pubKey, err := crypto.UnmarshalPubkey(key)
	if err != nil {
		return errors.New(fmt.Sprintf("[%s] - Failed to unmarshal public key. - [%s]", m.TransactionId, err))
	}

	address := crypto.PubkeyToAddress(*pubKey)

	if !cmh.isValidAddress(address.String()) {
		return errors.New(fmt.Sprintf("[%s] - Address is not valid - [%s]", m.TransactionId, address.String()))
	}

	mes, err := cmh.repository.GetTransaction(m.TransactionId, ethSig, hexHash)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New(fmt.Sprintf("Failed to retrieve messages for TxId [%s], with signature [%s]. - [%s]", m.TransactionId, ethSig, err))
	}

	if mes != nil || err == nil {
		return errors.New(fmt.Sprintf("Duplicated Transaction Id and Signature - [%s]-[%s]", m.TransactionId, ethSig))
	}

	err = cmh.repository.Create(&message.TransactionMessage{
		TransactionId:        m.TransactionId,
		EthAddress:           m.EthAddress,
		Amount:               m.Amount,
		Fee:                  m.Fee,
		Signature:            ethSig,
		Hash:                 hexHash,
		SignerAddress:        address.String(),
		TransactionTimestamp: m.TransactionTimestamp,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("Could not add Transaction Message with Transaction Id and Signature - [%s]-[%s] - [%s]", m.TransactionId, ethSig, err))
	}

	log.Infof("Successfully verified and saved signature for TX with ID [%s]", m.TransactionId)

	txSignatures, err := cmh.repository.GetByTransactionWith(m.TransactionId, hexHash)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not retrieve Transaction Signatures for Transaction [%s]", m.TransactionId))
	}

	if cmh.enoughSignaturesCollected(txSignatures, m.TransactionId) {
		log.Infof("Signatures for TX ID [%s] were collected. Proceeding with leader election.", m.TransactionId)
		go cmh.start(txSignatures, hexHash)
		log.Infof("Started leader election and preparation of aggregated signatures message.")
	}
	return nil
}

func (cmh ConsensusMessageHandler) start(txSignatures []message.TransactionMessage, hash string) {
	pos, err := cmh.findMyPosition(txSignatures)
	if err != nil {
		log.Errorf("Failed in finding leader position: [%s]", err)
		return
	}

	now := time.Now()
	deadline := now.Add(time.Second * time.Duration(cmh.deadline*pos))

	for now.Before(deadline) {
		if cmh.transactionSent() {
			return
		}
		now = time.Now()
		time.Sleep(10 * time.Second)
	}

	tasm := &validatorproto.TopicAggregatedSignaturesMessage{
		Hash:      hash,
		EthTxHash: "0x12345",
	}

	_, err = proto.Marshal(tasm) // topicMessageBytes
	if err != nil {
		log.Error("Could not marshal Topic Aggregated Signatures Message [%s] - [%s]", tasm, err)
	}

	//_, err = cmh.hederaNodeClient.SubmitTopicConsensusMessage(cmh.topicID, topicMessageBytes)
	//if err != nil {
	//	log.Error("Could not submit Topic Consensus Message [%s] - [%s]", topicMessageBytes, err)
	//}
}

func (cmh ConsensusMessageHandler) enoughSignaturesCollected(txSignatures []message.TransactionMessage, transactionId string) bool {
	requiredSigCount := len(cmh.operatorsEthAddresses)/2 + len(cmh.operatorsEthAddresses)%2
	log.Infof("Required signatures: [%v]", requiredSigCount)

	if len(txSignatures) < requiredSigCount {
		log.Infof("Insignificant amount of Transaction Signatures for Transaction [%s] - [%d] signatures", transactionId, len(txSignatures))
		return false
	}
	return true
}

func (cmh ConsensusMessageHandler) transactionSent() bool {
	// TODO: Implement Transaction status check
	return false
}

func (cmh ConsensusMessageHandler) findMyPosition(messages []message.TransactionMessage) (int, error) {
	for i := 0; i < len(messages); i++ {
		if messages[i].SignerAddress == cmh.operatorAddress {
			return i, nil
		}
	}

	return -1, errors.New(fmt.Sprintf("Operator is not amongst the potential leaders - [%v]", cmh.operatorAddress))
}

func (cmh ConsensusMessageHandler) isValidAddress(key string) bool {
	for _, k := range cmh.operatorsEthAddresses {
		if strings.ToLower(k) == strings.ToLower(key) {
			return true
		}
	}
	return false
}
