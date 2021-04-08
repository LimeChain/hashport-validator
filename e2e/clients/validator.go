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

package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	transfers "github.com/limechain/hedera-eth-bridge-validator/app/domain/service"
	apiresponse "github.com/limechain/hedera-eth-bridge-validator/app/router/response"
)

type Validator struct {
	http.Client
	baseUrl string
}

// NewValidatorClient returns new instance of validator client
func NewValidatorClient(url string) *Validator {
	return &Validator{baseUrl: url}
}

// GetMetadata retrieves the Metadata for a specified gasPrice from the Validator node
func (v *Validator) GetMetadata(gasPriceGwei string) (*apiresponse.MetadataResponse, error) {
	url := v.baseUrl + "/api/v1/metadata?gasPriceGwei=" + gasPriceGwei
	response, err := v.Client.Get(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Get Metadata resolved with status [%d].", response.StatusCode))
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	var metadataResponse *apiresponse.MetadataResponse
	err = json.Unmarshal(bodyBytes, &metadataResponse)
	if err != nil {
		return nil, err
	}

	return metadataResponse, nil
}

func (v *Validator) GetTransferData(transactionID string) (*transfers.TransferData, error) {
	url := v.baseUrl + "/api/v1/transfers/" + transactionID
	response, err := v.Client.Get(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Get Metadata resolved with status [%d].", response.StatusCode))
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	var transferDataResponse *transfers.TransferData
	err = json.Unmarshal(bodyBytes, &transferDataResponse)
	if err != nil {
		return nil, err
	}

	return transferDataResponse, nil
}
