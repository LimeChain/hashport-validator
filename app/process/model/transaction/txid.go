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

package transaction

import (
	"fmt"
	"github.com/hashgraph/hedera-sdk-go"
	"strings"
)

// TODO not necessary -> remove
type TxId struct {
	AccountId string
	Seconds   string
	Nanos     string
}

func FromHederaTransactionID(id *hedera.TransactionID) TxId {
	stringTxId := id.String()
	split := strings.Split(stringTxId, "@")
	accId := split[0]

	split = strings.Split(split[1], ".")

	return TxId{
		AccountId: accId,
		Seconds:   split[0],
		Nanos:     split[1],
	}
}

func (txId *TxId) String() string {
	return fmt.Sprintf("%s-%s-%s", txId.AccountId, txId.Seconds, txId.Nanos)
}

func (txId *TxId) Timestamp() string {
	return fmt.Sprintf("%s.%s", txId.Seconds, txId.Nanos)
}
