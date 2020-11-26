package ethereum

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"strconv"
)

var (
	LogEventItemSetName    = "ItemSet"
	logEventItemSetSig     = []byte("ItemSet(uint256,address)")
	LogEventItemSetSigHash = crypto.Keccak256Hash(logEventItemSetSig)
)

func generateArguments() (abi.Arguments, error) {
	bytesType, err := abi.NewType("bytes", "", nil)
	if err != nil {
		return nil, err
	}

	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, err
	}

	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return nil, err
	}

	return abi.Arguments{
		{
			Type: bytesType,
		},
		{
			Type: addressType,
		},
		{
			Type: uint256Type,
		},
		{
			Type: uint256Type,
		}}, nil
}

func EncodeData(txId string, ethAddress string, amount uint64, fee string) ([]byte, error) {
	args, err := generateArguments()
	if err != nil {
		return nil, err
	}

	amountBn, err := helper.ToBigInt(strconv.Itoa(int(amount)))
	if err != nil {
		return nil, err
	}
	feeBn, err := helper.ToBigInt(fee)
	if err != nil {
		return nil, err
	}

	return args.Pack(
		[]byte(txId),
		common.HexToAddress(ethAddress),
		amountBn,
		feeBn)
}
