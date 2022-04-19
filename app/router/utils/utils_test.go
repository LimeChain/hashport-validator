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

package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	"github.com/limechain/hedera-eth-bridge-validator/app/router/response"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strconv"
	"testing"
)

var (
	router           chi.Router
	evmHash          = "0xa83be7d95c58f57e11f5c27dedd963217d47bdeab897bc98f2f5410d9f6c0026"
	chainId          = uint64(80001)
	expectedBridgeTx = "0.0.1-123-123"
	conversionResult = &service.BridgeTxId{
		BridgeTxId: expectedBridgeTx,
	}
)

func setup() {
	mocks.Setup()

	router = NewRouter(mocks.MUtilsService)
}

func prepareRequest() *http.Request {
	req := new(http.Request)
	routeParams := chi.RouteParams{}
	routeParams.Add("evmHash", evmHash)
	routeParams.Add("chainId", strconv.FormatUint(chainId, 10))
	chiCtx := &chi.Context{
		URLParams: routeParams,
	}
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, chiCtx)
	req = req.WithContext(ctx)
	return req
}

func Test_NewRouter(t *testing.T) {
	setup()

	actual := NewRouter(mocks.MUtilsService)

	assert.NotNil(t, actual)
}

func Test_convertEvmTxHashToBridgeTxId(t *testing.T) {
	setup()

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	responseContent := conversionResult
	if err := enc.Encode(responseContent); err != nil {
		t.Fatalf("Failed to encode response for ResponseWriter. Err: [%s]", err.Error())
	}
	responseContentBytes := buf.Bytes()
	req := prepareRequest()

	mocks.MUtilsService.
		On("ConvertEvmHashToBridgeTxId", evmHash, chainId).
		Return(conversionResult, nil)
	mocks.MResponseWriter.On("Header").Return(http.Header{})
	mocks.MResponseWriter.On("Write", responseContentBytes).Return(len(responseContentBytes), nil)

	convertEvmTxHashToBridgeTxId(mocks.MUtilsService)(mocks.MResponseWriter, req)

	mocks.MUtilsService.AssertCalled(t, "ConvertEvmHashToBridgeTxId", evmHash, chainId)
	mocks.MResponseWriter.AssertCalled(t, "Header")
	mocks.MResponseWriter.AssertCalled(t, "Write", responseContentBytes)
}

func Test_convertEvmTxHashToBridgeTxId_NotFound(t *testing.T) {
	setup()

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	responseContent := response.ErrorResponse(service.ErrNotFound)
	if err := enc.Encode(responseContent); err != nil {
		t.Fatalf("Failed to encode response for ResponseWriter. Err: [%s]", err.Error())
	}
	responseContentBytes := buf.Bytes()
	req := prepareRequest()

	mocks.MUtilsService.
		On("ConvertEvmHashToBridgeTxId", evmHash, chainId).
		Return(conversionResult, service.ErrNotFound)
	mocks.MResponseWriter.On("Header").Return(http.Header{})
	mocks.MResponseWriter.On("Write", responseContentBytes).Return(len(responseContentBytes), nil)
	mocks.MResponseWriter.On("WriteHeader", http.StatusNotFound).Return()

	convertEvmTxHashToBridgeTxId(mocks.MUtilsService)(mocks.MResponseWriter, req)

	mocks.MUtilsService.AssertCalled(t, "ConvertEvmHashToBridgeTxId", evmHash, chainId)
	mocks.MResponseWriter.AssertCalled(t, "Header")
	mocks.MResponseWriter.AssertCalled(t, "Write", responseContentBytes)
	mocks.MResponseWriter.AssertCalled(t, "WriteHeader", http.StatusNotFound)
}

func Test_convertEvmTxHashToBridgeTxId_InternalServerError(t *testing.T) {
	setup()
	err := errors.New("some error")
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	responseContent := response.ErrorResponse(response.ErrorInternalServerError)
	if err := enc.Encode(responseContent); err != nil {
		t.Fatalf("Failed to encode response for ResponseWriter. Err: [%s]", err.Error())
	}
	responseContentBytes := buf.Bytes()
	req := prepareRequest()

	mocks.MUtilsService.
		On("ConvertEvmHashToBridgeTxId", evmHash, chainId).
		Return(conversionResult, err)
	mocks.MResponseWriter.On("Header").Return(http.Header{})
	mocks.MResponseWriter.On("Write", responseContentBytes).Return(len(responseContentBytes), nil)
	mocks.MResponseWriter.On("WriteHeader", http.StatusInternalServerError).Return()

	convertEvmTxHashToBridgeTxId(mocks.MUtilsService)(mocks.MResponseWriter, req)

	mocks.MUtilsService.AssertCalled(t, "ConvertEvmHashToBridgeTxId", evmHash, chainId)
	mocks.MResponseWriter.AssertCalled(t, "Header")
	mocks.MResponseWriter.AssertCalled(t, "Write", responseContentBytes)
	mocks.MResponseWriter.AssertCalled(t, "WriteHeader", http.StatusInternalServerError)
}
