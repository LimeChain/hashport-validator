package hedera

import (
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

type HederaNodeClient struct {
	client *hedera.Client
}

func NewNodeClient(config config.Client) *HederaNodeClient {
	var client *hedera.Client
	switch config.NetworkType {
	case "mainnet":
		client = hedera.ClientForMainnet()
	case "testnet":
		client = hedera.ClientForTestnet()
	case "previewnet":
		client = hedera.ClientForPreviewnet()
	default:
		log.Fatalf("Invalid Client NetworkType provided: [%s]", config.NetworkType)
	}

	accID, err := hedera.AccountIDFromString(config.Operator.AccountId)
	if err != nil {
		log.Fatalf("Invalid Operator AccountId provided: [%s]", config.Operator.AccountId)
	}

	privateKey, err := hedera.PrivateKeyFromString(config.Operator.PrivateKey)
	if err != nil {
		log.Fatalf("Invalid Operator PrivateKey provided: [%s]", config.Operator.PrivateKey)
	}

	client.SetOperator(accID, privateKey)

	return &HederaNodeClient{client}
}

func (hc *HederaNodeClient) GetClient() *hedera.Client {
	return hc.client
}

func (hc *HederaNodeClient) SubmitTopicConsensusMessage(topicId hedera.TopicID, message []byte) (*hedera.TransactionID, error) {
	id, err := hedera.NewTopicMessageSubmitTransaction().
		SetTopicID(topicId).
		SetMessage(message).
		Execute(hc.client)

	if err != nil {
		return nil, err
	}

	receipt, err := id.GetReceipt(hc.client)
	if err != nil {
		return nil, err
	}

	if receipt.Status != hedera.StatusSuccess {
		return nil, errors.New(fmt.Sprintf("Transaction [%s] failed with status [%s]", id.TransactionID.String(), receipt.Status))
	}

	return &id.TransactionID, err
}
