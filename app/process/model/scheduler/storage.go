package scheduler

import "time"

type Storage struct {
	Executed bool
	Timer    *time.Timer
}
