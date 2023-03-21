package main

/*
go run ./scripts/common/unstuck-hedera-evm/unstuck-hedera-evm.go \
	--signatures "" \
	--transactionId "0.0.1320-1679324186-455906340" \
	--sourceChainId 296 \
	--targetChainId 80001 \
	--targetAsset 0xaA6844fBb7Df9f90FC135f9BCd6F592550e2Fef5 \
	--receiver 0xB075D644d3C46735C8c34AD61a1dEa146950a3F5 \
	--amount 28169014087

// signatures - comma seperated signature values
*/

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"strings"

	auth_message "github.com/limechain/hedera-eth-bridge-validator/app/model/auth-message"
)

type Transfer struct {
	TransactionId string
	SourceChainId uint64
	TargetChainId uint64
	TargetAsset   string
	Receiver      string
	Amount        string
}

func main() {
	signatures := flag.String("signatures", "", "signatures with which the message is signed")

	transactionId := flag.String("transactionId", "", "0.0.1320-1679317679-373659593")
	sourceChainId := flag.Uint64("sourceChainId", 0, "296")
	targetChainId := flag.Uint64("targetChainId", 0, "80001")
	targetAsset := flag.String("targetAsset", "", "0xaA6844fBb7Df9f90FC135f9BCd6F592550e2Fef5")
	receiver := flag.String("receiver", "", "0xB075D644d3C46735C8c34AD61a1dEa146950a3F5")
	amount := flag.String("amount", "", "84772370487")

	flag.Parse()
	if *signatures == "" {
		panic("no signatures provided")
	}
	if *transactionId == "" {
		panic("no transactionId provided")
	}
	if *sourceChainId == 0 {
		panic("no sourceChainId key provided")
	}
	if *targetChainId == 0 {
		panic("no targetChainId key provided")
	}
	if *targetAsset == "" {
		panic("no targetAsset provided")
	}
	if *receiver == "" {
		panic("no receiver provided")
	}
	if *amount == "" {
		panic("no amount provided")
	}

	tm := Transfer{
		TransactionId: *transactionId,
		SourceChainId: *sourceChainId,
		TargetChainId: *targetChainId,
		TargetAsset:   *targetAsset,
		Receiver:      *receiver,
		Amount:        *amount,
	}

	authMsgHash, err := auth_message.EncodeFungibleBytesFrom(
		tm.SourceChainId,
		tm.TargetChainId,
		tm.TransactionId,
		tm.TargetAsset,
		tm.Receiver,
		tm.Amount,
	)
	if err != nil {
		panic(err)
	}

	signatureSlice := strings.Split(*signatures, ",")
	for i := 0; i < len(signatureSlice); i++ {
		decoded, _ := hex.DecodeString(signatureSlice[i])
		pubKey, _ := EcRecover(authMsgHash, decoded)
		fmt.Println(pubKey)
	}

}

func EcRecover(data, sig hexutil.Bytes) (common.Address, error) {
	if len(sig) != 65 {
		return common.Address{}, fmt.Errorf("signature must be 65 bytes long")
	}
	if sig[64] != 27 && sig[64] != 28 {
		return common.Address{}, fmt.Errorf("invalid Ethereum signature (V is not 27 or 28)")
	}
	sig[64] -= 27 // Transform yellow paper V from 27/28 to 0/1

	rpk, err := crypto.Ecrecover(data, sig)
	if err != nil {
		return common.Address{}, err
	}
	pub, _ := crypto.UnmarshalPubkey(rpk)
	recoveredAddr := crypto.PubkeyToAddress(*pub)
	return recoveredAddr, nil
}