package shutdown

import (
	"context"
	"time"
)

var (
	globalShutdown = NewGracefulShutdown()

	timeout = time.Second * 5
)

// RegisterTimeout for a different shutdown timeout
func RegisterTimeout(duration time.Duration) {
	timeout = duration
}

// Timeout for shutdown
func Timeout() time.Duration {
	return timeout
}

// Add shutdown callback to a global GracefulShutdown
func Add(name string, fn CallbackFunc) {
	globalShutdown.Add(name, fn)
}

// Wait for a global GracefulShutdown, check GracefulShutdown.Wait
func Wait() error {
	return globalShutdown.Wait()
}

// Context of a global GracefulShutdown, check GracefulShutdown.Context
func Context() context.Context {
	return globalShutdown.Context()
}
