package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)


// Application encapsulates the lifecycle and routing engine of the HTTP server.
type Application struct {
	server      *http.Server
	multiplexer *http.ServeMux
}


// recoverPanic is a global middleware that ensures any goroutine panic does not crash the process.
func recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		defer func() {
			var recoveredValue any = recover()

			if recoveredValue != nil {
				fmt.Printf("PANIC RECOVERED: %v\n", recoveredValue)
				responseWriter.Header().Set("Connection", "close")
				http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(responseWriter, request)
	})
}


// NewApplication initializes and returns a new Application instance.
func NewApplication() *Application {
	var multiplexer *http.ServeMux = http.NewServeMux()

	return &Application{
		multiplexer: multiplexer,
	}
}


// GetRouter returns the underlying multiplexer for external route registration.
func (application *Application) GetRouter() *http.ServeMux {
	return application.multiplexer
}


// Run starts the HTTP server and blocks until a system termination signal is received.
func (application *Application) Run(address string) error {
	application.server = &http.Server{
		Addr: address,

		// The routing middleware with panic recovery is mounted here.
		Handler:      recoverPanic(application.multiplexer),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	var errorChannel chan error = make(chan error, 1)

	// Start the server in the background.
	go func() {
		fmt.Printf("Starting server on %s...\n", application.server.Addr)

		var executionError error = application.server.ListenAndServe()

		if executionError != nil && !errors.Is(executionError, http.ErrServerClosed) {
			errorChannel <- executionError
		}
	}()

	// Listen for system signals.
	var signalChannel chan os.Signal = make(chan os.Signal, 1)

	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	var terminationError error
	var receivedSignal os.Signal

	select {
	case terminationError = <-errorChannel:
		return fmt.Errorf("Server failed to start: %w", terminationError)

	case receivedSignal = <-signalChannel:
		fmt.Printf("\nShutdown signal (%v) received, initiating graceful shutdown...\n", receivedSignal)

		// Provide a 15-second buffer for pending requests to complete.
		var contextWithTimeout context.Context
		var cancel context.CancelFunc
		contextWithTimeout, cancel = context.WithTimeout(context.Background(), 15*time.Second)

		defer cancel()

		var shutdownError error = application.server.Shutdown(contextWithTimeout)

		if shutdownError != nil {
			application.server.Close() // Forcefully close all connections if graceful shutdown fails.

			return fmt.Errorf("Graceful shutdown failed: %w", shutdownError)
		}

		fmt.Println("Server stopped cleanly.")

		return nil
	}
}
