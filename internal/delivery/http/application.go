package http

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

// Application encapsulates the HTTP server lifecycle.
type Application struct {
	Server      *http.Server
	Multiplexer *http.ServeMux
}

// NewApplication creates a new Application instance with secure default timeouts.
func NewApplication(address string) *Application {
	var multiplexer *http.ServeMux = http.NewServeMux()

	// Default server configuration
	var server *http.Server = &http.Server{
		Addr:         address,
		Handler:      recoverPanic(multiplexer),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &Application{
		Server:      server,
		Multiplexer: multiplexer,
	}
}

func recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic details (in production, use structured logging).
				fmt.Printf("PANIC RECOVERED: %v\n", err)
				w.Header().Set("Connection", "close")
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// GetRouter returns the underlying HTTP request multiplexer.
func (application *Application) GetRouter() *http.ServeMux {
	return application.Multiplexer
}

// Run starts the HTTP server and blocks until a termination signal is received.
func (application *Application) Run() error {
	// Channel to listen for errors coming from the listener.
	var serverErrorChannel chan error = make(chan error, 1)

	// Start the server
	go func() {
		fmt.Printf("Starting server on %s\n", application.Server.Addr)
		var listenError error = application.Server.ListenAndServe()

		if listenError != nil && !errors.Is(listenError, http.ErrServerClosed) {
			serverErrorChannel <- listenError
		}
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	var shutdownSignalChannel chan os.Signal = make(chan os.Signal, 1)
	signal.Notify(shutdownSignalChannel, os.Interrupt, syscall.SIGTERM)

	// Block until we receive our signal.
	var serverError error
	var terminationSignal os.Signal

	select {
	case serverError = <-serverErrorChannel:
		return fmt.Errorf("server error: %w", serverError)

	case terminationSignal = <-shutdownSignalChannel:
		fmt.Printf("main: %v : Start shutdown\n", terminationSignal)

		// Give outstanding requests a deadline for completion.
		var shutdownContext context.Context
		var cancel context.CancelFunc
		shutdownContext, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Asking listener to shutdown and shed load.
		var shutdownError error = application.Server.Shutdown(shutdownContext)

		if shutdownError != nil {
			application.Server.Close()
			return fmt.Errorf("could not stop server gracefully: %w", shutdownError)
		}
	}

	return nil
}
