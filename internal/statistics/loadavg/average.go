package loadavg

import (
	"fmt"
	"time"
)

type average struct {
	LoadAvg1  float32
	LoadAvg5  float32
	LoadAvg10 float32

	TimeBegin time.Time
	TimeEnd   time.Time
	Period    time.Duration
}

func (a *average) String() string {
	return fmt.Sprintf("Load average: 1m: %.2f, 5m: %.2f, 10m: %.2f - Begin: %s, End: %s, Duration: %s",
		a.LoadAvg1,
		a.LoadAvg5,
		a.LoadAvg10,
		a.TimeBegin.Format(time.RFC3339),
		a.TimeEnd.Format(time.RFC3339),
		a.Period,
	)
}
