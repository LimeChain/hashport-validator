package ethereum

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	"strconv"
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

func EncodeData(ctm *proto.CryptoTransferMessage) ([]byte, error) {
	args, err := generateArguments()
	if err != nil {
		return nil, err
	}

	amountBn, err := helper.ToBigInt(strconv.Itoa(int(ctm.Amount)))
	if err != nil {
		return nil, err
	}
	feeBn, err := helper.ToBigInt(ctm.Fee)
	if err != nil {
		return nil, err
	}

	return args.Pack(
		[]byte(ctm.TransactionId),
		common.HexToAddress(ctm.EthAddress),
		amountBn,
		feeBn)
}

func DecodeSignature(signature string) ([]byte, error) {
	decodedSig, err := hex.DecodeString(signature)
	if err != nil {
		return nil, err
	}

	// note: https://github.com/ethereum/go-ethereum/issues/1975
	decodedSig[64] -= 27

	return decodedSig, nil
}
