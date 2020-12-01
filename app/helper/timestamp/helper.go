package timestamp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	nanosInSecond = 1000000000
)

func FromString(timestamp string) (int64, error) {
	var err error
	stringTimestamp := strings.Split(timestamp, ".")

	seconds, err := strconv.ParseInt(stringTimestamp[0], 10, 64)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Could not parse the whole part of a timestamp: [%s] - [%s]", timestamp, err))
	}
	nano, err := strconv.ParseInt(stringTimestamp[1], 10, 64)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Could not parse the decimal part of a timestamp: [%s] - [%s]", timestamp, err))
	}

	return seconds*nanosInSecond + nano, nil
}

func ToString(timestamp int64) string {
	seconds := timestamp / nanosInSecond
	nano := timestamp % nanosInSecond
	return fmt.Sprintf("%d.%d", seconds, nano)
}

// ToNanos - converts timestamp in seconds to timestamp in nanos
func ToNanos(timestampInSec int64) int64 {
	return timestampInSec * nanosInSecond
}
