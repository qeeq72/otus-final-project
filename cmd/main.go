package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/qeeq72/otus-final-project/internal/daemon"
	"github.com/qeeq72/otus-final-project/internal/gatherer"
	"github.com/qeeq72/otus-final-project/internal/printer"
	"github.com/qeeq72/otus-final-project/utils/file"
)

func main() {
	flag.Parse()

	cfg := &gatherer.GathererConfig{}
	err := file.ReadYamlFile("../../otus-final-project/configs/config.yml", cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	fmt.Println(time.Now())
	d := daemon.NewDaemon(nil, 5, 2)

	wg := sync.WaitGroup{}

	data := make(chan printer.IPrinter, 2)
	done := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := d.Run(ctx, data)
		if err != nil {
			fmt.Println(err)
		}
		close(data)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for d := range data {
			fmt.Println(d)
		}
		close(done)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	select {
	case <-stop:
	case <-done:
	}
	//<-stop
	cancel()
	wg.Wait()
}
