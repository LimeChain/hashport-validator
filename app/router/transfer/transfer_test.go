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

package transfer

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
	transferIdUrlParamKey = "id"
	transferId            = "1"
	transfer              = service.TransferData{}
)

func Test_NewRouter(t *testing.T) {
	router := NewRouter(mocks.MTransferService)

	assert.NotNil(t, router)
}

func Test_getTransfer(t *testing.T) {
	mocks.Setup()

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)

	transferResponseContent := transfer
	var err error
	if err := enc.Encode(transferResponseContent); err != nil {
		t.Fatalf("Failed to encode response for ResponseWriter. Err: [%s]", err.Error())
	}
	transferResponseAsBytes := buf.Bytes()
	request := prepareRequest()

	mocks.MTransferService.On("TransferData", transferId).Return(transfer, err)
	mocks.MResponseWriter.On("Header").Return(http.Header{})
	mocks.MResponseWriter.On("Write", transferResponseAsBytes).Return(len(transferResponseAsBytes), nil)

	transferResponseHandler := getTransfer(mocks.MTransferService)
	transferResponseHandler(mocks.MResponseWriter, request)

	assert.Nil(t, err)
	assert.NotNil(t, transferResponseHandler)
	assert.NotNil(t, transferResponseAsBytes)
}

func Test_getTransfer_ErrNotFound(t *testing.T) {
	mocks.Setup()

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)

	transferResponseContent := response.ErrorResponse(service.ErrNotFound)
	if err := enc.Encode(transferResponseContent); err != nil {
		t.Fatalf("Failed to encode response for ResponseWriter. Err: [%s]", err.Error())
	}
	transferResponseAsBytes := buf.Bytes()
	request := prepareRequest()

	mocks.MTransferService.On("TransferData", transferId).Return(transfer, service.ErrNotFound)
	mocks.MResponseWriter.On("Header").Return(http.Header{})
	mocks.MResponseWriter.On("Write", transferResponseAsBytes).Return(len(transferResponseAsBytes), nil)
	mocks.MResponseWriter.On("WriteHeader", http.StatusNotFound).Return()

	transferResponseHandler := getTransfer(mocks.MTransferService)
	transferResponseHandler(mocks.MResponseWriter, request)

	assert.NotNil(t, transferResponseHandler)
	assert.NotNil(t, transferResponseAsBytes)
	mocks.MTransferService.AssertCalled(t, "TransferData", transferId)
	mocks.MResponseWriter.AssertCalled(t, "Header")
	mocks.MResponseWriter.AssertCalled(t, "Write", transferResponseAsBytes)
	mocks.MResponseWriter.AssertCalled(t, "WriteHeader", http.StatusNotFound)
}

func Test_getTransfer_InternalServerErr(t *testing.T) {
	mocks.Setup()

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err := errors.New("some internal err")

	transferResponseContent := response.ErrorResponse(response.ErrorInternalServerError)
	if err := enc.Encode(transferResponseContent); err != nil {
		t.Fatalf("Failed to encode response for ResponseWriter. Err: [%s]", err.Error())
	}
	transferResponseAsBytes := buf.Bytes()
	request := prepareRequest()

	mocks.MTransferService.On("TransferData", transferId).Return(transfer, err)
	mocks.MResponseWriter.On("Header").Return(http.Header{})
	mocks.MResponseWriter.On("Write", transferResponseAsBytes).Return(len(transferResponseAsBytes), nil)
	mocks.MResponseWriter.On("WriteHeader", http.StatusInternalServerError).Return()

	transferResponseHandler := getTransfer(mocks.MTransferService)
	transferResponseHandler(mocks.MResponseWriter, request)

	assert.NotNil(t, transferResponseHandler)
	assert.NotNil(t, transferResponseAsBytes)
	mocks.MTransferService.AssertCalled(t, "TransferData", transferId)
	mocks.MResponseWriter.AssertCalled(t, "Header")
	mocks.MResponseWriter.AssertCalled(t, "Write", transferResponseAsBytes)
	mocks.MResponseWriter.AssertCalled(t, "WriteHeader", http.StatusInternalServerError)
}

func prepareRequest() *http.Request {
	request := new(http.Request)
	chiCtx := &chi.Context{
		URLParams: chi.RouteParams{
			Keys:   []string{transferIdUrlParamKey},
			Values: []string{transferId},
		},
	}
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, chiCtx)
	request = request.WithContext(ctx)
	return request
}
