package connectivity

import (
	"Event-Listener/hedera/config"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
)

func NewTestNetAccount(client *hederasdk.Client) hederasdk.TransactionID {
	privateKey, _ := hederasdk.GenerateEd25519PrivateKey()
	publicKey := privateKey.PublicKey()

	balance := int64(100)

	newAccount, _ := hederasdk.NewAccountCreateTransaction().
		SetKey(publicKey).
		SetInitialBalance(hederasdk.HbarFromTinybar(balance)).
		Execute(client)

	return newAccount
}

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
