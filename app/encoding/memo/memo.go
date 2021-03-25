/*
 * Copyright 2021 LimeChain Ltd.
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

package memo

import (
	"encoding/base64"
	"fmt"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/ethereum"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

// Memo represents the *required* encoded information as part of a Transfer to the Bridge Threshold Account
type Memo struct {
	// EthereumAddress that will be the receiver of the funds
	EthereumAddress string
	// TxReimbursementFee that will be paid to the transaction sender
	TxReimbursementFee string
	// GasPrice the gas price that must be used in the mint transaction, converted to wei
	GasPrice string
}

// FromBase64String sanity checks and instantiates new Memo struct from base64 encoded string
func FromBase64String(base64Str string) (*Memo, error) {
	encodingFormat := regexp.MustCompile("^0x([A-Fa-f0-9]){40}-[0-9]+-[0-9]+$")
	decodedMemo, e := base64.StdEncoding.DecodeString(base64Str)
	if e != nil {
		return nil, errors.New(fmt.Sprintf("Invalid base64 string provided: [%s]", e))
	}

	if len(decodedMemo) < 46 || !encodingFormat.MatchString(string(decodedMemo)) {
		return nil, errors.New(fmt.Sprintf("Memo is invalid or has invalid encoding format: [%s]", string(decodedMemo)))
	}

	memoSplit := strings.Split(string(decodedMemo), "-")
	ethAddress := memoSplit[0]
	txReimbursement := strings.TrimLeft(memoSplit[1], "0")
	gasPriceGwei := strings.TrimLeft(memoSplit[2], "0")

	if txReimbursement == "" {
		txReimbursement = "0"
	}
	if gasPriceGwei == "" {
		gasPriceGwei = "0"
	}

	gasPriceGweiBn, err := helper.ToBigInt(gasPriceGwei)
	if err != nil {
		return nil, err
	}

	weiBn := ethereum.GweiToWei(gasPriceGweiBn)

	return &Memo{
		EthereumAddress:    ethAddress,
		TxReimbursementFee: txReimbursement,
		GasPrice:           weiBn.String(),
	}, nil
}
