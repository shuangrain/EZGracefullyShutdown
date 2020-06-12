package ez_gracefully_shutdown

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Hook func(ctx context.Context)

var signalChannel = make(chan os.Signal)

func WaitGracefullyShutdown(logger *log.Logger, timeout time.Duration, hooks ...Hook) {
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	logger.Println("Waiting system shutdown event...")
	signalName := (<-signalChannel).String()

	logger.Println("Received signal:", signalName, "the hooks will trigger.")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	var wg sync.WaitGroup
	for _, hook := range hooks {
		wg.Add(1)
		go func(hook Hook) {
			hook(ctx)
			wg.Done()
		}(hook)
	}

	isGraceful := false
	go func() {
		wg.Wait()
		cancel()
		logger.Println("All the hooks executed done.")
		isGraceful = true
	}()

	<-ctx.Done()
	if !isGraceful {
		logger.Println("One of the hooks had timeout or unhandled context.")
	}
	logger.Println("The application will shutdown...")
}
