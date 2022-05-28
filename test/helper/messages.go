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

package helper

import (
	"github.com/limechain/hedera-eth-bridge-validator/app/clients/hedera/mirror-node/model/message"
	"github.com/limechain/hedera-eth-bridge-validator/app/helper/timestamp"
)

func MakeMessagePerChunk(chunks []string, consensusTimestampStr, topicId string) []message.Message {
	total := len(chunks)
	msgs := make([]message.Message, total)
	currConsensusTimestampStr := consensusTimestampStr
	currConsensusTimestamp, _ := timestamp.FromString(consensusTimestampStr)
	for i := 0; i < total; i++ {
		number := int64(i + 1)
		msgs[i] = NewMessage(currConsensusTimestampStr, topicId, chunks[i], number, number, int64(total))
		currConsensusTimestamp = currConsensusTimestamp + 1
		currConsensusTimestampStr = timestamp.String(currConsensusTimestamp)
	}

	return msgs
}

func NewMessage(consensusTimestamp, topicId, content string, sequenceNumber, number, total int64) message.Message {
	return message.Message{
		ConsensusTimestamp: consensusTimestamp,
		TopicId:            topicId,
		Contents:           content,
		RunningHash:        "",
		SequenceNumber:     sequenceNumber,
		ChunkInfo: &message.ChunkInfo{
			InitialTransactionId: message.InitialTransactionId{
				AccountId:             "",
				Nonce:                 0,
				Scheduled:             false,
				TransactionValidStart: "",
			},
			Number: number,
			Total:  total,
		},
	}
}
