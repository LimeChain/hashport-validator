package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	nanosecInSecond = 1000000000
)

func StringToTimestamp(timestamp string) (int64, error) {
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

	return seconds*nanosecInSecond + nano, nil
}

func TimestampToString(timestamp int64) string {
	seconds := timestamp / nanosecInSecond
	nano := timestamp % nanosecInSecond
	return fmt.Sprintf("%d.%d", seconds, nano)
}
