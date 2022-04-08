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

package error

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	errorMsg       *ErrorMessage
	notFoundErrMsg *ErrorMessage
	status         *Status
	msg            = "Some message."
)

func Test_ErrorMessage_String(t *testing.T) {
	setup()
	expectedMsg := fmt.Sprintf(errMsgTemplate, msg)

	actualMsg := errorMsg.String()

	assert.Equal(t, expectedMsg, actualMsg)
}

func Test_ErrorMessage_IsNotFound(t *testing.T) {
	setup()

	assert.Equal(t, true, notFoundErrMsg.IsNotFound())
	assert.Equal(t, false, errorMsg.IsNotFound())
}

func Test_Status_String(t *testing.T) {
	setup()
	expectedMsg := statusStartSymbol + errorMsg.String() + statusMsgSeparator + notFoundErrMsg.String() + statusEndSymbol

	actualMsg := status.String()

	assert.Equal(t, expectedMsg, actualMsg)
}

func setup() {
	errorMsg = &ErrorMessage{
		Message: msg,
	}

	notFoundErrMsg = &ErrorMessage{
		Message: NotFoundMsg,
	}

	status = &Status{
		Messages: []ErrorMessage{
			{Message: msg},
			{Message: NotFoundMsg},
		},
	}
}
