package scheduler

import "time"

type Storage struct {
	SubmitterAddress string
	Timer            *time.Timer
}
