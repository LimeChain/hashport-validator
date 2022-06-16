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
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
)

func EncodeBodyContent(content interface{}) (io.ReadCloser, error) {
	encodedResponseBuffer := new(bytes.Buffer)
	encodeErr := json.NewEncoder(encodedResponseBuffer).Encode(content)
	if encodeErr != nil {
		return nil, encodeErr
	}
	encodedResponseReader := bytes.NewReader(encodedResponseBuffer.Bytes())
	encodedResponseReaderCloser := ioutil.NopCloser(encodedResponseReader)

	return encodedResponseReaderCloser, nil
}
