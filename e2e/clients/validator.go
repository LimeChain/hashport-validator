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

package clients

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

type Validator struct {
	http.Client
	baseUrl string
}

// NewValidatorClient returns new instance of validator client
func NewValidatorClient(url string) *Validator {
	return &Validator{baseUrl: url}
}

func (v *Validator) GetTransferData(transactionID string) ([]byte, error) {
	url := v.baseUrl + "/api/v1/transfers/" + transactionID
	return v.get(url)
}

func (v *Validator) GetEventTransactionID(eventId string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/events/%s/tx", v.baseUrl, eventId)

	bodyBytes, err := v.get(url)
	if err != nil {
		return "", err
	}

	var txID string
	err = json.Unmarshal(bodyBytes, &txID)
	if err != nil {
		return "", err
	}

	return txID, nil
}

func (v *Validator) GetCalculatedFeeFor(targetChainId uint64, account string, token string, amount int64) (int64, error) {
	params := url.Values{
		"targetChain": {strconv.FormatUint(targetChainId, 10)},
		"account":     {account},
		"token":       {token},
		"amount":      {strconv.FormatInt(amount, 10)},
	}

	url := v.baseUrl + "/api/v1/fees/calculate-for?" + params.Encode()

	bodyBytes, err := v.get(url)
	if err != nil {
		return 0, err
	}

	var result int64
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (v *Validator) get(url string) ([]byte, error) {
	response, err := v.Client.Get(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET resolved with status [%d].", response.StatusCode)
	}

	return ioutil.ReadAll(response.Body)
}
