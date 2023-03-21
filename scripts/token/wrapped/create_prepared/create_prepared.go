package main

import (
	"encoding/hex"
	"flag"
	"fmt"

	utils "github.com/limechain/hedera-eth-bridge-validator/scripts/client"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"strings"
	"time"
)

/* -Run the script like:

go run ./scripts/token/wrapped/create_prepared/create_prepared.go \
    --executorAccountID 0.0.540286 \
    --network mainnet \
    --bridgeID 0.0.540219 \
    --tokenName "Token Name" \
    --tokenSymbol "symbol[hts]" \
    --threshold 5 \
    --memberKeys "25075f86a93019a22420c845b14add02c7f80acce5ee63aa2476cc928247e53a,5eca016e197434e11afd8008d39f83d2d6b7342dbb3745526b53dd4071f6eeeb,1b44857eb54139c20c2b3d548d8f45308d546a7a974f666a90435d48969834f3,ec4f80087c309eaf50d814b504d2be08577149456470a6484f34984ccbbcf4af,db439b30e19e66177dae79c4018e488ddeff2b62747ad2e4f7a8184cfd264a79,1e7129725857fbdae26ba032092c232fd3d720c60b0b292f25bea1166ca410b7,99ccf8a1f660d62f7268abe2b14357a4b11ab445ba04ccc7f3e842707b846a3f,bf2468d0475d4a9bf002801b20504e5db7c3a9f0593096f311815767401c51d4,5b6dd26306ecdc89013f067613afd535384a7c3b6abd96f224444e391c7d5bba" \
    --initialSupply 0 \
    --decimals 8 \
    --nodeAccountID 0.0.3

*/

func main() {
	executorAccountID := flag.String("executorAccountID", "0.0", "Hedera Executor Account ID")
	network := flag.String("network", "", "Hedera Network")
	bridgeID := flag.String("bridgeID", "0.0", "Bridge account ID")
	tokenName := flag.String("tokenName", "", "Token Name")
	tokenSymbol := flag.String("tokenSymbol", "", "Token Symbol")
	threshold := flag.Uint("threshold", 0, "Threshold Key")
	memberKeys := flag.String("memberKeys", "", "The keys of the validators accounts from the mirror node")
	initialSupply := flag.Uint64("initialSupply", 0, "Token initial supply")
	decimals := flag.Uint("decimals", 8, "Token decimals")
	nodeAccountId := flag.String("nodeAccountID", "0.0.3", "Node account id on which to process the transaction.")
	flag.Parse()

	if *network == "" {
		panic("network not provided")
	}
	if *bridgeID == "0.0" {
		panic("sender account id was not provided")
	}
	if *executorAccountID == "0.0" {
		panic("executor account id was not provided")
	}
	if *tokenName == "" {
		panic("invalid token name")
	}
	if *tokenSymbol == "" {
		panic("invalid token symbol")
	}
	if *threshold == 0 {
		panic("threshold not provided")
	}
	if *memberKeys == "" {
		panic("validators keys not provided")
	}
	if *decimals == 0 {
		panic("decimals not provided")
	}

	executor, err := hedera.AccountIDFromString(*executorAccountID)
	if err != nil {
		panic(err)
	}
	nodeAccount, er := hedera.AccountIDFromString(*nodeAccountId)
	if er != nil {
		panic(fmt.Sprintf("Invalid Node Account Id. Err: %s", er))
	}
	bridgeAccount, err := hedera.AccountIDFromString(*bridgeID)
	if err != nil {
		panic(err)
	}

	validatorsSlice := strings.Split(*memberKeys, ",")
	validatorsPublicKeys := hedera.KeyListWithThreshold(*threshold)
	for _, sk := range validatorsSlice {
		key, err := hedera.PublicKeyFromStringEd25519(sk)
		if err != nil {
			panic(fmt.Sprintf("failed to parse supply key [%s]. error [%s]", sk, err))
		}
		validatorsPublicKeys.Add(key)
	}

	client := utils.GetClientForNetwork(*network)
	additionTime := time.Minute * 1 // 2 minutes
	transactionID := hedera.NewTransactionIDWithValidStart(executor, time.Now().Add(additionTime))

	frozen, err := hedera.NewTokenCreateTransaction().
		SetAutoRenewAccount(executor).
		SetNodeAccountIDs([]hedera.AccountID{nodeAccount}).
		SetTransactionID(transactionID).
		SetTreasuryAccountID(bridgeAccount).
		SetAdminKey(validatorsPublicKeys).
		SetSupplyKey(validatorsPublicKeys).
		SetPauseKey(validatorsPublicKeys).
		SetTokenName(*tokenName).
		SetTokenSymbol(*tokenSymbol).
		SetInitialSupply(*initialSupply).
		SetDecimals(*decimals).
		FreezeWith(client)

	if err != nil {
		panic(err)
	}
	bytes, err := frozen.ToBytes()
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(bytes))
}
