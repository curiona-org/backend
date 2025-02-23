//go:build linux || dragonfly || freebsd || netbsd || openbsd || darwin

package server

import (
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

func (s *Server) waitForSignals() chan os.Signal {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, unix.SIGTERM, unix.SIGINT)
	return sigs
}
