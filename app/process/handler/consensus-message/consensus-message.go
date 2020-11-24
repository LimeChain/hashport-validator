package consensusmessage

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/services/signer/eth"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type ConsensusMessageHandler struct {
	repository      repositories.MessageRepository
	validAddresses  []string
	operatorAddress string
}

func NewConsensusMessageHandler(r repositories.MessageRepository) *ConsensusMessageHandler {
	return &ConsensusMessageHandler{
		repository:      r,
		validAddresses:  config.LoadConfig().Hedera.Handler.ConsensusMessage.Addresses,
		operatorAddress: eth.PrivateToPublicKeyToAddress(config.LoadConfig().Hedera.Client.Operator.EthPrivateKey).String(),
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

	decodedSig, err := hex.DecodeString(m.GetSignature())
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to decode signature. - [%s]", err))
	}

	hash := crypto.Keccak256([]byte(fmt.Sprintf("%s-%s-%d-%s", ctm.TransactionId, ctm.EthAddress, ctm.Amount, ctm.Fee)))
	messageHash := hex.EncodeToString(hash)

	key, err := crypto.Ecrecover(hash, decodedSig)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to recover public key. - [%s]", err))
	}

	pubKey, err := crypto.UnmarshalPubkey(key)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to unmarshal public key. - [%s]", err))
	}

	address := crypto.PubkeyToAddress(*pubKey)

	if !cmh.isValidAddress(address.String()) {
		return errors.New(fmt.Sprintf("Address is not valid - [%s]", address.String()))
	}

	messages, err := cmh.repository.GetByTxIdAndSignature(m.TransactionId, m.Signature)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to retrieve messages for TxId [%s], with signature [%s]. - [%s]", m.TransactionId, m.Signature, err))
	}

	if len(messages) > 0 {
		return errors.New(fmt.Sprintf("Duplicated Transaction Id and Signature - [%s]-[%s]", m.TransactionId, m.Signature))
	}

	err = cmh.repository.Create(&message.TransactionMessage{
		// Add Tx Timestamp to TransactionMessage
		TransactionId: m.TransactionId,
		EthAddress:    m.EthAddress,
		Amount:        m.Amount,
		Fee:           m.Fee,
		Signature:     m.Signature,
		Hash:          messageHash,
		Leader:        false,
		SignerAddress: cmh.operatorAddress,
	})
	if err != nil {
		return errors.New(fmt.Sprintf("Could not add Transaction Message with Transaction Id and Signature - [%s]-[%s]", m.TransactionId, m.Signature))
	}

	txSignatures, err := cmh.repository.GetByTransactionId(m.TransactionId, messageHash)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not retrieve Transaction Signatures for Transaction [%s]", m.TransactionId))
	}

	// Gather Signatures
	// Elect leader by Timestamp
	// Post-Signature-Gathering function

	// Elects a leader if there is no leader at the time
	if len(txSignatures) == 1 {
		err := cmh.repository.Elect(messageHash, m.Signature)
		if err != nil {
			return errors.New(fmt.Sprintf("Could not soft elect leader for Transaction [%s]", m.TransactionId))
		}
	}

	requiredSigCount := len(cmh.validAddresses)/2 + len(cmh.validAddresses)%2
	for len(txSignatures) < requiredSigCount {
		txSignatures, err = cmh.repository.GetByTransactionId(m.TransactionId, messageHash)
		if err != nil {
			return errors.New(fmt.Sprintf("Could not retrieve Transaction Signatures for Transaction [%s]", m.TransactionId))
		}
		time.Sleep(5 * time.Second)
	}

	// Check if I am a Leader -> Send
	if cmh.operatorAddress == txSignatures[0].SignerAddress {
		// Send Tx
		// Submit HCS Topic
	}

	// Start polling every amount of seconds
	for {
		// If transaction was sent -> return
		// If not -> wait
		// If a new leader is elected -> loop
		time.Sleep(5 * time.Second)
	}

	return nil
}

func (cmh ConsensusMessageHandler) isValidAddress(key string) bool {
	for _, k := range cmh.validAddresses {
		if strings.ToLower(k) == strings.ToLower(key) {
			return true
		}
	}
	return false
}
