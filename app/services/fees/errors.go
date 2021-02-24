/*
 * Copyright 2021 LimeChain Ltd.
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

package fees

import "errors"

var (
	InvalidTransferAmount = errors.New("INVALID_TRANSFER_AMOUNT")
	InvalidTransferFee    = errors.New("INVALID_TRANSFER_FEE")
	InvalidGasPrice       = errors.New("INVALID_GAS_PRICE")
	InsufficientFee       = errors.New("INSUFFICIENT_FEE")
	RateProviderFailure   = errors.New("RATE_PROVIDER_FAILURE")
)
