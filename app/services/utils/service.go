/*
 * Copyright 2022 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/pkg/errors"
	"strings"
)

type utilsService struct {
	evmClients map[uint64]client.EVM
	burnEvt    service.BurnEvent
}

func New(evmClients map[uint64]client.EVM, burnEvt service.BurnEvent) *utilsService {
	return &utilsService{
		evmClients: evmClients,
		burnEvt:    burnEvt,
	}
}

func (s *utilsService) ConvertEvmTxIdToHederaTxId(txId string, chainId uint64) (*service.HederaTxId, error) {
	bridgeAbi, err := abi.JSON(strings.NewReader(router.RouterABI))
	if err != nil {
		return nil, err
	}

	burnHash := bridgeAbi.Events["Burn"].ID
	lockHash := bridgeAbi.Events["Lock"].ID
	burnERC721Hash := bridgeAbi.Events["BurnERC721"].ID

	client, ok := s.evmClients[chainId]
	if !ok {
		return nil, errors.New("invalid chain id")
	}

	txHash := common.HexToHash(txId)
	receipt, err := client.WaitForTransactionReceipt(txHash)
	if err != nil {
		return nil, errors.Wrap(err, "error while waiting for receipt")
	}

	logIdx := 0
	for i, log := range receipt.Logs {
		switch log.Topics[0] {
		case burnHash:
			logIdx = i
		case lockHash:
			logIdx = i
		case burnERC721Hash:
			logIdx = i
			fallthrough
		default:

		}
	}

	txIdWithLogIndex := fmt.Sprintf("%s-%d", txId, logIdx)
	hederaTx, err := s.burnEvt.TransactionID(txIdWithLogIndex)
	if err != nil {
		return nil, err
	}

	return &service.HederaTxId{
		HederaTxId: hederaTx,
	}, nil
}
