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

package main

import (
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/wtoken"
	mirrorNode "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	"github.com/limechain/hedera-eth-bridge-validator/scripts/bridge/setup/parser"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math"
	"strconv"
)

const (
	HederaMainnetNetworkId = 295
	HederaTestnetNetworkId = 296
	outputFilePath         = "./scripts/bridge/setup/extend-config/extended-bridge.yml"
	defaultSupply          = uint64(100000000000000)
	evmBlockConfirmations  = uint64(5)
)

var (
	///////////////////////////////////////
	// FILL THE EVM NODE URLS BEFORE RUN //
	///////////////////////////////////////
	evmNodeUrls = map[uint64]string{
		80001: "https://rpc-mumbai.maticvigil.com/",
		5:     "https://goerli-light.eth.linkpool.io/",
		97:    "https://data-seed-prebsc-1-s1.binance.org:8545/",
		43113: "https://api.avax-test.network/ext/bc/C/rpc",
	}
	evmClients = make(map[uint64]client.EVM)

	// Hedera //
	hederaNetworkId           uint64
	mirrorNodeConfigByNetwork = map[uint64]config.MirrorNode{
		HederaMainnetNetworkId: {
			ClientAddress: "mainnet-public.mirrornode.hedera.com/:443",
			ApiAddress:    "https://mainnet.mirrornode.hedera.com/api/v1/",
		},
		HederaTestnetNetworkId: {
			ClientAddress: "hcs.testnet.mirrornode.hedera.com:5600",
			ApiAddress:    "https://testnet.mirrornode.hedera.com/api/v1/",
		},
	}
)

func main() {
	evmPrivateKey := flag.String("evmPrivateKey", "", "EVM Private Key")
	network := flag.String("network", "testnet", "Hedera Network Type")
	configPath := flag.String("configPath", "scripts/bridge/setup/extend-config/bridge.yml", "Path to the 'bridge.yaml' config file")
	flag.Parse()

	validateArguments(configPath, network, evmPrivateKey)
	if *network == "testnet" {
		hederaNetworkId = HederaTestnetNetworkId
	} else {
		hederaNetworkId = HederaMainnetNetworkId
	}

	mirrorNodeClient := mirrorNode.NewClient(mirrorNodeConfigByNetwork[hederaNetworkId])
	extendedBridgeCfg := parseExtendedBridge(configPath)
	updateAdditionalFieldsToCfg(extendedBridgeCfg, evmPrivateKey, mirrorNodeClient)
	createOutputFile(extendedBridgeCfg)
	fmt.Printf("Successfully created extended config file at: %s\n", outputFilePath)
}

func createOutputFile(extendedBridgeCfg *parser.ExtendedBridge) {
	extendedConfig := parser.ExtendedConfig{Bridge: *extendedBridgeCfg}
	extendedBridgeYml, err := yaml.Marshal(extendedConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal extended bridge config to yaml. Err: [%s]", err))
	}
	err = ioutil.WriteFile(outputFilePath, extendedBridgeYml, 0644)
	if err != nil {
		panic(fmt.Sprintf("failed to write new-bridge.yml file. Err: [%s]", err))
	}
}

func updateAdditionalFieldsToCfg(extendedBridgeCfg *parser.ExtendedBridge, evmPrivateKey *string, mirrorNodeClient *mirrorNode.Client) {
	for networkId, networkContent := range extendedBridgeCfg.Networks {
		if networkId != hederaNetworkId {

			nodeUrl, ok := evmNodeUrls[networkId]
			if !ok {
				panic(fmt.Sprintf("EVM nodeUrl missing for chainId: %d. Note: Add it in the script at 'evmNodeUrls' map.", networkId))
			}

			evmClients[networkId] = evm.NewClient(config.Evm{
				BlockConfirmations: evmBlockConfirmations,
				NodeUrl:            nodeUrl,
				PrivateKey:         *evmPrivateKey,
			}, networkId)
		}

		// Fungible Tokens
		for fungibleTokenAddress, fungibleTokenInfo := range networkContent.Tokens.Fungible {
			if networkId == hederaNetworkId {
				err := updateHederaFungibleAssetInfo(fungibleTokenAddress, fungibleTokenInfo, mirrorNodeClient)
				if err != nil {
					panic(err)
				}
			} else {
				token, err := wtoken.NewWtoken(common.HexToAddress(fungibleTokenAddress), evmClients[networkId])
				if err != nil {
					panic(fmt.Sprintf("EVM Fungible Token client failed to initialize for address [%s]. Err: [%s]", fungibleTokenAddress, err))
				}
				err = updateEvmFungibleAssetInfo(fungibleTokenInfo, networkId, fungibleTokenAddress, token)
				if err != nil {
					panic(err)
				}
			}
		}

		// Non-Fungible Tokens
		for nftAddress, nftInfo := range networkContent.Tokens.Nft {
			if networkId == hederaNetworkId {
				err := updateHederaNonFungibleAssetInfo(nftAddress, nftInfo, mirrorNodeClient)
				if err != nil {
					panic(err)
				}
			} else {
				token, err := wtoken.NewWtoken(common.HexToAddress(nftAddress), evmClients[networkId])
				if err != nil {
					panic(fmt.Sprintf("EVM NFT client failed to initialize for address [%s]. Err: [%s]", nftAddress, err))
				}
				err = updateEvmNonFungibleAssetInfo(nftInfo, networkId, nftAddress, token)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func updateHederaFungibleAssetInfo(
	assetId string,
	assetInfo *parser.FungibleTokenForDeploy,
	mirrorNode client.MirrorNode,
) error {

	if assetId == constants.Hbar {
		assetInfo.Name = constants.Hbar
		assetInfo.Symbol = constants.Hbar
		assetInfo.Decimals = uint(constants.HederaDefaultDecimals)
		assetInfo.Supply = 50_000_000_000

		return nil
	}

	assetInfoResponse, err := mirrorNode.GetToken(assetId)
	if err != nil {
		log.Errorf("Hedera Mirror Node method GetToken for Asset [%s] - Error: [%s]", assetId, err)
		return err
	}

	assetInfo.Name = assetInfoResponse.Name
	assetInfo.Symbol = assetInfoResponse.Symbol
	parsedDecimals, _ := strconv.Atoi(assetInfoResponse.Decimals)
	assetInfo.Decimals = uint(parsedDecimals)
	parsedTotalSupply, _ := strconv.Atoi(assetInfoResponse.TotalSupply)
	assetInfo.Supply = uint64(parsedTotalSupply)
	if assetInfo.Supply == 0 {
		assetInfo.Supply = defaultSupply
	} else if assetInfo.Supply > math.MaxInt64 || assetInfo.Supply < 0 {
		assetInfo.Supply = math.MaxInt64
	}

	return nil
}

func updateHederaNonFungibleAssetInfo(
	assetId string,
	assetInfo *parser.NonFungibleTokenForDeploy,
	mirrorNode client.MirrorNode,
) error {

	assetInfoResponse, e := mirrorNode.GetToken(assetId)
	if e != nil {
		log.Errorf("Hedera Mirror Node method GetToken for Asset [%s] - Error: [%s]", assetId, e)
	} else {
		assetInfo.Name = assetInfoResponse.Name
		assetInfo.Symbol = assetInfoResponse.Symbol
	}

	return e
}

func updateEvmFungibleAssetInfo(
	assetInfo *parser.FungibleTokenForDeploy,
	networkId uint64,
	assetAddress string,
	evmTokenClient client.EvmFungibleToken,
) (err error) {
	name, err := evmTokenClient.Name(&bind.CallOpts{})
	if err != nil {
		log.Errorf("Failed to get Name for Asset [%s] for EVM with networkId [%d]  - Error: [%s]", assetAddress, networkId, err)
		return err
	}
	assetInfo.Name = name

	symbol, err := evmTokenClient.Symbol(&bind.CallOpts{})
	if err != nil {
		log.Errorf("EVM with networkId [%d] for Asset [%s], and method Symbol - Error: [%s]", networkId, assetAddress, err)
		return err
	}
	assetInfo.Symbol = symbol

	decimals, err := evmTokenClient.Decimals(&bind.CallOpts{})
	if err != nil {
		log.Errorf("EVM with networkId [%d] for Asset [%s], and method Decimals - Error: [%s]", networkId, assetAddress, err)
		return err
	}
	assetInfo.Decimals = uint(decimals)
	totalSupply, err := evmTokenClient.TotalSupply(&bind.CallOpts{})
	if err != nil {
		log.Errorf("EVM with networkId [%d] for Asset [%s], and method Total supply - Error: [%s]", networkId, assetAddress, err)
		return err
	}
	assetInfo.Supply = totalSupply.Uint64()
	if assetInfo.Supply == 0 {
		assetInfo.Supply = defaultSupply
	} else if assetInfo.Supply > math.MaxInt64 || assetInfo.Supply < 0 {
		assetInfo.Supply = math.MaxInt64
	}

	return err
}

func updateEvmNonFungibleAssetInfo(
	assetInfo *parser.NonFungibleTokenForDeploy,
	networkId uint64,
	assetAddress string,
	evmTokenClient client.EvmNft,
) (err error) {
	name, err := evmTokenClient.Name(&bind.CallOpts{})
	if err != nil {
		log.Errorf("Failed to get Name for Asset [%s] for EVM with networkId [%d]  - Error: [%s]", assetAddress, networkId, err)
		return err
	}
	assetInfo.Name = name

	symbol, err := evmTokenClient.Symbol(&bind.CallOpts{})
	if err != nil {
		log.Errorf("EVM with networkId [%d] for Asset [%s], and method Symbol - Error: [%s]", networkId, assetAddress, err)
		return err
	}
	assetInfo.Symbol = symbol

	return nil
}

func parseExtendedBridge(configPath *string) *parser.ExtendedBridge {
	ExtendedConfig := parser.ExtendedConfig{}
	err := config.GetConfig(&ExtendedConfig, *configPath)
	if err != nil {
		panic("[ERROR] Failed to parse Bridge config.")
	}

	return &ExtendedConfig.Bridge
}

func validateArguments(configPath *string, network *string, evmPrivateKey *string) {
	if configPath == nil || *configPath == "" {
		panic("configPath value is missing")
	}
	if network == nil || (*network != "testnet" && *network != "mainnet") {
		panic(fmt.Sprintf("invalid network '%d'", network))
	}
	if evmPrivateKey == nil || *evmPrivateKey == "" {
		panic("evmPrivateKey value is missing")
	}
}
