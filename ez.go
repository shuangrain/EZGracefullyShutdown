package ez

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Hook func(ctx context.Context)
type ILogger interface {
	Println(v ...interface{})
}

var signalChannel = make(chan os.Signal)

func WaitGracefullyShutdown(logger ILogger, timeout time.Duration, hooks ...Hook) {
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

	go func() {
		wg.Wait()
		cancel()
		logger.Println("All the hooks executed done.")
	}()

	<-ctx.Done()
	logger.Println("The application will shutdown...")
}
