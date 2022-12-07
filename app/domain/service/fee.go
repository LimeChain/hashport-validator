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

package service

// Fee interface is implemented by the Calculator Service
type Fee interface {
	// CalculateFee calculates the fee and remainder of a given amount, based on a specified token fee percentage
	CalculateFee(token string, amount int64) (fee, remainder int64)

	// CalculatePercentageFee performs the actual percentage calculation with provided params using constants.FeeMaxPercentage
	CalculatePercentageFee(amount int64, feePercentage int64) (fee, remainder int64)
}
