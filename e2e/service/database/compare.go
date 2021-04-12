package database

import "github.com/limechain/hedera-eth-bridge-validator/app/persistence/entity"

func transfersFieldsMatch(comparing, comparable entity.Transfer) bool {
	return comparable.TransactionID == comparing.TransactionID &&
		comparable.Receiver == comparing.Receiver &&
		comparable.NativeToken == comparing.NativeToken &&
		comparable.WrappedToken == comparing.WrappedToken &&
		comparable.Amount == comparing.Amount &&
		comparable.TxReimbursement == comparing.TxReimbursement &&
		comparable.Status == comparing.Status &&
		comparable.SignatureMsgStatus == comparing.SignatureMsgStatus &&
		//comparable.EthTxMsgStatus == comparing.EthTxMsgStatus && // TODO: Currently, this field is updated to its final stage, only in the db corresponding to the eth tx executioner.
		comparable.EthTxStatus == comparing.EthTxStatus &&
		comparable.EthTxHash == comparing.EthTxHash
}

func messagesFieldsMatch(comparing, comparable entity.Message) bool {
	return comparable.TransferID == comparing.TransferID &&
		comparable.Signature == comparing.Signature &&
		comparable.Hash == comparing.Hash &&
		transfersFieldsMatch(comparable.Transfer, comparing.Transfer) &&
		comparable.Signer == comparing.Signer
}
