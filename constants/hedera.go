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
	HederaFeeTransfer               = "HEDERA_FEE_TRANSFER"            // WEVM -> NH
	HederaTransferMessageSubmission = "HEDERA_TRANSFER_MSG_SUBMISSION" // NH -> WEVM
	HederaBurnMessageSubmission     = "BURN_TOPIC_MSG_SUBMISSION"      // WH -> NEVM
	HederaMintHtsTransfer           = "HEDERA_MINT_HTS_TRANSFER"       // NEVM -> WH
	HederaNativeNftTransfer         = "HEDERA_NATIVE_NFT_TRANSFER"     // NH NFT -> WEVM
	TopicMessageSubmission          = "TOPIC_MSG_SUBMISSION"           // WEVM -> WEVM
	TopicMessageValidation          = "TOPIC_MSG_VALIDATION"           // Messages coming from HCS Topic submission
)

// Read-only handler topics
const (
	ReadOnlyHederaFeeTransfer       = "READ_ONLY_HEDERA_FEE_TRANSFER"      // NH -> WEVM
	ReadOnlyHederaTransfer          = "READ_ONLY_HEDERA_NATIVE_TRANSFER"   // WEVM -> NH
	ReadOnlyHederaBurn              = "READ_ONLY_HEDERA_BURN"              // WH -> NEVM
	ReadOnlyHederaMintHtsTransfer   = "READ_ONLY_HEDERA_MINT_HTS_TRANSFER" // NEVM -> WH
	ReadOnlyTransferSave            = "READ_ONLY_SAVE_TRANSFER"            // WEVM -> WEVM
	ReadOnlyHederaNativeNftTransfer = "READ_ONLY_HEDERA_NFT_TRANSFER"      // NH NFT -> WEVM
)
