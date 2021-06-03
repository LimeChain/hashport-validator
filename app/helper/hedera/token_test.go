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

package hedera

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	validTokenID   = "0.0.1234"
	invalidTokenID = "0.01234"
)

func Test_IsTokenID(t *testing.T) {
	res := IsTokenID(validTokenID)
	assert.True(t, res)
}

func Test_IsTokenIDError(t *testing.T) {
	res := IsTokenID(invalidTokenID)
	assert.False(t, res)
}
