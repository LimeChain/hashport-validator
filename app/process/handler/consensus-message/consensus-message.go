package consensusmessage

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/message"
	hcstopicmessage "github.com/limechain/hedera-eth-bridge-validator/app/process/model/hcs-topic-message"
	log "github.com/sirupsen/logrus"
)

// TODO: Consensus message event handler

type ConsensusMessageHandler struct {
	repository message.MessageRepository
}

func NewConsensusMessageHandler(repository message.MessageRepository) *ConsensusMessageHandler {
	return &ConsensusMessageHandler{
		repository: repository,
	}
}

func (cmh ConsensusMessageHandler) Handle(payload []byte) error {
	// receive new message
	// proto unmarshall
	// verify authentication message
	// store message into db with a TX Id

	m := &hcstopicmessage.ConsensusMessage{
		TransactionID: "sometxid",
		EthAddress:    "someethaddress",
		Amount:        123,
		Fee:           "123321",
		Signature:     "soemsignaturee",
	}
	key, err := crypto.Ecrecover()

	messages, err := cmh.repository.Get(m.TransactionID)
	if err != nil {
		log.Errorf("Error - [%s]", err)
		return err
	}

	err = cmh.repository.Add(m)
	if err != nil {
		log.Errorf("Error - [%s]", err)
		return err
	}

	log.Printf("Log: [%s]\n", messages)

	//err = json.Unmarshal(payload, message)
	//if err != nil {
	//	log.Errorf("Error - [%s]", err)
	//	return err
	//}
	//
	//log.Println(message)
	return nil
}
