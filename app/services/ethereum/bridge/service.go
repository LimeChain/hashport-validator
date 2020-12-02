package bridge

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	ethclient "github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/ethereum/contracts/bridge"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/proto"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type BridgeContractService struct {
	contractInstance *bridge.Bridge
	client           *ethclient.EthereumClient
}

func NewBridgeContractService(client *ethclient.EthereumClient, config config.Ethereum) *BridgeContractService {
	bridgeContractAddress, err := client.ValidateContractAddress(config.BridgeContractAddress)
	if err != nil {
		log.Fatal(err)
	}

	contractInstance, err := bridge.NewBridge(*bridgeContractAddress, client.Client)
	if err != nil {
		log.Fatalf("Failed to initialize Bridge Contract Instance at [%s]. Error [%s].", config.BridgeContractAddress, err)
	}

	return &BridgeContractService{
		client:           client,
		contractInstance: contractInstance,
	}
}

func (bsc *BridgeContractService) SubmitSignatures(opts *bind.TransactOpts, ctm *proto.CryptoTransferMessage, signatures [][]byte) (*types.Transaction, error) {
	amountBn, err := helper.ToBigInt(strconv.Itoa(int(ctm.Amount)))
	if err != nil {
		return nil, err
	}

	feeBn, err := helper.ToBigInt(ctm.Fee)
	if err != nil {
		return nil, err
	}

	return bsc.contractInstance.Mint(
		opts,
		[]byte(ctm.TransactionId),
		common.HexToAddress(ctm.EthAddress),
		amountBn,
		feeBn,
		signatures)
}

func (bsc *BridgeContractService) WatchBurnEventLogs(opts *bind.WatchOpts, sink chan<- *bridge.BridgeBurn) (event.Subscription, error) {
	return bsc.contractInstance.WatchBurn(opts, sink)
}
