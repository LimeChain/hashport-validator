package timestamp

import (
	"fmt"
)

type Timestamp struct {
	Whole int64
	Dec   int64
}

func NewTimestamp(whole int64, dec int64) *Timestamp {
	return &Timestamp{
		Whole: whole,
		Dec:   dec,
	}
}

func (t Timestamp) IsValid() bool {
	return t.Whole > 0 || t.Dec > 0
}

func (t Timestamp) ToString() string {
	return fmt.Sprintf("%d.%d", t.Whole, t.Dec)
}
