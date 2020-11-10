package connectivity

import (
	"Event-Listener/hedera/config"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
)

func NewClient() *hederasdk.Client {
	var client *hederasdk.Client
	switch config.NetworkType {
	case "testnet":
		client = hederasdk.ClientForTestnet()
	case "mainnet":
		client = hederasdk.ClientForMainnet()
	default:
		panic("Cannot instantiate client. No [config.NetworkType] provided!")
	}

	accountId, _ := hederasdk.AccountIDFromString(config.AccountData.Operator.AccountId)
	privateKey, _ := hederasdk.Ed25519PrivateKeyFromString(config.AccountData.Operator.PrivateKey)

	client.SetOperator(accountId, privateKey)
	return client
}

func MainAccount() (hederasdk.AccountID, error) {
	return hederasdk.AccountIDFromString(config.AccountData.Operator.AccountId)
}

func NewMirrorClient() (hederasdk.MirrorClient, error) {
	return hederasdk.NewMirrorClient(config.MirrorNodeClientAddress)
}
