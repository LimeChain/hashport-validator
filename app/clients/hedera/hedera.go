package hedera

import (
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	config "github.com/limechain/hedera-eth-bridge-validator/config"
)

func NewClient(clientData config.Client) *hederasdk.Client {
	var client *hederasdk.Client
	switch clientData.NetworkType {
	case "testnet":
		client = hederasdk.ClientForTestnet()
	case "mainnet":
		client = hederasdk.ClientForMainnet()
	default:
		panic("Cannot instantiate client. No [config.NetworkType] provided!")
	}

	accountId, _ := hederasdk.AccountIDFromString(clientData.Operator.AccountId)
	privateKey, _ := hederasdk.Ed25519PrivateKeyFromString(clientData.Operator.PrivateKey)

	client.SetOperator(accountId, privateKey)
	return client
}

func MainAccount(operator config.Operator) (hederasdk.AccountID, error) {
	return hederasdk.AccountIDFromString(operator.AccountId)
}

func NewMirrorClient(mirrorConfig config.MirrorNode) (hederasdk.MirrorClient, error) {
	return hederasdk.NewMirrorClient(mirrorConfig.Client)
}
