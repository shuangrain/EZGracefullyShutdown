# EZGracefullyShutdown
This project will easy to hook shutdown events to avoid losing application states and keep data correct.

## How to use
1. Use goroutine to running application (ex: web server, worker...).
2. Import this project and use `WaitGracefullyShutdown` on the `main.go` function of the last line.

## Example
See details in the file: `sample/main.go`
````
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
````