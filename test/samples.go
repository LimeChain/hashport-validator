package test

import hederasdk "github.com/hashgraph/hedera-sdk-go"

func NewAccount(client *hederasdk.Client) hederasdk.TransactionID {
	privateKey, _ := hederasdk.GenerateEd25519PrivateKey()
	publicKey := privateKey.PublicKey()

	balance := int64(100)

	newAccount, _ := hederasdk.NewAccountCreateTransaction().
		SetKey(publicKey).
		SetInitialBalance(hederasdk.HbarFromTinybar(balance)).
		Execute(client)

	return newAccount
}

func TopicCreation(client *hederasdk.Client) (hederasdk.TransactionReceipt, error) {
	txId, _ := hederasdk.NewConsensusTopicCreateTransaction().Execute(client)
	return txId.GetReceipt(client)
}

func SubmitMessageToTopic(client *hederasdk.Client, topic hederasdk.ConsensusTopicID, message string) (hederasdk.TransactionReceipt, error) {
	receipt, _ := hederasdk.NewConsensusMessageSubmitTransaction().
		SetTopicID(topic).
		SetMessage([]byte(message)).
		Execute(client)
	return receipt.GetReceipt(client)
}
