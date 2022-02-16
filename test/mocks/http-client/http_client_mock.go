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

package http_client

import (
	"github.com/stretchr/testify/mock"
	"net/http"
)

type MockHttpClient struct {
	mock.Mock
}

func (m *MockHttpClient) Get(url string) (resp *http.Response, err error) {
	args := m.Called(url)
	if args[0] == nil && args[1] == nil {
		return nil, nil
	}
	if args[0] == nil {
		return nil, args[1].(error)
	}
	if args[1] == nil {
		return args[0].(*http.Response), nil
	}
	return args[0].(*http.Response), args[1].(error)
}
