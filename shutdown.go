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
		done chan struct{}

		dependencyTree DependencyTree
	}
)

// NewGracefulShutdown constructor
func NewGracefulShutdown() *GracefulShutdown {
	osCTX, cancel := signal.NotifyContext(context.Background(), signals...)

	shutdown := &GracefulShutdown{
		done: make(chan struct{}),

		dependencyTree: NewDependencyTree(),
	}

	go func() {
		defer cancel()

		<-osCTX.Done()

		shutdown.ForceShutdown()
	}()

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
	defer close(s.done)

	ctx, cancel := context.WithTimeout(context.Background(), Timeout())
	defer cancel()

	s.dependencyTree.Shutdown(ctx)
}

// Wait for it! Shutdown can be forced, cancelled by timeout, finished correctly.
// Required to use this method before application process termination.
func (s *GracefulShutdown) Wait() error {
	<-s.done

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, signals...)

	select {
	case <-s.done:
		return nil
	case <-time.After(Timeout()):
		return nil
	case <-stop:
		return ErrForceStop
	}
}
