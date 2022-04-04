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

package burn_event

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
	"testing"
)

var (
	eventIdUrlParamKey = "id"
	eventId            = "1"
	transactionId      = "1"
)

func Test_NewRouter(t *testing.T) {
	router := NewRouter(mocks.MBurnService)

	assert.NotNil(t, router)
}
func Test_getTxID(t *testing.T) {
	mocks.Setup()

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)

	txIdResponseContent := transactionId
	var err error
	if err := enc.Encode(txIdResponseContent); err != nil {
		t.Fatalf("Failed to encode response for ResponseWriter. Err: [%s]", err.Error())
	}
	txIdResponseAsBytes := buf.Bytes()
	request := prepareRequest()

	mocks.MBurnService.On("TransactionID", eventId).Return(transactionId, err)
	mocks.MResponseWriter.On("Header").Return(http.Header{})
	mocks.MResponseWriter.On("Write", txIdResponseAsBytes).Return(len(txIdResponseAsBytes), nil)

	txIdResponseHandler := getTxID(mocks.MBurnService)
	txIdResponseHandler(mocks.MResponseWriter, request)

	assert.Nil(t, err)
	assert.NotNil(t, txIdResponseHandler)
	assert.NotNil(t, txIdResponseAsBytes)
}

func Test_getTxID_ErrNotFound(t *testing.T) {
	mocks.Setup()

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)

	txIdResponseContent := response.ErrorResponse(service.ErrNotFound)
	if err := enc.Encode(txIdResponseContent); err != nil {
		t.Fatalf("Failed to encode response for ResponseWriter. Err: [%s]", err.Error())
	}
	txIdResponseAsBytes := buf.Bytes()
	request := prepareRequest()

	mocks.MBurnService.On("TransactionID", eventId).Return(transactionId, service.ErrNotFound)
	mocks.MResponseWriter.On("Header").Return(http.Header{})
	mocks.MResponseWriter.On("Write", txIdResponseAsBytes).Return(len(txIdResponseAsBytes), nil)
	mocks.MResponseWriter.On("WriteHeader", http.StatusNotFound).Return()

	txIdResponseHandler := getTxID(mocks.MBurnService)
	txIdResponseHandler(mocks.MResponseWriter, request)

	assert.NotNil(t, txIdResponseHandler)
	assert.NotNil(t, txIdResponseAsBytes)
	mocks.MBurnService.AssertCalled(t, "TransactionID", eventId)
	mocks.MResponseWriter.AssertCalled(t, "Header")
	mocks.MResponseWriter.AssertCalled(t, "Write", txIdResponseAsBytes)
	mocks.MResponseWriter.AssertCalled(t, "WriteHeader", http.StatusNotFound)
}

func Test_getTxID_InternalServerErr(t *testing.T) {
	mocks.Setup()

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err := errors.New("some internal err")

	txIdResponseContent := response.ErrorResponse(response.ErrorInternalServerError)
	if err := enc.Encode(txIdResponseContent); err != nil {
		t.Fatalf("Failed to encode response for ResponseWriter. Err: [%s]", err.Error())
	}
	txIdResponseAsBytes := buf.Bytes()
	request := prepareRequest()

	mocks.MBurnService.On("TransactionID", eventId).Return(transactionId, err)
	mocks.MResponseWriter.On("Header").Return(http.Header{})
	mocks.MResponseWriter.On("Write", txIdResponseAsBytes).Return(len(txIdResponseAsBytes), nil)
	mocks.MResponseWriter.On("WriteHeader", http.StatusInternalServerError).Return()

	txIdResponseHandler := getTxID(mocks.MBurnService)
	txIdResponseHandler(mocks.MResponseWriter, request)

	assert.NotNil(t, txIdResponseHandler)
	assert.NotNil(t, txIdResponseAsBytes)
	mocks.MBurnService.AssertCalled(t, "TransactionID", eventId)
	mocks.MResponseWriter.AssertCalled(t, "Header")
	mocks.MResponseWriter.AssertCalled(t, "Write", txIdResponseAsBytes)
	mocks.MResponseWriter.AssertCalled(t, "WriteHeader", http.StatusInternalServerError)
}

func prepareRequest() *http.Request {
	request := new(http.Request)
	chiCtx := &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{eventIdUrlParamKey},
			Values: []string{eventId},
		},
	}
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, chiCtx)
	request = request.WithContext(ctx)
	return request
}
