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
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	"github.com/limechain/hedera-watcher-sdk/queue"
	log "github.com/sirupsen/logrus"
	"sort"
	"strings"
	"time"
)

type ConsensusMessageHandler struct {
	repository       repositories.MessageRepository
	validAddresses   []string
	operatorAddress  string
	deadline         int
	hederaNodeClient *hederaClient.HederaNodeClient
	topicID          hedera.ConsensusTopicID
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
		repository:       r,
		validAddresses:   config.LoadConfig().Hedera.Handler.ConsensusMessage.Addresses,
		operatorAddress:  eth.PrivateToPublicKeyToAddress(config.LoadConfig().Hedera.Client.Operator.EthPrivateKey).String(),
		deadline:         config.LoadConfig().Hedera.Handler.ConsensusMessage.SendDeadline,
		hederaNodeClient: hederaNodeClient,
		topicID:          topicID,
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

	decodedSig, err := hex.DecodeString(m.GetSignature())
	if err != nil {
		return errors.New(fmt.Sprintf("[%s] - Failed to decode signature. - [%s]", m.TransactionId, err))
	}

	hash := crypto.Keccak256([]byte(fmt.Sprintf("%s-%s-%d-%s", ctm.TransactionId, ctm.EthAddress, ctm.Amount, ctm.Fee)))
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

	messages, err := cmh.repository.GetTransaction(m.TransactionId, m.Signature, hexHash)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to retrieve messages for TxId [%s], with signature [%s]. - [%s]", m.TransactionId, m.Signature, err))
	}

	if len(messages) > 0 {
		return errors.New(fmt.Sprintf("Duplicated Transaction Id and Signature - [%s]-[%s]", m.TransactionId, m.Signature))
	}

	err = cmh.repository.Create(&message.TransactionMessage{
		TransactionId:        m.TransactionId,
		EthAddress:           m.EthAddress,
		Amount:               m.Amount,
		Fee:                  m.Fee,
		Signature:            m.Signature,
		Hash:                 hexHash,
		SignerAddress:        address.String(),
		TransactionTimestamp: m.TransactionTimestamp,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("Could not add Transaction Message with Transaction Id and Signature - [%s]-[%s]", m.TransactionId, m.Signature))
	}

	log.Infof("Successfully verified and persisted TX with ID [%s]", m.TransactionId)

	txSignatures, err := cmh.repository.GetByTransactionId(m.TransactionId, hexHash)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not retrieve Transaction Signatures for Transaction [%s]", m.TransactionId))
	}

	requiredSigCount := len(cmh.validAddresses)/2 + len(cmh.validAddresses)%2
	log.Infof("Required signatures: [%v]", requiredSigCount)

	txSignatures, err = cmh.repository.GetByTransactionId(m.TransactionId, hexHash)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not retrieve Transaction Signatures for Transaction [%s]", m.TransactionId))
	}

	if len(txSignatures) < requiredSigCount {
		log.Warnf(fmt.Sprintf("Insignificant amount of Transaction Signatures for Transaction [%s] - [%d] signatures", m.TransactionId, len(txSignatures)))
		return nil
	}

	pos := cmh.findMyPosition(txSignatures)
	if pos == -1 {
		return errors.New(fmt.Sprintf("Operator is not amongst the potential leaders - [%v]", pos))
	}

	log.Infof("My position: [%v]", pos)

	now := time.Now()
	deadline := now.Add(time.Minute * time.Duration(cmh.deadline*pos))

	log.Infof("NOW: [%v]", now)
	log.Infof("DEADLINE: [%v]", deadline)

	for now.Before(deadline) {
		if cmh.transactionSent() {
			return nil
		}
		now = time.Now()
		time.Sleep(10 * time.Second)
	}
	log.Infof("I should submit the transaction now!")

	tasm := &validatorproto.TopicAggregatedSignaturesMessage{
		Hash:      hexHash,
		EthTxHash: "0x12345",
	}

	_, err = proto.Marshal(tasm) // topicMessageBytes
	if err != nil {
		log.Error("Could not marshal Topic Aggregated Signatures Message [%s] - [%s]", tasm, err)
	}

	log.Infof("I should send the message! - [%v]", tasm)
	//_, err = cmh.hederaNodeClient.SubmitTopicConsensusMessage(cmh.topicID, topicMessageBytes)
	//if err != nil {
	//	log.Error("Could not Submit Topic Consensus Message [%s] - [%s]", topicMessageBytes, err)
	//}

	return nil
}

func (cmh ConsensusMessageHandler) transactionSent() bool {
	// TODO: Implement Transaction status check
	return false
}

func (cmh ConsensusMessageHandler) findMyPosition(messages []message.TransactionMessage) int {
	sort.Sort(message.ByTimestamp(messages))
	for i := 0; i < len(messages); i++ {
		if messages[i].SignerAddress == cmh.operatorAddress {
			return i
		}
	}
	return -1
}

func (cmh ConsensusMessageHandler) isValidAddress(key string) bool {
	for _, k := range cmh.validAddresses {
		if strings.ToLower(k) == strings.ToLower(key) {
			return true
		}
	}
	return false
}
