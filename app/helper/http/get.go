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

package http

import (
	"encoding/json"
	"github.com/limechain/hedera-eth-bridge-validator/app/domain/client"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func Get(client client.HttpClient, url string, headers map[string]string, responseStruct interface{}, log *log.Entry, statusCode *int) (err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("Error while creating http request struct. Error: [%v]", err)
		return err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Error while sending request to server. Error: [%v]", err)
		return err
	}

	if statusCode != nil {
		*statusCode = resp.StatusCode
	}
	err = json.NewDecoder(resp.Body).Decode(responseStruct)

	return err
}
