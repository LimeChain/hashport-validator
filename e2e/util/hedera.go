package util

import (
	"github.com/hashgraph/hedera-sdk-go/v2"
	"testing"
)

func GetHederaAccountBalance(client *hedera.Client, account hedera.AccountID, t *testing.T) hedera.AccountBalance {
	// Get bridge account hbar balance before transfer
	receiverBalance, err := hedera.NewAccountBalanceQuery().
		SetAccountID(account).
		Execute(client)
	if err != nil {
		t.Fatalf("Unable to query the balance of the account [%s], Error: [%s]", account.String(), err)
	}
	return receiverBalance
}
