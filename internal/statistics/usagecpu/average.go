package usagecpu

import (
	"fmt"
	"time"
)

type average struct {
	User   float32
	Nice   float32
	System float32
	Idle   float32

	TimeBegin time.Time
	TimeEnd   time.Time
	Period    time.Duration
}

func (a *average) String() string {
	return fmt.Sprintf("CPU usage: User: %.2f, Nice: %.2f, System: %.2f, Idle: %.2f - Begin: %s, End: %s, Duration: %s",
		a.User,
		a.Nice,
		a.System,
		a.Idle,
		a.TimeBegin.Format(time.RFC3339),
		a.TimeEnd.Format(time.RFC3339),
		a.Period,
	)
}
