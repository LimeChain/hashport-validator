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
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strings"
)

type utilsService struct {
	evmClients     map[uint64]client.EVM
	burnEvt        service.BurnEvent
	burnHash       common.Hash
	burnErc721Hash common.Hash
	lockHash       common.Hash
	log            log.FieldLogger
}

func New(evmClients map[uint64]client.EVM, burnEvt service.BurnEvent) *utilsService {
	bridgeAbi, err := abi.JSON(strings.NewReader(router.RouterABI))
	if err != nil {
		log.Fatalf("failed to parse router abi")
		return nil
	}

	return &utilsService{
		evmClients:     evmClients,
		burnEvt:        burnEvt,
		burnHash:       bridgeAbi.Events["Burn"].ID,
		burnErc721Hash: bridgeAbi.Events["BurnERC721"].ID,
		lockHash:       bridgeAbi.Events["Lock"].ID,
		log:            config.GetLoggerFor("Utils Service"),
	}
}

func (s *utilsService) ConvertEvmHashToBridgeTxId(txId string, chainId uint64) (*service.BridgeTxId, error) {
	evmClient, ok := s.evmClients[chainId]
	if !ok {
		return nil, errors.New("invalid chain id")
	}

	txHash := common.HexToHash(txId)
	receipt, err := evmClient.WaitForTransactionReceipt(txHash)
	if err != nil {
		return nil, errors.Wrap(err, "error while waiting for receipt")
	}

	var txIdWithLogIndex string
	for _, l := range receipt.Logs {
		switch l.Topics[0] {
		case s.burnHash, s.burnErc721Hash, s.lockHash:
			txIdWithLogIndex = fmt.Sprintf("%s-%d", txId, l.Index)
			goto finish
		}
	}

finish:
	if txIdWithLogIndex == "" {
		return nil, service.ErrNotFound
	}

	hederaTx, err := s.burnEvt.TransactionID(txIdWithLogIndex)
	if err != nil {
		return nil, err
	}

	return &service.BridgeTxId{
		BridgeTxId: hederaTx,
	}, nil
}
