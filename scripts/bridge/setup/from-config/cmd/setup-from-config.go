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
	"math"
	"os"
	"strings"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/config"

	"github.com/limechain/hedera-eth-bridge-validator/constants"
	bridgeSetup "github.com/limechain/hedera-eth-bridge-validator/scripts/bridge/setup"
	"github.com/limechain/hedera-eth-bridge-validator/scripts/bridge/setup/parser"

	"github.com/limechain/hedera-eth-bridge-validator/scripts/bridge/update-config"
	"github.com/limechain/hedera-eth-bridge-validator/scripts/client"
	"github.com/limechain/hedera-eth-bridge-validator/scripts/token/associate"
	nativeFungibleCreate "github.com/limechain/hedera-eth-bridge-validator/scripts/token/native/create"
	nativeNftCreate "github.com/limechain/hedera-eth-bridge-validator/scripts/token/native/nft/create"
	wrappedFungibleCreate "github.com/limechain/hedera-eth-bridge-validator/scripts/token/wrapped/create"
	"gopkg.in/yaml.v2"
)

const (
	HederaMainnetNetworkId = 295
	HederaTestnetNetworkId = 296
	outputFilePath         = "./scripts/bridge/setup/from-config/new-bridge.yml"
)

var (
	hederaNetworkId uint64
)

func main() {
	privateKey := flag.String("privateKey", "0x0", "Hedera Private Key")
	accountID := flag.String("accountID", "0.0", "Hedera Account ID")
	network := flag.String("network", "testnet", "Hedera Network Type")
	memberPrivateKeys := flag.String("memberPrivateKeys", "", "Member private keys array, seperated by ','")
	adminKey := flag.String("adminKey", "", "The admin key")
	configPath := flag.String("configPath", "scripts/bridge/setup/extend-config/extended-bridge.yml", "Path to the 'bridge.yaml' config file")
	flag.Parse()

	prKeysSlice := strings.Split(*memberPrivateKeys, ",")
	var hederaPrivateKeys []hedera.PrivateKey

	members := len(prKeysSlice)

	// element [0] can be empty string if the user does not provide any private keys
	if prKeysSlice[0] != "" {
		for i := 0; i < len(prKeysSlice); i++ {
			privateKeyFromStr, err := hedera.PrivateKeyFromString(prKeysSlice[i])
			if err != nil {
				panic(err)
			}
			hederaPrivateKeys = append(hederaPrivateKeys, privateKeyFromStr)
		}
	}

	treshold := uint(math.Ceil(float64(members) * 0.51))

	validateArguments(privateKey, accountID, adminKey, configPath, network)
	if *network == "testnet" {
		hederaNetworkId = HederaTestnetNetworkId
	} else {
		hederaNetworkId = HederaMainnetNetworkId
	}
	parsedBridgeCfgForDeploy := parseExtendedBridge(configPath)
	bridgeDeployResult := deployBridge(privateKey, accountID, adminKey, network, &members, hederaPrivateKeys, treshold, parsedBridgeCfgForDeploy)
	createAndAssociateTokens(
		&treshold,
		bridgeDeployResult,
		privateKey,
		accountID,
		network,
		parsedBridgeCfgForDeploy,
	)

	newParsedBridge := parsedBridgeCfgForDeploy.ToBridgeParser()
	newBridgeBytes, err := yaml.Marshal(*newParsedBridge)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal updated bridge config to yaml. Err: [%s]", err))
	}

	nodeAccId := "0.0.3"
	content, topicIdParsed, executor, nodeAccount := parseParams(newBridgeBytes, &parsedBridgeCfgForDeploy.ConfigTopicId, accountID, &nodeAccId)
	client := hedera.ClientForTestnet()
	operatorKey, _ := hedera.PrivateKeyFromString(*privateKey)
	operatorId, _ := hedera.AccountIDFromString(*accountID)
	client.SetOperator(operatorId, operatorKey)
	frozenTopicTx := update_config.CreateNewTopicFroxenTx(client, content, topicIdParsed, executor, nodeAccount, 0)
	for _, key := range bridgeDeployResult.MembersPrivateKeys {
		frozenTopicTx.Sign(key)
	}
	tx, err := frozenTopicTx.Execute(client)
	if err != nil {
		panic(fmt.Sprintf("Failed to execute frozen topic tx. Err: [%s]", err))
	}

	txReceipt, err := tx.GetReceipt(client)
	if err != nil {
		panic(fmt.Sprintf("Failed to get receipt for frozen topic tx. Err: [%s]", err))
	}

	fmt.Printf("Update new config topic - Status: %v \n", txReceipt.Status)

	// finally write the file
	err = os.WriteFile(outputFilePath, newBridgeBytes, 0644)
	if err != nil {
		panic(fmt.Sprintf("failed to write new-bridge.yml file. Err: [%s]", err))
	}
	fmt.Printf("Successfully created new config file at: %s\n", outputFilePath)
}

func createAndAssociateTokens(wrappedFungibleThreshold *uint, bridgeDeployResult bridgeSetup.DeployResult, privateKey *string, accountID *string, network *string, parsedBridgeCfgForDeploy *parser.ExtendedBridge) {
	printTitle("Starting Creation and Association of tokens ...")
	supplyKey := hedera.KeyListWithThreshold(*wrappedFungibleThreshold)
	for _, pubKey := range bridgeDeployResult.MembersPublicKeys {
		supplyKey.Add(pubKey)
	}

	client := client.Init(*privateKey, *accountID, *network)
	for network, networkInfo := range parsedBridgeCfgForDeploy.Networks {
		if network == hederaNetworkId {
			createAndAssociateNativeTokens(networkInfo, client, bridgeDeployResult, parsedBridgeCfgForDeploy)
		} else {
			createAndAssociateWrappedTokens(network, networkInfo, client, supplyKey, bridgeDeployResult, parsedBridgeCfgForDeploy)
		}
	}
	fmt.Println("====================================")
}

func deployBridge(privateKey *string, accountID *string, adminKey *string, network *string, members *int, hederaPrivateKeys []hedera.PrivateKey, treshold uint, parsedBridgeCfgForDeploy *parser.ExtendedBridge) bridgeSetup.DeployResult {
	printTitle("Starting Deployment of Bridge ...")
	bridgeDeployResult := bridgeSetup.Deploy(privateKey, accountID, adminKey, network, members, hederaPrivateKeys, treshold)
	if bridgeDeployResult.Error != nil {
		panic(bridgeDeployResult.Error)
	}
	parsedBridgeCfgForDeploy.Networks[hederaNetworkId].BridgeAccount = bridgeDeployResult.BridgeAccountID.String()
	parsedBridgeCfgForDeploy.Networks[hederaNetworkId].PayerAccount = bridgeDeployResult.PayerAccountID.String()
	parsedBridgeCfgForDeploy.Networks[hederaNetworkId].Members = make([]string, len(bridgeDeployResult.MembersPrivateKeys))
	for index, accountId := range bridgeDeployResult.MembersAccountIDs {
		parsedBridgeCfgForDeploy.Networks[hederaNetworkId].Members[index] = accountId.String()
	}
	monitored_accounts := map[string]string{}
	var index_of_monitored_account int = 0
	for key, _ := range parsedBridgeCfgForDeploy.MonitoredAccounts {
		monitored_accounts[key] = bridgeDeployResult.MembersAccountIDs[index_of_monitored_account].String()
		index_of_monitored_account++
	}
	parsedBridgeCfgForDeploy.MonitoredAccounts = monitored_accounts
	parsedBridgeCfgForDeploy.TopicId = bridgeDeployResult.TopicId.String()

	parsedBridgeCfgForDeploy.ConfigTopicId = bridgeDeployResult.ConfigTopicId.String()

	fmt.Println("====================================")
	return bridgeDeployResult
}

func parseExtendedBridge(configPath *string) *parser.ExtendedBridge {
	ExtendedConfig := parser.ExtendedConfig{}
	err := config.GetConfig(&ExtendedConfig, *configPath)
	if err != nil {
		panic("[ERROR] Failed to parse ExtendedBridge config.")
	}
	ExtendedConfig.Bridge.Validate(hederaNetworkId)

	return &ExtendedConfig.Bridge
}

func createAndAssociateWrappedTokens(network uint64, networkInfo *parser.NetworkForDeploy, client *hedera.Client, supplyKey *hedera.KeyList, bridgeDeployResult bridgeSetup.DeployResult, ExtendedBridge *parser.ExtendedBridge) {
	for tokenAddress, tokenInfo := range networkInfo.Tokens.Fungible {
		if _, ok := tokenInfo.Networks[hederaNetworkId]; !ok {
			continue
		}
		if tokenInfo.Decimals > 8 {
			tokenInfo.Decimals = 8
		}
		fmt.Printf("Creating Hedera Wrapped Fungible Token based on info of token with address [%s] ...\n", tokenAddress)
		tokenId, err := wrappedFungibleCreate.WrappedFungibleToken(
			client,
			*bridgeDeployResult.BridgeAccountID,
			client.GetOperatorPublicKey(),
			supplyKey,
			bridgeDeployResult.MembersPrivateKeys,
			tokenInfo.Name,
			tokenInfo.Symbol,
			tokenInfo.Decimals,
			0,
		)
		if err != nil {
			fmt.Printf("[ERROR] Failed to Create Hedera Wrapped Fungible Token based on info of token [%s]. Error: [%s]\n", tokenAddress, err)
			continue
		}
		fmt.Printf("Successfully Created Hedera Wrapped Fungible Token with address [%s] based on info of token [%s]\n", tokenId.String(), tokenAddress)
		ExtendedBridge.Networks[network].Tokens.Fungible[tokenAddress].Networks[hederaNetworkId] = tokenId.String()
	}
}

func createAndAssociateNativeTokens(networkInfo *parser.NetworkForDeploy, client *hedera.Client, bridgeDeployResult bridgeSetup.DeployResult, ExtendedBridge *parser.ExtendedBridge) {
	for tokenAddress, tokenInfo := range networkInfo.Tokens.Fungible {
		if tokenAddress == constants.Hbar {
			continue
		}

		fmt.Printf("Creating Hedera Native Fungible Token based on info of token with address [%s] ...\n", tokenAddress)
		tokenId, err := nativeFungibleCreate.CreateNativeFungibleToken(
			client,
			client.GetOperatorAccountID(),
			tokenInfo.Name,
			tokenInfo.Symbol,
			tokenInfo.Decimals,
			tokenInfo.Supply,
			hedera.HbarFrom(20, "hbar"),
			true,
		)
		if err != nil {
			fmt.Printf("[ERROR] Failed to Created Hedera Native Fungible Token based on info of token [%s]. Error: [%s]\n", tokenAddress, err)
			continue
		}
		fmt.Printf("Successfully Created Hedera Native Fungible Token with address [%s] based on info of token [%s]\n", tokenId.String(), tokenAddress)
		err = associateToken(tokenId, client, *bridgeDeployResult.BridgeAccountID, "Bridge", bridgeDeployResult.MembersPrivateKeys)
		if err == nil {
			delete(ExtendedBridge.Networks[hederaNetworkId].Tokens.Fungible, tokenAddress)
			ExtendedBridge.Networks[hederaNetworkId].Tokens.Fungible[tokenId.String()] = tokenInfo
		}
	}

	for tokenAddress, tokenInfo := range networkInfo.Tokens.Nft {
		fmt.Printf("Creating Hedera Native Non-Fungible Token based on info of token with address [%s] ...\n", tokenAddress)
		tokenId, err := nativeNftCreate.Nft(
			client,
			client.GetOperatorPublicKey(),
			client.GetOperatorAccountID(),
			tokenInfo.Name,
			tokenInfo.Symbol,
			client.GetOperatorPublicKey(),
			true,
		)
		if err != nil {
			fmt.Printf("[ERROR] Failed to Created Hedera Native Non-Fungible Token with address [%s] based on info of token [%s]. Error: [%s]\n", tokenId.String(), tokenAddress, err)
			continue
		}
		fmt.Printf("Successfully Created Hedera Native Non-Fungible Token with address [%s] based on info of token [%s] ...\n", tokenId.String(), tokenAddress)

		errBridge := associateToken(tokenId, client, *bridgeDeployResult.BridgeAccountID, "Bridge", bridgeDeployResult.MembersPrivateKeys)
		errPayer := associateToken(tokenId, client, *bridgeDeployResult.PayerAccountID, "Payer", bridgeDeployResult.MembersPrivateKeys)
		errMembersAcc := associateMembersToToken(client, *tokenId, bridgeDeployResult.MembersAccountIDs, bridgeDeployResult.MembersPrivateKeys)

		if errBridge == nil && errPayer == nil && errMembersAcc == nil {
			delete(ExtendedBridge.Networks[hederaNetworkId].Tokens.Nft, tokenAddress)
			ExtendedBridge.Networks[hederaNetworkId].Tokens.Nft[tokenId.String()] = tokenInfo
		}
	}
}

func associateMembersToToken(
	client *hedera.Client,
	tokenId hedera.TokenID,
	memberAccIDs []hedera.AccountID,
	memberPrivKeys []hedera.PrivateKey,
) error {
	var err error
	for i, memberAccID := range memberAccIDs {
		var singlePKArray = []hedera.PrivateKey{memberPrivKeys[i]}
		_, err = associate.TokenToAccountWithCustodianKey(client, tokenId, memberAccID, singlePKArray)
		if err != nil {
			fmt.Printf("[ERROR] Failed to associate Hedera Native Fungible Token [%s] with %s Member Account. Error: [%s]\n", tokenId.String(), memberAccID, err)
			break
		}
		fmt.Printf("Successfully Associated Hedera Native Fungible Token [%s] with %s Member Account.\n", tokenId.String(), memberAccID)
	}
	return err
}

func associateToken(tokenId *hedera.TokenID, client *hedera.Client, accountId hedera.AccountID, accountName string, custodianKey []hedera.PrivateKey) error {
	fmt.Printf("Associating Hedera Native Fungible Token [%s] with %s Account ...\n", tokenId.String(), accountName)
	_, err := associate.TokenToAccountWithCustodianKey(client, *tokenId, accountId, custodianKey)
	if err != nil {
		fmt.Printf("[ERROR] Failed to associate Hedera Native Fungible Token [%s] with %s Account. Error: [%s]\n", tokenId.String(), accountName, err)
	}
	fmt.Printf("Successfully Associated Hedera Native Fungible Token [%s] with %s Account.\n", tokenId.String(), accountName)
	return err
}

func parseParams(content []byte, topicId *string, executorId *string, nodeAccountId *string) ([]byte, hedera.TopicID, hedera.AccountID, hedera.AccountID) {
	contentLength := len(content)
	if contentLength == 0 {
		panic("config file is empty")
	}
	topicIdParsed, err := hedera.TopicIDFromString(*topicId)
	if err != nil {
		panic(err)
	}
	executor, err := hedera.AccountIDFromString(*executorId)
	if err != nil {
		panic(err)
	}
	nodeAccount, err := hedera.AccountIDFromString(*nodeAccountId)
	if err != nil {
		panic(fmt.Sprintf("Invalid Node Account Id. Err: %s", err))
	}
	return content, topicIdParsed, executor, nodeAccount
}

func validateArguments(privateKey *string, accountID *string, adminKey *string, configPath *string, network *string) {
	err := bridgeSetup.ValidateArguments(privateKey, accountID, adminKey)
	if err != nil {
		panic(err)
	}
	if configPath == nil || *configPath == "" {
		panic("configPath value is missing")
	}
	if network == nil || (*network != "testnet" && *network != "mainnet") {
		panic(fmt.Sprintf("invalid network '%d'", network))
	}

}

func printTitle(title string) {
	delimiterLine := strings.Repeat("=", len(title))
	fmt.Println(delimiterLine)
	fmt.Println(title)
	fmt.Println(delimiterLine)
}
