//go:build windows

package server

import (
	"os"
	"os/signal"

	"golang.org/x/sys/windows"
)

func (s *Server) waitForSignals() chan os.Signal {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, windows.SIGTERM, windows.SIGINT)
	return sigs
}
