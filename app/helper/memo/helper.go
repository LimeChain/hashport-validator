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

package memo

import (
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"regexp"
)

// ValidateMemo sanity checks and instantiates new Memo struct from base64 encoded string
func ValidateMemo(ethAddress string) error {
	encodingFormat := regexp.MustCompile("^0x([A-Fa-f0-9]){40}$")
	decodedMemo, e := base64.StdEncoding.DecodeString(ethAddress)
	if e != nil {
		return errors.New(fmt.Sprintf("Invalid base64 string provided: [%s]", e))
	}

	if !encodingFormat.MatchString(string(decodedMemo)) {
		return errors.New(fmt.Sprintf("Memo is invalid or has invalid encoding format: [%s]", string(decodedMemo)))
	}

	return nil
}
