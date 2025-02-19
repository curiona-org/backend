package server_test

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/roadmap-thesis/backend/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestServer_New(t *testing.T) {
	t.Parallel()
	port := "8080"
	srv := server.New(port)
	assert.NotNil(t, srv)
	assert.Equal(t, port, srv.Port())
	assert.IsType(t, &echo.Echo{}, srv.Echo)
}

func TestServer_Listen(t *testing.T) {
	t.Parallel()
	port := "8080"
	srv := server.New(port)

	exitSignal := srv.Listen()
	assert.NotNil(t, exitSignal)

	// Simulate sending an interrupt signal
	go func() {
		time.Sleep(1 * time.Second)
		exitSignal <- os.Interrupt
	}()

	sig := <-exitSignal
	assert.Equal(t, os.Interrupt, sig)
}

func TestServer_Shutdown(t *testing.T) {
	t.Parallel()
	port := "8081"
	srv := server.New(port)

	ctx := context.Background()
	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		time.Sleep(1 * time.Second)
		exitSignal <- os.Interrupt
	}()

	sig := <-exitSignal
	assert.Equal(t, os.Interrupt, sig)

	go srv.Listen()
	time.Sleep(1 * time.Second) // Give the server time to start

	srv.Shutdown(ctx, sig)
}
