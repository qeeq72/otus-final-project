// +build linux

package loadavg

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
	errPrefix = "loadavg"
)

type metrics struct {
	LoadAvg1  float32
	LoadAvg5  float32
	LoadAvg10 float32
}

type LoadAverage struct {
	buf  *bytes.Buffer
	pipe *cmdpipe.LinuxCommandPipe
}

func NewLoadAverage() *LoadAverage {
	buf := &bytes.Buffer{}
	pipe := cmdpipe.NewLinuxCommandPipe(nil, buf,
		cmdpipe.LinuxCommand{
			Name: "top",
			Args: []string{"-b", "-n1"},
		}, cmdpipe.LinuxCommand{
			Name: "head",
			Args: []string{"-n1"},
		})

	return &LoadAverage{
		buf:  buf,
		pipe: pipe,
	}
}

func (la *LoadAverage) GatherMetrics() (interface{}, error) {
	err := la.pipe.Execute()
	if err != nil {
		fmt.Println(err)
	}

	columns := strings.Fields(la.buf.String())
	la.buf.Reset()
	if len(columns) != 12 {
		return nil, fmt.Errorf("%s: %w", errPrefix, ErrSourceDataInvalid)
	}

	m := metrics{}

	for i := 9; i < 12; i++ {
		validStr := strings.TrimRight(columns[i], ",")
		validStr = strings.Replace(validStr, ",", ".", 1)
		val, err := strconv.ParseFloat(validStr, 32)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", errPrefix, err)
		}
		switch i {
		case 9:
			m.LoadAvg1 = float32(val)
		case 10:
			m.LoadAvg5 = float32(val)
		case 11:
			m.LoadAvg10 = float32(val)
		}
	}

	return m, nil
}

func (la *LoadAverage) Average(mList []interface{}, ts time.Time, period int) (printer.IPrinter, error) {
	count := len(mList)
	if count < 1 {
		return nil, fmt.Errorf("%s averaging: %w", errPrefix, ErrMetricsInvalid)
	}
	if count < 2 {
		if m, ok := mList[0].(metrics); ok {
			return &average{
				LoadAvg1:  m.LoadAvg1,
				LoadAvg5:  m.LoadAvg5,
				LoadAvg10: m.LoadAvg10,
				TimeBegin: ts,
				TimeEnd:   ts.Add(time.Duration(period) * time.Second),
				Period:    time.Second,
			}, nil
		}
	}
	avg := &average{}

	for i := range mList {
		if m, ok := mList[i].(metrics); ok {
			avg.LoadAvg1 += m.LoadAvg1
			avg.LoadAvg5 += m.LoadAvg5
			avg.LoadAvg10 += m.LoadAvg10
			continue
		}
		return nil, fmt.Errorf("%s averaging: %w", errPrefix, ErrMetricsInvalid)
	}

	avg.LoadAvg1 = avg.LoadAvg1 / float32(count)
	avg.LoadAvg5 = avg.LoadAvg5 / float32(count)
	avg.LoadAvg10 = avg.LoadAvg10 / float32(count)
	avg.TimeBegin = ts
	avg.TimeEnd = ts.Add(time.Duration(period) * time.Second)
	avg.Period = time.Duration(count) * time.Second

	return avg, nil
}
