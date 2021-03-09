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

package mirror_node

import (
	"fmt"
)

type (
	ErrorMessage struct {
		Message string `json:"message"`
	}
	Status struct {
		Messages []ErrorMessage
	}
)

// String converts ErrorMessage struct to human readable string
func (m *ErrorMessage) String() string {
	return fmt.Sprintf("message: %s", m.Message)
}

// IsNotFound returns true/false whether the message is equal to "not found" or not
func (m *ErrorMessage) IsNotFound() bool {
	return m.Message == "Not found"
}

// String converts the Status struct to human readable string
func (s *Status) String() string {
	r := "["
	for i, m := range s.Messages {
		r += m.String()
		if i != len(s.Messages)-1 {
			r += ", "
		}
	}
	r += "]"
	return r
}
