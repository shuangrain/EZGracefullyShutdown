package ez_gracefully_shutdown

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestWaitGracefullyShutdown_Success(t *testing.T) {
	waitTimeForShutdown := time.Second * 10
	waitTimeForRequest := time.Second * 5

	var mutex sync.Mutex
	var urlChannel = make(chan string)
	mutex.Lock()
	go startApplication(waitTimeForShutdown, waitTimeForRequest, urlChannel, &mutex)
	go sendShutdownSignal()
	go mockBrowserRequest(t, <-urlChannel)
	mutex.Lock()
}

func TestWaitGracefullyShutdown_Timeout(t *testing.T) {
	waitTimeForShutdown := time.Second * 5
	waitTimeForRequest := time.Second * 10

	var mutex sync.Mutex
	var urlChannel = make(chan string)
	mutex.Lock()
	go startApplication(waitTimeForShutdown, waitTimeForRequest, urlChannel, &mutex)
	go sendShutdownSignal()
	go mockBrowserRequest(t, <-urlChannel)
	mutex.Lock()
}

func TestWaitGracefullyShutdown_UnHandledContext(t *testing.T) {
	waitTimeForShutdown := time.Second * 5
	waitTimeForRequest := time.Second * 10

	var mutex sync.Mutex
	mutex.Lock()
	go func() {
		logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
		WaitGracefullyShutdown(logger, waitTimeForShutdown, func(ctx context.Context) {
			time.Sleep(waitTimeForRequest)
		})
		mutex.Unlock()
	}()
	go sendShutdownSignal()
	mutex.Lock()
}

func sendShutdownSignal() {
	time.Sleep(time.Second * 2)
	signalChannel <- syscall.SIGTERM
}

func startApplication(waitTimeForShutdown time.Duration, waitTimeForRequest time.Duration, urlChannel chan string, mutex *sync.Mutex) {
	var server http.Server
	go func() {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(waitTimeForRequest)
			_, _ = fmt.Fprint(w, "ok")
		})

		testServer := httptest.NewServer(handler)
		server = http.Server{
			Addr:    testServer.Listener.Addr().String(),
			Handler: handler,
		}
		testServer.Close()
		go func() {
			urlChannel <- testServer.URL
			_ = server.ListenAndServe()
		}()
	}()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	WaitGracefullyShutdown(logger, waitTimeForShutdown, func(ctx context.Context) {
		_ = server.Shutdown(ctx)
	})
	mutex.Unlock()
}

func mockBrowserRequest(t *testing.T, url string) {
	response, err := http.Get(url)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	assert.Equal(t, "ok", string(body))
}
