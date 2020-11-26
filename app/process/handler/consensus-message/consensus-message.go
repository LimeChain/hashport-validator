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
	"strconv"
	"strings"
	"time"
)

type ConsensusMessageHandler struct {
	repository       repositories.MessageRepository
	validAddresses   []string
	operatorAddress  string
	deadline         int64
	hederaNodeClient *hederaClient.HederaNodeClient
	topicID          hedera.ConsensusTopicID
}

func (cmh ConsensusMessageHandler) Recover(queue *queue.Queue) {
	log.Println("Recovery method not implemented yet.")
}

func NewConsensusMessageHandler(r repositories.MessageRepository, hederaNodeClient *hederaClient.HederaNodeClient, topicID hedera.ConsensusTopicID) *ConsensusMessageHandler {
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

	log.Printf("New Consensus Message for processing Transaction ID [%s] was received\n", m.TransactionId)

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
		Leader:               false,
		SignerAddress:        cmh.operatorAddress,
		TransactionTimestamp: m.TransactionTimestamp,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("Could not add Transaction Message with Transaction Id and Signature - [%s]-[%s]", m.TransactionId, m.Signature))
	}

	log.Printf("Successfully verified and persisted TX with ID [%s]\n", m.TransactionId)

	txSignatures, err := cmh.repository.GetByTransactionId(m.TransactionId, hexHash)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not retrieve Transaction Signatures for Transaction [%s]", m.TransactionId))
	}

	requiredSigCount := len(cmh.validAddresses)/2 + len(cmh.validAddresses)%2
	for len(txSignatures) < requiredSigCount {
		txSignatures, err = cmh.repository.GetByTransactionId(m.TransactionId, hexHash)
		if err != nil {
			return errors.New(fmt.Sprintf("Could not retrieve Transaction Signatures for Transaction [%s]", m.TransactionId))
		}
		time.Sleep(5 * time.Second)
	}

	pos := int64(cmh.findMyPosition(txSignatures))
	if pos == -1 {
		return errors.New(fmt.Sprintf("Operator is not amongst the potential leaders - [%v]", pos))
	}

	now := int64(time.Now().Minute())
	deadline := now + cmh.deadline*pos

	for now < deadline {
		// check if transaction was sent
		// if yes -> kill process

		now = int64(time.Now().Minute())
		time.Sleep(10 * time.Second)
	}

	// send transaction
	tasm := &validatorproto.TopicAggregatedSignaturesMessage{
		Hash:      hexHash,
		EthTxHash: "0x12345",
	}

	topicMessageBytes, err := proto.Marshal(tasm)
	if err != nil {
		log.Error("Could not marshal Topic Aggregated Signatures Message [%s] - [%s]", tasm, err)
	}

	_, err = cmh.hederaNodeClient.SubmitTopicConsensusMessage(cmh.topicID, topicMessageBytes)
	if err != nil {
		log.Error("Could not Submit Topic Consensus Message [%s] - [%s]", topicMessageBytes, err)
	}

	return nil
}

func (cmh ConsensusMessageHandler) listenForTx(messages []message.TransactionMessage) {
	leader, err := cmh.electLeader(messages)
	if err != nil {
		log.Error(err)
	}

	// Check if I am a Leader -> Send
	if cmh.operatorAddress == leader {
		log.Infof("Leader [%s]", cmh.operatorAddress)
		// I am leader!
		// Send Tx
		// Submit HCS Topic
		return
	}
	time.Sleep(5 * time.Second)

	// Check if Tx was sent
	// if yes -> return
	// if not -> listen
	cmh.listenForTx(messages)
}

func (cmh ConsensusMessageHandler) electLeader(messages []message.TransactionMessage) (string, error) {
	sort.Sort(message.ByTimestamp(messages))
	for _, m := range messages {
		txTimestamp, err := strconv.ParseInt(m.TransactionTimestamp, 10, 32)
		if err != nil {
			log.Error("Invalid Transaction Timestamp [%v] - [%s]", m.TransactionTimestamp, err)
		}
		deadline := txTimestamp + cmh.deadline

		if time.Now().Unix() > deadline {
			continue
		}
		return m.SignerAddress, nil
	}
	return "", errors.New("could not assign leader")
}

func (cmh ConsensusMessageHandler) findMyPosition(messages []message.TransactionMessage) int {
	sort.Sort(message.ByTimestamp(messages))

	for i := 0; i < len(messages); i++ {
		if messages[i].SignerAddress == cmh.operatorAddress {
			return i + 1
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
