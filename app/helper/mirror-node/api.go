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

package mirror_node

import (
	"encoding/base64"
	"errors"
	mirrorNodeModel "github.com/limechain/hedera-eth-bridge-validator/app/model/mirror-node"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"strings"
)

func GetUpdatedFileRateFromParsedResponseForHBARPrice(parsedResponse mirrorNodeModel.TransactionsResponse, log *log.Entry) (response mirrorNodeModel.UpdatedFileRateData, err error) {
	if len(parsedResponse.Transactions) == 0 {
		return response, errors.New("No transactions received from HBAR Price Hedera Response.")
	}

	decodedMemo, err := base64.StdEncoding.DecodeString(parsedResponse.Transactions[0].MemoBase64)
	if err != nil {
		log.Errorf("Error while decoding MemoBase64 from HBAR Price Hedera Response. Error: [%v]", err)
		return response, err
	}

	hederaFileRate := parseHederaFileRateFromMemo(decodedMemo, log)
	return hederaFileRate, nil
}

func parseHederaFileRateFromMemo(decodedMemo []byte, log *log.Entry) mirrorNodeModel.UpdatedFileRateData {
	result := mirrorNodeModel.UpdatedFileRateData{}
	asString := string(decodedMemo)
	allFieldsWithValues := strings.Split(asString, ", ")

	var err error
	for _, fieldWithValue := range allFieldsWithValues {
		fieldWithValueTokens := strings.Split(fieldWithValue, " : ")
		fieldName := fieldWithValueTokens[0]
		fieldValue := fieldWithValueTokens[1]

		switch fieldName {
		case "currentRate":
			result.CurrentRate, err = decimal.NewFromString(fieldValue)
			if err != nil {
				log.Errorf("Couldn't parse 'currentRate' from Hedera file for price rate. Error [%v]", err)
			}
		case "nextRate":
			result.NextRate, err = decimal.NewFromString(fieldValue)
			if err != nil {
				log.Errorf("Couldn't parse 'nextRate' from Hedera file for price rate. Error [%v]", err)
			}
		}
	}

	return result
}
