package daemon

import (
	"context"
	"fmt"
	"sync"

	"github.com/qeeq72/otus-final-project/internal/gatherer"
	"github.com/qeeq72/otus-final-project/internal/printer"
	"github.com/qeeq72/otus-final-project/internal/statistics/loadavg"
	"github.com/qeeq72/otus-final-project/internal/statistics/usagecpu"
)

type Daemon struct {
	gatherers []gatherer.IGatherer
	wg        sync.WaitGroup
	bufferCh  chan printer.IPrinter
	errCh     chan error
	stopCh    chan struct{}
}

func NewDaemon(cfg *gatherer.GathererConfig, depth, rate int) *Daemon {
	var gatherers []gatherer.IGatherer
	if /*cfg.LoadAvg.Enable || */ true {
		gatherers = append(gatherers, gatherer.NewGatherer(depth, rate, loadavg.NewLoadAverage()))
	}
	if /*cfg.UsageCPU.Enable || */ true {
		gatherers = append(gatherers, gatherer.NewGatherer(depth, rate, usagecpu.NewUsageCPU()))
	}
	/*if cfg.Disks.Enable {
		//
	}
	if cfg.TopTalkers.Enable {
		//
	}
	if cfg.Network.Enable {
		//
	}*/
	return &Daemon{
		gatherers: gatherers,
		bufferCh:  make(chan printer.IPrinter, len(gatherers)),
		errCh:     make(chan error),
		stopCh:    make(chan struct{}),
	}
}

func (d *Daemon) Run(ctx context.Context, out chan printer.IPrinter) (err error) {
	defer func() {
		close(d.bufferCh)
		close(d.errCh)
		close(d.stopCh)
	}()

	ctx, cancel := context.WithCancel(ctx)

	// Запускаем сборщиков статистики
	for i := range d.gatherers {
		d.wg.Add(1)
		go func(g gatherer.IGatherer) {
			defer d.wg.Done()
			err := g.Run(ctx, d.bufferCh)
			if err != nil {
				d.errCh <- err
			}
		}(d.gatherers[i])
	}

	// Запускаем горутину для контроля заполнения буфера
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case data := <-d.bufferCh:
				out <- data
			}
		}
	}()

	select {
	case <-ctx.Done():
	case <-d.stopCh:
	case e := <-d.errCh:
		err = fmt.Errorf("daemon running: %w", e)
	}
	cancel()

	d.wg.Wait()
	return
}

func (d *Daemon) Stop() {
	d.stopCh <- struct{}{}
}
