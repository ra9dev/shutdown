package shutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}

type (
	// CallbackFunc to Add
	CallbackFunc func(ctx context.Context)

	// GracefulShutdown handles all of your application Closable/Shutdownable dependencies
	GracefulShutdown struct {
		stop chan os.Signal
		done chan struct{}

		dependencyTree DependencyTree
	}
)

func newStopChan() chan os.Signal {
	stop := make(chan os.Signal, len(signals))

	signal.Notify(stop, signals...)

	return stop
}

// NewGracefulShutdown constructor
func NewGracefulShutdown() *GracefulShutdown {
	shutdown := &GracefulShutdown{
		stop: newStopChan(),
		done: make(chan struct{}),

		dependencyTree: NewDependencyTree(),
	}

	return shutdown
}

// Add adds a callback to a GracefulShutdown instance
func (s *GracefulShutdown) Add(name string, fn CallbackFunc) error {
	if err := s.dependencyTree.Insert(dependenciesRootKey, NewDependencyNode(name, fn)); err != nil {
		return err
	}

	return nil
}

// MustAdd adds a callback to a GracefulShutdown instance
func (s *GracefulShutdown) MustAdd(name string, fn CallbackFunc) {
	if err := s.dependencyTree.Insert(dependenciesRootKey, NewDependencyNode(name, fn)); err != nil {
		panic(err)
	}
}

// AddDependant adds a dependant callback to a GracefulShutdown instance
func (s *GracefulShutdown) AddDependant(dependsOn, name string, fn CallbackFunc) error {
	if err := s.dependencyTree.Insert(dependsOn, NewDependencyNode(name, fn)); err != nil {
		return err
	}

	return nil
}

// MustAddDependant adds a dependant callback to a GracefulShutdown instance
func (s *GracefulShutdown) MustAddDependant(dependsOn, name string, fn CallbackFunc) {
	if err := s.dependencyTree.Insert(dependsOn, NewDependencyNode(name, fn)); err != nil {
		panic(err)
	}
}

// ForceShutdown processes all shutdown callbacks concurrently in a limited time frame (Timeout)
func (s *GracefulShutdown) ForceShutdown() {
	close(s.stop)

	defer close(s.done)

	ctx, cancel := context.WithTimeout(context.Background(), Timeout())
	defer cancel()

	s.dependencyTree.Shutdown(ctx)
}

// Wait for it! Shutdown can be forced, cancelled by timeout, finished correctly.
// Required to use this method before application process termination.
func (s *GracefulShutdown) Wait() chan error {
	done := make(chan error, 1)

	go func() {
		done <- s.wait()
	}()

	return done
}

func (s *GracefulShutdown) wait() error {
	<-s.stop

	go func() {
		s.ForceShutdown()
	}()

	forceStop := newStopChan()

	select {
	case <-s.done:
		return nil
	case <-time.After(Timeout()):
		return nil
	case <-forceStop:
		return ErrForceStop
	}
}
