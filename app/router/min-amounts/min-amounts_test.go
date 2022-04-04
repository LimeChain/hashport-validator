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

package min_amounts

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/response"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func Test_NewRouter(t *testing.T) {
	router := NewRouter(mocks.MPricingService)

	assert.NotNil(t, router)
}

func Test_minAmountsResponse(t *testing.T) {
	mocks.Setup()

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)

	minAmountsResponseContent := testConstants.MinAmountsForApi
	var err error
	if err := enc.Encode(minAmountsResponseContent); err != nil {
		t.Fatalf("Failed to encode response for ResponseWriter. Err: [%s]", err.Error())
	}
	minAmountsResponseAsBytes := buf.Bytes()

	mocks.MPricingService.On("GetMinAmountsForAPI").Return(testConstants.MinAmountsForApi)
	mocks.MResponseWriter.On("Header").Return(http.Header{})
	mocks.MResponseWriter.On("Write", minAmountsResponseAsBytes).Return(len(minAmountsResponseAsBytes), nil)

	minAmountsResponseHandler := minAmountsResponse(mocks.MPricingService)
	minAmountsResponseHandler(mocks.MResponseWriter, new(http.Request))

	assert.Nil(t, err)
	assert.NotNil(t, minAmountsResponseHandler)
	assert.NotNil(t, minAmountsResponseAsBytes)
}

func Test_minAmountsResponse_NoAmounts(t *testing.T) {
	mocks.Setup()

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)

	err := errors.New("Router resolved with an error. Error [No min amount records].")
	minAmountsResponseContent := response.ErrorResponse(err)
	if err := enc.Encode(minAmountsResponseContent); err != nil {
		t.Fatalf("Failed to encode response for ResponseWriter. Err: [%s]", err.Error())
	}
	minAmountsResponseAsBytes := buf.Bytes()

	mocks.MPricingService.On("GetMinAmountsForAPI").Return(make(map[uint64]map[string]string))
	mocks.MResponseWriter.On("Header").Return(http.Header{})
	mocks.MResponseWriter.On("Write", minAmountsResponseAsBytes).Return(len(minAmountsResponseAsBytes), nil)
	mocks.MResponseWriter.On("WriteHeader", http.StatusInternalServerError).Return()

	minAmountsResponseHandler := minAmountsResponse(mocks.MPricingService)
	minAmountsResponseHandler(mocks.MResponseWriter, new(http.Request))

	assert.NotNil(t, minAmountsResponseHandler)
	assert.NotNil(t, minAmountsResponseAsBytes)
	mocks.MResponseWriter.AssertCalled(t, "Header")
	mocks.MResponseWriter.AssertCalled(t, "Write", minAmountsResponseAsBytes)
	mocks.MResponseWriter.AssertCalled(t, "WriteHeader", http.StatusInternalServerError)
	mocks.MPricingService.AssertCalled(t, "GetMinAmountsForAPI")
}
