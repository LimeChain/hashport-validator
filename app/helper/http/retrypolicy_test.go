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

package http_test

import (
	"crypto/x509"
	"net/http"
	"net/url"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	httpHelper "github.com/limechain/hedera-eth-bridge-validator/app/helper/http"
)

type tempErr interface {
	Temporary() bool
}

type temporaryErr struct {
}

func (e temporaryErr) Error() string {
	return "temporary error"
}

func (e temporaryErr) Temporary() bool {
	return true
}

func TestRetryPolicy_ShouldRetryOnTemporaryError(t *testing.T) {
	expected := true
	err := temporaryErr{}

	actual := httpHelper.RetryPolicy(0, err)

	assert.Equal(t, expected, actual)
}

func TestRetryPolicy_ShouldNotRetryOnKnownErrors(t *testing.T) {
	err1 := &url.Error{
		Err: errors.New("stopped after"),
	}
	err2 := &url.Error{
		Err: errors.New("unsupported protocol scheme"),
	}
	err3 := &url.Error{
		Err: errors.New("no Host in request URL"),
	}
	expected := false

	actual1 := httpHelper.RetryPolicy(0, err1)
	actual2 := httpHelper.RetryPolicy(0, err2)
	actual3 := httpHelper.RetryPolicy(0, err3)

	assert.Equal(t, expected, actual1)
	assert.Equal(t, expected, actual2)
	assert.Equal(t, expected, actual3)
}

func TestRetryPolicy_ShouldNotRetryOnUnknownAuthorityError(t *testing.T) {
	err := &url.Error{
		Err: *new(x509.UnknownAuthorityError),
	}

	actual := httpHelper.RetryPolicy(0, err)

	assert.False(t, actual)
}

func TestRetryPolicy_ShouldNotRetryOnCertificateInvalidError(t *testing.T) {
	err := &url.Error{
		Err: *new(x509.CertificateInvalidError),
	}

	actual := httpHelper.RetryPolicy(0, err)

	assert.False(t, actual)
}

func TestRetryPolicy_ShouldNotRetryOnConstraintViolationError(t *testing.T) {
	err := &url.Error{
		Err: *new(x509.ConstraintViolationError),
	}

	actual := httpHelper.RetryPolicy(0, err)

	assert.False(t, actual)
}

func TestRetryPolicy_ShouldRetryOnError(t *testing.T) {
	err := errors.New("generic error")
	expected := true

	actual := httpHelper.RetryPolicy(0, err)

	assert.Equal(t, expected, actual)
}

func TestRetryPolicy_ShouldRetryOnRequestTimeout(t *testing.T) {
	status := http.StatusRequestTimeout

	actual := httpHelper.RetryPolicy(status, nil)

	assert.True(t, actual)
}

func TestRetryPolicy_ShouldRetryOnStatusConflict(t *testing.T) {
	status := http.StatusConflict

	actual := httpHelper.RetryPolicy(status, nil)

	assert.True(t, actual)
}

func TestRetryPolicy_ShouldRetryOnStatusLocked(t *testing.T) {
	status := http.StatusLocked

	actual := httpHelper.RetryPolicy(status, nil)

	assert.True(t, actual)
}

func TestRetryPolicy_ShouldRetryOnStatusTooManyRequests(t *testing.T) {
	status := http.StatusTooManyRequests

	actual := httpHelper.RetryPolicy(status, nil)

	assert.True(t, actual)
}

func TestRetryPolicy_ShouldRetryOnStatusInternalServerError(t *testing.T) {
	status := http.StatusInternalServerError

	actual := httpHelper.RetryPolicy(status, nil)

	assert.True(t, actual)
}

func TestRetryPolicy_ShouldRetryOnStatusBadGateway(t *testing.T) {
	status := http.StatusBadGateway

	actual := httpHelper.RetryPolicy(status, nil)

	assert.True(t, actual)
}

func TestRetryPolicy_ShouldRetryOnStatusServiceUnavailable(t *testing.T) {
	status := http.StatusServiceUnavailable

	actual := httpHelper.RetryPolicy(status, nil)

	assert.True(t, actual)
}

func TestRetryPolicy_ShouldRetryOnStatusGatewayTimeout(t *testing.T) {
	status := http.StatusGatewayTimeout

	actual := httpHelper.RetryPolicy(status, nil)

	assert.True(t, actual)
}

func TestRetryPolicy_ShouldRetryOnStatusInsufficientStorage(t *testing.T) {
	status := http.StatusInsufficientStorage

	actual := httpHelper.RetryPolicy(status, nil)

	assert.True(t, actual)
}

func TestRetryPolicy_ShouldRetryOnEmptyStatusCode(t *testing.T) {
	status := 0

	actual := httpHelper.RetryPolicy(status, nil)

	assert.True(t, actual)
}

func TestRetryPolicy_ShouldNotRetryOnOtherStatusCodes(t *testing.T) {
	status := http.StatusOK

	actual := httpHelper.RetryPolicy(status, nil)

	assert.False(t, actual)
}
