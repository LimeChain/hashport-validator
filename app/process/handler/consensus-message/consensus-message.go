package consensusmessage

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/repositories"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	validatorproto "github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
)

type ConsensusMessageHandler struct {
	repository repositories.MessageRepository
}

func NewConsensusMessageHandler(repository message.MessageRepository) *ConsensusMessageHandler {
	return &ConsensusMessageHandler{
		repository: repository,
	}
}

func (cmh ConsensusMessageHandler) Handle(payload []byte) {
	err := cmh.handlePayload(payload)
	if err != nil {
		log.Fatalf("Error - could not handle payload: [%s]", err)
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

	log.Printf("[%s] signature", m.GetSignature())

	decodedSig, err := hex.DecodeString(m.GetSignature())
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to decode signature. - [%s]", err))
	}

	hash := crypto.Keccak256([]byte(fmt.Sprintf("%s-%s-%d-%s", ctm.TransactionId, ctm.EthAddress, ctm.Amount, ctm.Fee)))

	log.Printf("[%s] hex hash", hex.EncodeToString(hash))

	key, err := crypto.Ecrecover(hash, decodedSig)
	pubKey, err := crypto.UnmarshalPubkey(key)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to unmarshal public key. - [%s]", err))
	}

	address := crypto.PubkeyToAddress(*pubKey)

	if !isValidAddress(address.String()) {
		return errors.New(fmt.Sprintf("Address is not valid - [%s]", address.String()))
	}

	messages, err := cmh.repository.Get(m.TransactionId, m.Signature)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to retrieve messages for TxId [%s], with signature [%s]. - [%s]", m.TransactionId, m.Signature, err))
	}

	if len(messages) > 0 {
		return errors.New(fmt.Sprintf("Duplicated Transaction Id and Signature - [%s]-[%s]", m.TransactionId, m.Signature))
	}

	err = cmh.repository.Add(&message.TransactionMessage{
		TransactionId: m.TransactionId,
		EthAddress:    m.EthAddress,
		Amount:        m.Amount,
		Fee:           m.Fee,
		Signature:     m.Signature,
		Hash:          hex.EncodeToString(hash),
	})
	if err != nil {
		return errors.New(fmt.Sprintf("Could not add Transaction Message with Transaction Id and Signature - [%s]-[%s]", m.TransactionId, m.Signature))
	}

	fmt.Println("Success.")
	return nil
}

func isValidAddress(key string) bool {
	keys := config.LoadConfig().Hedera.Handler.ConsensusMessage.Addresses
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}
