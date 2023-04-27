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

package validator_version

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/limechain/hedera-eth-bridge-validator/test/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_NewRouter(t *testing.T) {
	router := NewRouter()

	assert.NotNil(t, router)
}

func Test_configBridgeResponse(t *testing.T) {
	mocks.Setup()

	os.Setenv("VTAG", "1.0.0")

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)

	versionResp := &VersionResponse{
		Version: "1.0.0",
	}

	var err error
	if err := enc.Encode(versionResp); err != nil {
		t.Fatalf("Failed to encode response for ResponseWriter. Err: [%s]", err.Error())
	}

	versionResponseAsBytes := buf.Bytes()
	mocks.MResponseWriter.On("Header").Return(http.Header{})
	mocks.MResponseWriter.On("Write", versionResponseAsBytes).Return(len(versionResponseAsBytes), nil)

	versionResponseHandler := versionResponse()
	versionResponseHandler(mocks.MResponseWriter, new(http.Request))

	assert.Nil(t, err)
	assert.NotNil(t, versionResponseHandler)
	assert.NotNil(t, versionResponseAsBytes)
}
