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
	"bytes"
	"encoding/json"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/helper"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func Test_NewRouter(t *testing.T) {
	router := NewRouter(mocks.MAssetsService)

	assert.NotNil(t, router)
}

func Test_assetsResponse(t *testing.T) {
	mocks.Setup()
	helper.SetupNetworks()

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)

	mocks.MAssetsService.On("FungibleNetworkAssets").Return(testConstants.FungibleNetworkAssets)
	mocks.MAssetsService.On("NonFungibleNetworkAssets").Return(testConstants.NonFungibleNetworkAssets)
	for networkId, networkAssets := range testConstants.FungibleNetworkAssets {
		for _, networkAsset := range networkAssets {
			mocks.MAssetsService.On("FungibleAssetInfo", networkId, networkAsset).
				Return(testConstants.FungibleAssetInfos[networkId][networkAsset], true)
		}
	}
	for networkId, networkAssets := range testConstants.NonFungibleNetworkAssets {
		for _, networkAsset := range networkAssets {
			mocks.MAssetsService.On("NonFungibleAssetInfo", networkId, networkAsset).
				Return(testConstants.NonFungibleAssetInfos[networkId][networkAsset], true)
		}
	}

	assetsResponseContent := generateResponseContent(mocks.MAssetsService)
	var err error
	if err := enc.Encode(assetsResponseContent); err != nil {
		t.Fatalf("Failed to encode response for ResponseWriter. Err: [%s]", err.Error())
	}
	assetsResponseAsBytes := buf.Bytes()

	mocks.MResponseWriter.On("Header").Return(http.Header{})
	mocks.MResponseWriter.On("Write", assetsResponseAsBytes).Return(len(assetsResponseAsBytes), nil)

	assetsResponseHandler := assetsResponse(mocks.MAssetsService)
	assetsResponseHandler(mocks.MResponseWriter, new(http.Request))

	assert.Nil(t, err)
	assert.NotNil(t, assetsResponseHandler)
	assert.NotNil(t, assetsResponseAsBytes)
}
