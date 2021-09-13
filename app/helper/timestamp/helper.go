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
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	nanosInSecond = 1000000000
)

// FromString parses a string in the format `{seconds}.{nanos}` into int64 timestamp
func FromString(timestamp string) (int64, error) {
	var err error
	stringTimestamp := strings.Split(timestamp, ".")

	if len(stringTimestamp) < 2 {
		return 0, errors.New("invalid timestamp provided")
	}

	seconds, err := strconv.ParseInt(stringTimestamp[0], 10, 64)
	if err != nil {
		return 0, errors.New("invalid timestamp seconds provided")
	}
	nano, err := strconv.ParseInt(stringTimestamp[1], 10, 64)
	if err != nil {
		return 0, errors.New("invalid timestamp nanos provided")
	}

	return seconds*nanosInSecond + nano, nil
}

// String parses int64 timestamp into `{seconds}.{nanos}` string
func String(timestamp int64) string {
	seconds := timestamp / nanosInSecond
	nano := timestamp % nanosInSecond
	return fmt.Sprintf("%d.%d", seconds, nano)
}

// ToHumanReadable converts the timestamp into human readable string
func ToHumanReadable(timestampNanos int64) string {
	parsed := time.Unix(timestampNanos/nanosInSecond, timestampNanos&nanosInSecond)
	return parsed.Format(time.RFC3339Nano)
}
