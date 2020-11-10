package main

import (
	"Event-Listener/hedera/connectivity"
	"Event-Listener/hedera/observer"
	hederasdk "github.com/hashgraph/hedera-sdk-go"
	"time"
)

func main() {
	client := connectivity.NewClient()
	newAccount := connectivity.NewTestNetAccount(client)
	receipt, _ := newAccount.GetReceipt(client)
	newAccountId := receipt.GetAccountID()
	main, _ := connectivity.MainAccount()

	observer.ObserveAccount(main, client)
	observer.ObserveAccount(newAccountId, client)

	for i := 1; i <= 5; i++ {
		_, _ = hederasdk.NewCryptoTransferTransaction().
			AddSender(main, hederasdk.HbarFrom(float64(10*i), "microbar")).
			AddRecipient(newAccountId, hederasdk.HbarFrom(float64(10*i), "microbar")).Execute(client)
		time.Sleep(time.Second * 1)
	}
	time.Sleep(10 * time.Second)

	observer.Stop(main)
	observer.Stop(newAccountId)
	observer.Stop(hederasdk.AccountID{})

	//fmt.Println(main)
	//
	//balance, _ :=  hederasdk.NewAccountBalanceQuery().SetAccountID(main).Execute(client)
	//fmt.Println(balance)
	//
	//_, _ = hederasdk.NewAccountInfoQuery().SetAccountID(main).Execute(client)
	//
	//newBalance, _ :=  hederasdk.NewAccountBalanceQuery().SetAccountID(newAccount.AccountID).Execute(client)
	//fmt.Println(newBalance)
	//
	//res2, _ := hederasdk.NewAccountRecordsQuery().SetAccountID(main).Execute(client)
	//fmt.Println(res2)
	//
	//fmt.Scanln()
	//
	//topicId, _ := handler.TopicCreation(client)
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
	//handler.SubscribeToTopic(mirrorClient, consensusTopicId)
	//handler.SubmitMessageToTopic(client, consensusTopicId, "Someasod messa2ge")
	//
	//fmt.Scanln()
}
