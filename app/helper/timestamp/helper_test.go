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

package timestamp

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

const(
	validTimestamp = "1598924675.82525000"
	timestampInt64 int64 = 1598924675082525000
	nonValidNanos = "1598924675.82525000423423541521512"
	nonValidSeconds = "1598924675423423541521512.82525000"
)

func Test_Validate(t *testing.T)  {
	timestamp, err := FromString(validTimestamp)
	assert.Equal(t, timestampInt64,  timestamp)
	assert.Nil(t, err)
}

func Test_NonValidNanos(t *testing.T)  {
	_, err := FromString(nonValidNanos)
	assert.Error(t, err, errors.New("invalid timestamp nanos provided"))
	assert.NotNil(t, err)
}

func Test_NonValidSeconds(t *testing.T)  {
	_, err := FromString(nonValidSeconds)
	assert.Error(t, err, errors.New("invalid timestamp seconds provided"))
	assert.NotNil(t, err)
}

func Test_String(t *testing.T)  {
	res := String(timestampInt64)
	assert.Equal(t, validTimestamp, res)
}

func Test_ToHumanReadable(t *testing.T)  {
	res := ToHumanReadable(timestampInt64)
	expectedDate := "2020-09-01T04:44:35.008554496+03:00"
	assert.Equal(t, expectedDate, res)
}
