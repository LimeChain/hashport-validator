/*
 * Copyright 2022 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fees

import (
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func Test_NewRouter(t *testing.T) {
	router := NewRouter(mocks.MPricingService, mocks.MFeeService)

	assert.NotNil(t, router)
}

func Test_feesNftResponse(t *testing.T) {
	mocks.Setup()

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)

	var err error
	if err := enc.Encode(testConstants.ParserBridge); err != nil {
		http.Error(mocks.MResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
	bridgeConfigAsBytes := buf.Bytes()
	mocks.MResponseWriter.On("Header").Return(http.Header{})
	mocks.MResponseWriter.On("Write", bridgeConfigAsBytes).Return(len(bridgeConfigAsBytes), nil)

	bridgeResponseHandler := feesNftResponse(mocks.MPricingService)
	bridgeResponseHandler(mocks.MResponseWriter, new(http.Request))

	assert.Nil(t, err)
	assert.NotNil(t, bridgeResponseHandler)
	assert.NotNil(t, bridgeConfigAsBytes)
}

func Test_calculateForResponse(t *testing.T) {
	mocks.Setup()

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)

	bridgeConfigAsBytes := buf.Bytes()

	routeParams := chi.RouteParams{}
	routeParams.Add("network", "8001")
	routeParams.Add("account", "0.0.101")
	routeParams.Add("token", "0.0.3001")
	routeParams.Add("amount", "100")

	mocks.MResponseWriter.On("Header").Return(http.Header{})
	mocks.MResponseWriter.On("Write", bridgeConfigAsBytes).Return(len(bridgeConfigAsBytes), nil)

	bridgeResponseHandler := calculateForResponse(mocks.MFeeService)
	bridgeResponseHandler(mocks.MResponseWriter, new(http.Request))

	assert.NotNil(t, bridgeResponseHandler)
	assert.NotNil(t, bridgeConfigAsBytes)
}
