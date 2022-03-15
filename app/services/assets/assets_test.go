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

package assets

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/evm/contracts/router"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	"github.com/limechain/hedera-eth-bridge-validator/constants"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

var (
	hederaFeePercentages = make(map[string]int64)
	routerClients        = make(map[uint64]*router.Router)
	evmClients           = make(map[uint64]client.EVM)
	assets               = NewService(testConstants.Networks, hederaFeePercentages, routerClients, mocks.MHederaMirrorClient, evmClients)
)

func Test_LoadAssets(t *testing.T) {
	if reflect.TypeOf(assets).String() != "*assets.Service" {
		t.Fatalf(`Expected to return assets type *assets.Service, but returned: [%s]`, reflect.TypeOf(assets).String())
	}
}

func Test_IsNative(t *testing.T) {

	actual := assets.IsNative(0, constants.Hbar)
	assert.Equal(t, true, actual)

	actual = assets.IsNative(0, "0x0000000000000000000000000000000000000000")
	assert.Equal(t, false, actual)
}

func Test_GetOppositeAsset(t *testing.T) {

	actual := assets.GetOppositeAsset(33, 0, "0x0000000000000000000000000000000000000000")
	expected := constants.Hbar

	assert.Equal(t, expected, actual)

	actual = assets.GetOppositeAsset(0, 33, "0x0000000000000000000000000000000000000001")
	expected = constants.Hbar

	assert.Equal(t, expected, actual)

}

func Test_NativeToWrapped(t *testing.T) {

	actual := assets.NativeToWrapped(constants.Hbar, 0, 33)
	expected := "0x0000000000000000000000000000000000000001"

	assert.Equal(t, expected, actual)
}

func Test_WrappedToNative(t *testing.T) {

	actual := assets.WrappedToNative("0x0000000000000000000000000000000000000001", 33)
	expected := constants.Hbar

	assert.NotNil(t, actual)
	assert.Equal(t, expected, actual.Asset)
}
