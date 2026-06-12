//go:build !windows

package tui

import (
	"os"
	"os/signal"
	"syscall"
)

func watchResize() (<-chan struct{}, func()) {
	out := make(chan struct{}, 1)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGWINCH)

	go func() {
		for range signals {
			select {
			case out <- struct{}{}:
			default:
			}
		}
	}()

	return out, func() {
		signal.Stop(signals)
		close(signals)
	}
}
