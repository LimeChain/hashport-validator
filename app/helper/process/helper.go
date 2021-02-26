package process

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	"github.com/limechain/hedera-eth-bridge-validator/app/process/model/transaction"
	"regexp"
	"strings"
)

func IsValidAddress(key string, operatorsEthAddresses []string) bool {
	for _, k := range operatorsEthAddresses {
		if strings.ToLower(k) == strings.ToLower(key) {
			return true
		}
	}
	return false
}

func DecodeMemo(memo string) (*MemoInfo, error) {
	wholeMemoCheck := regexp.MustCompile("^0x([A-Fa-f0-9]){40}-[1-9][0-9]*-[1-9][0-9]*$")

	decodedMemo, e := base64.StdEncoding.DecodeString(memo)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("Could not parse transaction memo: [%s]", e))
	}

	if len(decodedMemo) < 46 || !wholeMemoCheck.MatchString(string(decodedMemo)) {
		return nil, errors.New(fmt.Sprintf("Transaction memo provides invalid or insufficient data - Memo: [%s]", string(decodedMemo)))
	}

	memoSplit := strings.Split(string(decodedMemo), "-")
	ethAddress := memoSplit[0]
	fee := memoSplit[1]
	gasPriceGwei := memoSplit[2]

	return &MemoInfo{EthAddress: ethAddress, Fee: fee, GasPriceGwei: gasPriceGwei}, nil
}

func ExtractAmount(tx transaction.HederaTransaction, accountID hedera.AccountID) int64 {
	var amount int64
	for _, tr := range tx.Transfers {
		if tr.Account == accountID.String() {
			amount = tr.Amount
			break
		}
	}
	return amount
}

type MemoInfo struct {
	EthAddress   string
	Fee          string
	GasPriceGwei string
}
