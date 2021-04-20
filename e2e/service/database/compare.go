package database

import "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"

func transfersFieldsMatch(comparing, comparable entity.Transfer) bool {
	return comparable.TransactionID == comparing.TransactionID &&
		comparable.Receiver == comparing.Receiver &&
		comparable.NativeToken == comparing.NativeToken &&
		comparable.WrappedToken == comparing.WrappedToken &&
		comparable.Amount == comparing.Amount &&
		comparable.Status == comparing.Status &&
		comparable.SignatureMsgStatus == comparing.SignatureMsgStatus
}

func messagesFieldsMatch(comparing, comparable entity.Message) bool {
	return comparable.TransferID == comparing.TransferID &&
		comparable.Signature == comparing.Signature &&
		comparable.Hash == comparing.Hash &&
		transfersFieldsMatch(comparable.Transfer, comparing.Transfer) &&
		comparable.Signer == comparing.Signer
}

func burnEventsFieldsMatch(comparing, comparable *entity.BurnEvent) bool {
	return comparable.ScheduleID == comparing.ScheduleID &&
		comparable.TransactionId == comparing.TransactionId &&
		comparable.Status == comparing.Status &&
		comparable.Recipient == comparing.Recipient &&
		comparable.Amount == comparing.Amount &&
		comparable.Id == comparing.Id
}
