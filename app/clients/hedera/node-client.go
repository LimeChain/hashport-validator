package hedera

import (
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"log"
)

type HederaNodeClient struct {
	client *hedera.Client
}

func (hc *HederaNodeClient) SubmitTopicConsensusMessage(topicId hedera.ConsensusTopicID, message []byte) (string, error) {
	id, err := hedera.NewConsensusMessageSubmitTransaction().
		SetTopicID(topicId).
		SetMessage(message).
		Execute(hc.client)

	if err != nil {
		return "", err
	}

	receipt, err := id.GetReceipt(hc.client)
	if err != nil {
		return "", err
	}

	if receipt.Status != hedera.StatusOk {
		// TODO: what happens if the tx fails to be submitted?
		return "", errors.New(fmt.Sprintf("Transaction [%s] failed with status [%s]", id.String(), receipt.Status))
	}

	return id.String(), err
}

func NewClient(config config.Client) *HederaNodeClient {
	var client *hedera.Client
	switch config.NetworkType {
	case "mainnet":
		client = hedera.ClientForMainnet()
	case "testnet":
		client = hedera.ClientForTestnet()
	case "previewnet":
		client = hedera.ClientForPreviewnet()
	default:
		log.Fatal(fmt.Sprintf("Invalid Client NetworkType provided: [%s]", config.NetworkType))
	}

	accID, err := hedera.AccountIDFromString(config.Operator.AccountId)
	if err != nil {
		log.Fatal(fmt.Sprintf("Invalid Operator AccountId provided: [%s]", config.Operator.AccountId))
	}

	privateKey, err := hedera.Ed25519PrivateKeyFromString(config.Operator.PrivateKey)
	if err != nil {
		log.Fatal(fmt.Sprintf("Invalid Operator PrivateKey provided: [%s]", config.Operator.PrivateKey))
	}

	client.SetOperator(accID, privateKey)

	return &HederaNodeClient{client}
}
