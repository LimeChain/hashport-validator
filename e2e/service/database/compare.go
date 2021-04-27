package database

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"
)

func transfersFieldsMatch(comparing, comparable entity.Transfer) bool {
	return comparable.TransactionID == comparing.TransactionID &&
		comparable.Receiver == comparing.Receiver &&
		comparable.NativeAsset == comparing.NativeAsset &&
		comparable.WrappedAsset == comparing.WrappedAsset &&
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
	return comparing.Status == comparable.Status &&
		comparing.ScheduleID == comparable.ScheduleID &&
		comparing.Recipient == comparable.Recipient &&
		comparing.Amount == comparable.Amount &&
		comparing.Id == comparable.Id
}
