/*
 * Copyright 2022 LimeChain Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package fee_policy

func getInterfaceValue(input interface{}, key string) (interface{}, bool) {
	mapObject, ok := input.(map[interface{}]interface{})
	if !ok {
		return nil, false
	}

	for eleKey, eleValue := range mapObject {
		if eleKey.(string) == key {
			return eleValue, true
		}
	}

	return nil, false
}

func networkAllowed(networks []uint64, networkId uint64) bool {
	if networks == nil {
		return true
	}

	for _, ele := range networks {
		if ele == networkId {
			return true
		}
	}

	return false
}
