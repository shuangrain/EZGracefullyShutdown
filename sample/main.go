package main

import (
	"context"
	"fmt"
	"github.com/shuangrain/EZGracefullyShutdown"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	waitTimeForShutdown := time.Second * 5
	waitTimeForRequest := time.Second * 10
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(waitTimeForRequest)
		_, _ = fmt.Fprint(w, "ok")
	})
	server := http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	go func() { _ = server.ListenAndServe() }()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	ez.WaitGracefullyShutdown(logger, waitTimeForShutdown, func(ctx context.Context) {
		_ = server.Shutdown(ctx)
	})
}
