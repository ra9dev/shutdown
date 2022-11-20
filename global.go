package shutdown

import (
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
func Add(name string, fn CallbackFunc) error {
	return globalShutdown.Add(name, fn)
}

// MustAdd shutdown callback to a global GracefulShutdown
func MustAdd(name string, fn CallbackFunc) {
	globalShutdown.MustAdd(name, fn)
}

// AddDependant shutdown callback to a global GracefulShutdown
func AddDependant(dependsOn, name string, fn CallbackFunc) error {
	return globalShutdown.AddDependant(dependsOn, name, fn)
}

// MustAddDependant shutdown callback to a global GracefulShutdown
func MustAddDependant(dependsOn, name string, fn CallbackFunc) {
	globalShutdown.MustAddDependant(dependsOn, name, fn)
}

// Wait for a global GracefulShutdown, check GracefulShutdown.Wait
func Wait() chan error {
	return globalShutdown.Wait()
}
