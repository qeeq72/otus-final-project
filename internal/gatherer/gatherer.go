package gatherer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/qeeq72/otus-final-project/internal/printer"
	"github.com/qeeq72/otus-final-project/utils/buffer"
)

type IGatherer interface {
	Run(context.Context, chan printer.IPrinter) error
}

type IStatsHandler interface {
	GatherMetrics() (interface{}, error)
	Average([]interface{}, time.Time, int) (printer.IPrinter, error)
}

type GathererConfig struct {
	LoadAvg struct {
		Enable bool `yaml:"enable"`
	} `yaml:"loadavg"`
	UsageCPU struct {
		Enable bool `yaml:"enable"`
	} `yaml:"usage_cpu"`
	Disks struct {
		Enable bool `yaml:"enable"`
	} `yaml:"disks"`
	TopTalkers struct {
		Enable bool `yaml:"enable"`
	} `yaml:"top_talkers"`
	Network struct {
		Enable bool `yaml:"enable"`
	} `yaml:"network"`
}

type Gatherer struct {
	buffer  buffer.IBufferGetSetter
	handler IStatsHandler
	errCh   chan error
	depth   int
	rate    int
}

func NewGatherer(depth, rate int, handler IStatsHandler) *Gatherer {
	return &Gatherer{
		buffer:  buffer.NewBuffer(depth),
		handler: handler,
		errCh:   make(chan error),
		depth:   depth,
		rate:    rate,
	}
}

var ErrBufferInvalidValue = errors.New("buffer invalid value")

func (c *Gatherer) Run(ctx context.Context, out chan printer.IPrinter) error {
	var count int
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case err := <-c.errCh:
			return fmt.Errorf("gathering: %w", err)
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			value, err := c.handler.GatherMetrics()
			if err != nil {
				return fmt.Errorf("gathering: %w", err)
			}
			c.buffer.Set(time.Now(), value)
			count++
			if count == c.depth {
				count -= c.rate
				data, ok := c.buffer.GetPeriod(time.Now().Add(time.Duration(1-c.depth)*time.Second), c.depth)
				if !ok {
					return fmt.Errorf("gathering: %w", ErrBufferInvalidValue)
				}
				go func(data []interface{}) {
					avg, err := c.handler.Average(data, time.Now().Add(time.Duration(-c.depth)*time.Second), c.depth)
					if err != nil {
						c.errCh <- err
						return
					}
					out <- avg
				}(data)
			}
		}
	}
}
