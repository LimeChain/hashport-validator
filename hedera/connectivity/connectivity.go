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

func NewTestnetClient() *hederasdk.Client {
	accountId, _ := hederasdk.AccountIDFromString(config.AccountData.Operator.AccountId)
	privateKey, _ := hederasdk.Ed25519PrivateKeyFromString(config.AccountData.Operator.PrivateKey)

	client := hederasdk.ClientForTestnet()
	client.SetOperator(accountId, privateKey)

	return client
}

func SubmitTransaction(client *hederasdk.Client, topic hederasdk.ConsensusTopicID, message string) (hederasdk.TransactionReceipt, error) {
	receipt, _ := hederasdk.NewConsensusMessageSubmitTransaction().
		SetTopicID(topic).
		SetMessage([]byte(message)).
		Execute(client)
	return receipt.GetReceipt(client)
}

func MainAccount() (hederasdk.AccountID, error) {
	return hederasdk.AccountIDFromString(config.AccountData.Operator.AccountId)
}

func NewMirrorClient() (hederasdk.MirrorClient, error) {
	return hederasdk.NewMirrorClient(config.MirrorNodeClientAddress)
}
