package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/monitor"
	"github.com/user/portwatch/internal/notify"
)

func main() {
	configPath := flag.String("config", "config.json", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	notifier := notify.New(cfg)
	mon := monitor.New(cfg, notifier)

	log.Printf("portwatch starting, monitoring %d target(s)", len(cfg.Targets))

	ctx := contextWithSignal()
	mon.Start(ctx)

	<-mon.Done()
	log.Println("portwatch stopped")
}

func contextWithSignal() interface{ Done() <-chan struct{} } {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	ctx := &signalContext{done: make(chan struct{})}
	go func() {
		<-quit
		close(ctx.done)
	}()
	return ctx
}

type signalContext struct {
	done chan struct{}
}

func (s *signalContext) Done() <-chan struct{} {
	return s.done
}
