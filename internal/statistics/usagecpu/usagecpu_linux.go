// +build linux

package usagecpu

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/qeeq72/otus-final-project/internal/printer"
	"github.com/qeeq72/otus-final-project/utils/cmdpipe"
)

var (
	ErrSourceDataInvalid error = errors.New("invalid source data")
	ErrMetricsInvalid    error = errors.New("invalid metrics")
)

const (
	errPrefix = "usage cpu"
)

type metrics struct {
	User   float32
	Nice   float32
	System float32
	Idle   float32
}

type UsageCPU struct {
	buf  *bytes.Buffer
	pipe *cmdpipe.LinuxCommandPipe
}

func NewUsageCPU() *UsageCPU {
	buf := &bytes.Buffer{}
	pipe := cmdpipe.NewLinuxCommandPipe(nil, buf,
		cmdpipe.LinuxCommand{
			Name: "top",
			Args: []string{"-b", "-n1"},
		}, cmdpipe.LinuxCommand{
			Name: "head",
			Args: []string{"-n3"},
		})

	return &UsageCPU{
		buf:  buf,
		pipe: pipe,
	}
}

func (uc *UsageCPU) GatherMetrics() (interface{}, error) {
	err := uc.pipe.Execute()
	if err != nil {
		fmt.Println(err)
	}

	columns := strings.Fields(uc.buf.String())
	uc.buf.Reset()
	if len(columns) != 40 {
		return nil, fmt.Errorf("%s: %w", errPrefix, ErrSourceDataInvalid)
	}

	m := metrics{}

	for i := 24; i < 31; i = i + 2 {
		validStr := strings.TrimRight(columns[i], ",")
		validStr = strings.Replace(validStr, ",", ".", 1)
		val, err := strconv.ParseFloat(validStr, 32)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", errPrefix, err)
		}
		switch i {
		case 24:
			m.User = float32(val)
		case 26:
			m.System = float32(val)
		case 28:
			m.Nice = float32(val)
		case 30:
			m.Idle = float32(val)
		}
	}

	return m, nil
}

func (uc *UsageCPU) Average(mList []interface{}, ts time.Time, period int) (printer.IPrinter, error) {
	count := len(mList)
	if count < 1 {
		return nil, fmt.Errorf("%s averaging: %w", errPrefix, ErrMetricsInvalid)
	}
	if count < 2 {
		if m, ok := mList[0].(metrics); ok {
			return &average{
				User:      m.User,
				Nice:      m.Nice,
				System:    m.System,
				Idle:      m.Idle,
				TimeBegin: ts,
				TimeEnd:   ts.Add(time.Duration(period) * time.Second),
				Period:    time.Second,
			}, nil
		}
	}
	avg := &average{}

	for i := range mList {
		if m, ok := mList[i].(metrics); ok {
			avg.User += m.User
			avg.Nice += m.Nice
			avg.System += m.System
			avg.Idle += m.Idle
			continue
		}
		return nil, fmt.Errorf("%s averaging: %w", errPrefix, ErrMetricsInvalid)
	}

	avg.User = avg.User / float32(count)
	avg.Nice = avg.Nice / float32(count)
	avg.System = avg.System / float32(count)
	avg.Idle = avg.Idle / float32(count)
	avg.TimeBegin = ts
	avg.TimeEnd = ts.Add(time.Duration(period) * time.Second)
	avg.Period = time.Duration(count) * time.Second

	return avg, nil
}
