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
 * imitations under the License.
 */

package http

import (
	"github.com/stretchr/testify/mock"
	"net/http"
)

type MockResponseWriter struct {
	mock.Mock
}

func (m *MockResponseWriter) Header() http.Header {
	args := m.Called()
	return args.Get(0).(http.Header)
}

func (m *MockResponseWriter) WriteHeader(n int) {
	m.Called(n)
}

func (m *MockResponseWriter) Write(b []byte) (n int, err error) {
	args := m.Called(b)
	return args.Get(0).(int), args.Error(1)
}
