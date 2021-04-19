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
