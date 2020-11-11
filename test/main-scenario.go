package test

import (
	"Event-Listener/app/clients/hedera"
	"Event-Listener/app/process/watcher/crypto-transfer"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	"time"
)

func main() {
	client := hedera.NewClient()
	newAccount := NewAccount(client)
	receipt, _ := newAccount.GetReceipt(client)
	newAccountId := receipt.GetAccountID()
	main, _ := hedera.MainAccount()

	watcher := crypto_transfer.CryptoTransferWatcher{Account: main}
	watcher.Watch()

	watcherNew := crypto_transfer.CryptoTransferWatcher{Account: newAccountId}
	watcherNew.Watch()

	for i := 1; i <= 5; i++ {
		_, _ = hederasdk.NewCryptoTransferTransaction().
			AddSender(main, hederasdk.HbarFrom(float64(10*i), "microbar")).
			AddRecipient(newAccountId, hederasdk.HbarFrom(float64(10*i), "microbar")).Execute(client)
		time.Sleep(time.Second * 1)
	}
	time.Sleep(10 * time.Second)

	//fmt.Println(sample-scenarios)
	//
	//balance, _ :=  hederasdk.NewAccountBalanceQuery().SetAccountID(sample-scenarios).Execute(client)
	//fmt.Println(balance)
	//
	//_, _ = hederasdk.NewAccountInfoQuery().SetAccountID(sample-scenarios).Execute(client)
	//
	//newBalance, _ :=  hederasdk.NewAccountBalanceQuery().SetAccountID(newAccount.AccountID).Execute(client)
	//fmt.Println(newBalance)
	//
	//res2, _ := hederasdk.NewAccountRecordsQuery().SetAccountID(sample-scenarios).Execute(client)
	//fmt.Println(res2)
	//
	//fmt.Scanln()
	//
	//topicId, _ := consensus-topic.TopicCreation(client)
	//consensusTopicId := topicId.GetConsensusTopicID()
	//fmt.Println(consensusTopicId)
	//
	//_, e := hederasdk.NewConsensusTopicInfoQuery().SetTopicID(consensusTopicId).Execute(client)
	//if e!=nil {
	//	fmt.Println(e)
	//	return
	//}
	//
	//mirrorClient, e := connectivity.NewMirrorClient()
	//if e!=nil{
	//	fmt.Println(e)
	//	return
	//}
	//
	//time.Sleep(30 * time.Second)
	//
	//consensus-topic.SubscribeToTopic(mirrorClient, consensusTopicId)
	//consensus-topic.SubmitMessageToTopic(client, consensusTopicId, "Someasod messa2ge")
	//
	//fmt.Scanln()
}
