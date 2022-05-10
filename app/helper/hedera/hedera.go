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

package hedera

import (
	"errors"
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/transaction"
)

func OriginatorFromTx(tx transaction.Transaction) (string, error) {
	if tx.TokenTransfers != nil {
		for _, t := range tx.TokenTransfers {
			if t.Amount < 0 {
				return t.Account, nil
			}
		}
	}

	//if tx.NftTransfers != nil {
	//	for _, t := range tx.NftTransfers {
	//		// which transfer?
	//	}
	//}

	if tx.Transfers != nil {
		for _, t := range tx.Transfers {
			if t.Amount < 0 {
				return t.Account, nil
			}
		}
	}

	return "", errors.New("no transfers found")
}
