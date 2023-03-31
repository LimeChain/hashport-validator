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

// import (
// 	"fmt"
// 	"github.com/hashgraph/hedera-sdk-go/v2"
// 	"github.com/stretchr/testify/assert"
// 	"testing"
// )

// func Test_shouldRetryTransactionInvalidNodeAccount(t *testing.T) {
// 	err := hedera.ErrHederaPreCheckStatus{Status: hedera.StatusInvalidNodeAccount}
// 	assert.True(t, shouldRetryTransaction(&err))
// }

// func Test_shouldRetryNoError(t *testing.T) {
// 	assert.False(t, shouldRetryTransaction(nil))
// }

// func Test_shouldRetryTransactionUnknownError(t *testing.T) {
// 	err := fmt.Errorf("some error")
// 	assert.False(t, shouldRetryTransaction(err))

// 	err = &hedera.ErrHederaPreCheckStatus{Status: hedera.StatusAccountDeleted}
// 	assert.False(t, shouldRetryTransaction(err))
// }
