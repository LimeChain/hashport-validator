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

package http

import (
	"errors"
	"github.com/limechain/hedera-eth-bridge-validator/config"
	testConstants "github.com/limechain/hedera-eth-bridge-validator/test/constants"
	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

var (
	urlPath  = "http://localhost:80"
	nilErr   error
	headers  = map[string]string{"Accepts": "application/json"}
	respData = &responseData{Name: "some name"}
)

type responseData struct {
	Name string
}

func Test_Get(t *testing.T) {
	request, response := setup(t)

	mocks.MHTTPClient.On("Do", request).Return(response, nilErr)
	err := Get(mocks.MHTTPClient, urlPath, headers, respData, config.GetLoggerFor("Http"), nil)

	assert.Nil(t, err)
}

func Test_Get_ErrorOnSendingRequest(t *testing.T) {
	request, response := setup(t)
	expectedErr := errors.New("something failed")
	mocks.MHTTPClient.On("Do", request).Return(response, expectedErr)

	err := Get(mocks.MHTTPClient, urlPath, headers, respData, config.GetLoggerFor("Http"), nil)

	assert.Equal(t, expectedErr, err)
	mocks.MHTTPClient.AssertCalled(t, "Do", request)
}


func Test_Get_RequestCodeErr(t *testing.T) {
	request, response := setup(t)
	response.StatusCode = 400
	mocks.MHTTPClient.On("Do", request).Return(response, nil)

	var statusCode int 
	_ = Get(mocks.MHTTPClient, urlPath, headers, respData, config.GetLoggerFor("Http"), &statusCode)

	assert.Equal(t, response.StatusCode, statusCode)
	mocks.MHTTPClient.AssertCalled(t, "Do", request)
}

func Test_Get_RequestErr(t *testing.T) {
	mocks.Setup()
	brokenUrlPath := "#%"
	request, _ := http.NewRequest("GET", brokenUrlPath, nil)
	expectedErr := &url.Error{Op: "parse", URL: brokenUrlPath, Err: url.EscapeError("%")}

	err := Get(mocks.MHTTPClient, brokenUrlPath, headers, respData, config.GetLoggerFor("Http"), nil)

	assert.Equal(t, expectedErr, err)
	mocks.MHTTPClient.AssertNotCalled(t, "Do", request)
}

func setup(t *testing.T) (*http.Request, *http.Response) {
	mocks.Setup()
	request, _ := http.NewRequest("GET", urlPath, nil)
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	encodedResponseReaderCloser, encodeErr := EncodeBodyContent(testConstants.SimplePriceResponse)
	if encodeErr != nil {
		t.Fatal(encodeErr)
	}
	response := &http.Response{
		StatusCode: 200,
		Body:       encodedResponseReaderCloser,
	}

	return request, response
}
