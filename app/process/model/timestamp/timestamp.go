package timestamp

import (
	"fmt"
)

type Timestamp struct {
	Seconds     int64
	NanoSeconds int64
}

func NewTimestamp(seconds int64, nanoSeconds int64) *Timestamp {
	return &Timestamp{
		Seconds:     seconds,
		NanoSeconds: nanoSeconds,
	}
}

func (t Timestamp) IsValid() bool {
	return t.Seconds > 0 || t.NanoSeconds > 0
}

func (t Timestamp) ToString() string {
	return fmt.Sprintf("%d.%d", t.Seconds, t.NanoSeconds)
}
