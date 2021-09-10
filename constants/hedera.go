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

package constants

const Hbar = "HBAR"

// Handler topics
const (
	HederaFeeTransfer               = "HEDERA_FEE_TRANSFER"
	HederaTransferMessageSubmission = "HEDERA_TRANSFER_MSG_SUBMISSION"
	HederaBurnMessageSubmission     = "BURN_TOPIC_MSG_SUBMISSION"
	HederaMintHtsTransfer           = "HEDERA_MINT_HTS_TRANSFER"
	TopicMessageSubmission          = "TOPIC_MSG_SUBMISSION"
	TopicMessageValidation          = "TOPIC_MSG_VALIDATION"
)

// Read-only handler topics
const (
	ReadOnlyHederaFeeTransfer     = "READ_ONLY_HEDERA_FEE_TRANSFER"
	ReadOnlyHederaTransfer        = "READ_ONLY_HEDERA_NATIVE_TRANSFER"
	ReadOnlyHederaBurn            = "READ_ONLY_HEDERA_BURN"
	ReadOnlyHederaMintHtsTransfer = "READ_ONLY_HEDERA_MINT_HTS_TRANSFER"
	ReadOnlyTransferSave          = "READ_ONLY_SAVE_TRANSFER"
)
