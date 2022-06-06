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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"

	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/token"

	mirror_node "github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node"

	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	log "github.com/sirupsen/logrus"
)

const mainnetMirrorNodeUrl = `https://mainnet-public.mirrornode.hedera.com:443/api/v1/`
const maxMintAmount = 50
const baseDir = "scripts/token/bulk/nft"

type mainnetToken struct {
	DeployedTokenId string `yaml:"deployed_token_id"`
	Name            string `yaml:"name"`
}

type operator struct {
	PrivateKey string `yaml:"private_key"`
	AccountId  string `yaml:"account_id"`
}

type cfg struct {
	Operator operator       `yaml:"operator"`
	Tokens   []mainnetToken `yaml:"tokens"`
}

type tokenMirror struct {
	config                    cfg
	hederaTestnetClient       *hedera.Client
	hederaMainnetMirrorClient client.MirrorNode
}

func newTokenMirror(config cfg, hederaClient *hedera.Client, hederaMainnetMirrorClient client.MirrorNode) *tokenMirror {
	return &tokenMirror{
		config:                    config,
		hederaTestnetClient:       hederaClient,
		hederaMainnetMirrorClient: hederaMainnetMirrorClient,
	}
}

func main() {
	c := new(cfg)
	err := config.GetConfig(c, fmt.Sprintf("%s/%s", baseDir, "config.yml"))
	if err != nil {
		log.Fatal(err)
	}
	testnetClient := hedera.ClientForTestnet()
	accId, err := hedera.AccountIDFromString(c.Operator.AccountId)
	if err != nil {
		log.Fatal(err)
	}
	pk, err := hedera.PrivateKeyFromString(c.Operator.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	testnetClient.SetOperator(accId, pk)

	mainnetMirrorNode := mirror_node.NewClient(config.MirrorNode{
		ApiAddress:    mainnetMirrorNodeUrl,
		ClientAddress: "hcs.mainnet.mirrornode.hedera.com:5600",
	})

	tm := newTokenMirror(*c, testnetClient, mainnetMirrorNode)

	log.Infof("getting %d tokens and minted NFTs from mainnet", len(c.Tokens))
	tokens, err := tm.collectTokensAndMintedNFTsFromMainnet()
	if err != nil {
		log.Fatal(err)
	}
	err = tm.saveMainnetData(tokens)
	if err != nil {
		log.Fatal(err)
	}
	for i, t := range tokens {
		log.Infof("processing token %s [%d/%d]", t.Token.Name, i+1, len(tokens))
		log.Infof("creating mirrored NFT on testnet for %s", t.Token.Name)
		err := tm.createHederaNft(tokens[i])
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("created mirrored NFT token on testnet for %s with id %s [https://testnet.mirrornode.hedera.com/api/v1/tokens/%s/]", t.Token.Name, t.Token.TokenID, t.Token.TokenID)

		log.Infof("minting NFTs on testnet for %s", t.Token.Name)
		err = tm.mintNftCopies(tokens[i])
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (tm *tokenMirror) saveMainnetData(tokens []*nftWithMinted) error {
	log.Infof("saving mainnet token data")
	for _, t := range tokens {
		data, err := json.MarshalIndent(t, "", "\t")
		if err != nil {
			return err
		}
		err = os.WriteFile(fmt.Sprintf("%s/%s", baseDir, fmt.Sprintf("%s.json", t.Token.Name)), data, 0666)
		if err != nil {
			return err
		}
	}
	return nil
}

type nftWithMinted struct {
	Token  *token.TokenResponse `json:"token"`
	Minted []*transaction.Nft   `json:"minted"`
}

func (tm *tokenMirror) collectTokensAndMintedNFTsFromMainnet() ([]*nftWithMinted, error) {
	res := make([]*nftWithMinted, 0)
	for i, t := range tm.config.Tokens {
		// check if file with token name.json exists
		cur := new(nftWithMinted)
		fileName := fmt.Sprintf("%s/%s.json", baseDir, t.Name)
		if _, err := os.Stat(fileName); err == nil {
			file, err := os.Open(fileName)
			if err != nil {
				return nil, err
			}

			err = json.NewDecoder(file).Decode(cur)
			if err != nil {
				return nil, err
			}
		} else {
			log.Infof("getting token data for %s [%d/%d]", t.Name, i+1, len(tm.config.Tokens))
			mainnetNft, err := tm.hederaMainnetMirrorClient.GetToken(t.DeployedTokenId)
			if err != nil {
				return nil, err
			}

			supply, err := strconv.ParseInt(mainnetNft.TotalSupply, 10, 64)
			if err != nil {
				return nil, err
			}

			log.Infof("getting minted NFTs for %s [%d/%d]", t.Name, i+1, len(tm.config.Tokens))
			nfts, err := tm.fetchAllNFTs(supply, t)
			if err != nil {
				return nil, err
			}

			cur.Token = mainnetNft
			cur.Minted = nfts
		}

		res = append(res, cur)
	}
	return res, nil
}

func (tm *tokenMirror) fetchAllNFTs(supply int64, t mainnetToken) ([]*transaction.Nft, error) {
	res := make([]*transaction.Nft, 0, supply)
	for i := int64(1); i <= supply; i++ {
		log.Infof("[%d/%d] getting NFT with serial number %d", i, supply, i)
		if nft, err := tm.hederaMainnetMirrorClient.GetNft(t.DeployedTokenId, i); err != nil {
			return nil, err
		} else {
			res = append(res, nft)
		}
		if i == maxMintAmount {
			break
		}
	}
	return res, nil
}

func (tm *tokenMirror) createHederaNft(t *nftWithMinted) error {
	s, err := strconv.ParseInt(t.Token.TotalSupply, 10, 64)
	tx := hedera.NewTokenCreateTransaction().
		SetAdminKey(tm.hederaTestnetClient.GetOperatorPublicKey()).
		SetSupplyKey(tm.hederaTestnetClient.GetOperatorPublicKey()).
		SetTreasuryAccountID(tm.hederaTestnetClient.GetOperatorAccountID()).
		SetTokenType(hedera.TokenTypeNonFungibleUnique).
		SetSupplyType(hedera.TokenSupplyTypeFinite).
		SetTokenName(t.Token.Name).
		SetTokenSymbol(t.Token.Symbol).
		SetMaxSupply(s)
	if err != nil {
		return err
	}

	if rf := t.Token.CustomFees.RoyaltyFees; rf != nil {
		fees := make([]hedera.Fee, 0)
		for _, rf := range t.Token.CustomFees.RoyaltyFees {
			fee := hedera.NewCustomRoyaltyFee()
			fee.SetFeeCollectorAccountID(tm.hederaTestnetClient.GetOperatorAccountID())
			fee.SetNumerator(int64(rf.Amount.Numerator))
			fee.SetDenominator(int64(rf.Amount.Denominator))

			if rf.FallbackFee != nil {
				fbFee := hedera.NewCustomFixedFee()
				fbFee.SetAmount(rf.FallbackFee.Amount)
				fbFee.SetFeeCollectorAccountID(tm.hederaTestnetClient.GetOperatorAccountID())

				if rf.FallbackFee.DenominatingTokenId != "" {
					if tokenId, err := hedera.TokenIDFromString(rf.FallbackFee.DenominatingTokenId); err == nil {
						fbFee.SetDenominatingTokenID(tokenId)
					}
				}

				fee.SetFallbackFee(fbFee)
			}

			fees = append(fees, fee)
		}
		tx.SetCustomFees(fees)
	}

	res, err := tx.Execute(tm.hederaTestnetClient)
	if err != nil {
		return err
	}
	receipt, err := res.GetReceipt(tm.hederaTestnetClient)
	if err != nil {
		return err
	}

	t.Token.TokenID = receipt.TokenID.String()
	return nil
}

func (tm *tokenMirror) mintNftCopies(t *nftWithMinted) error {
	tokenId, err := hedera.TokenIDFromString(t.Token.TokenID)
	if err != nil {
		return err
	}

	// Wait for all NFTs to be minted
	supply, err := strconv.ParseInt(t.Token.TotalSupply, 10, 64)
	if err != nil {
		return err
	}
	if supply > maxMintAmount {
		supply = maxMintAmount
	}

	// Send minting transactions to workers
	for _, nft := range t.Minted {
		decodedMeta, err := base64.StdEncoding.DecodeString(nft.Metadata)
		if err != nil {
			log.Fatal(err)
		}

		tx, err := hedera.NewTokenMintTransaction().
			SetTokenID(tokenId).
			SetMetadata(decodedMeta).
			Execute(tm.hederaTestnetClient)
		if err != nil {
			return err
		}

		rx, err := tx.GetReceipt(tm.hederaTestnetClient)
		if err != nil {
			return err
		}

		log.Infof("[%d/%d] minted NFT with serial number %d", rx.SerialNumbers[0], supply, rx.SerialNumbers)
	}

	log.Infof("minted %d/%d NFTs for %s on testnet", supply, supply, t.Token.Name)
	return nil
}
