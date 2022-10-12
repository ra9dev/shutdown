package shutdown

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var signals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}

type (
	// CallbackFunc to Add
	CallbackFunc func(ctx context.Context)

	// GracefulShutdown handles all of your application Closable/Shutdownable dependencies
	GracefulShutdown struct {
		ctx       context.Context
		mu        *sync.RWMutex
		callbacks []CallbackFunc
		done      chan struct{}
	}
)

// NewGracefulShutdown constructor
func NewGracefulShutdown() *GracefulShutdown {
	osCTX, cancel := signal.NotifyContext(context.Background(), signals...)

	shutdown := &GracefulShutdown{
		ctx:       osCTX,
		mu:        new(sync.RWMutex),
		callbacks: make([]CallbackFunc, 0),
		done:      make(chan struct{}),
	}

	go func() {
		defer cancel()

		<-shutdown.ctx.Done()

		shutdown.ForceShutdown()
	}()

	return shutdown
}

// Add adds a callback to a GracefulShutdown instance
func (s *GracefulShutdown) Add(fn CallbackFunc) {
	s.mu.Lock()
	s.callbacks = append(s.callbacks, fn)
	s.mu.Unlock()
}

// ForceShutdown processes all shutdown callbacks concurrently in a limited time frame (Timeout)
func (s *GracefulShutdown) ForceShutdown() {
	defer close(s.done)

	ctx, cancel := context.WithTimeout(context.Background(), Timeout())
	defer cancel()

	s.mu.RLock()
	callbacks := s.callbacks
	s.mu.RUnlock()

	wg := new(sync.WaitGroup)
	wg.Add(len(callbacks))

	for _, callback := range callbacks {
		threadSafeCallback := callback

		go func() {
			defer wg.Done()

			threadSafeCallback(ctx)
		}()
	}

	wg.Wait()
}

// Context will be cancelled whenever OS termination signals are sent to your application process.
// Becomes handy when you want to propagate process context.
func (s *GracefulShutdown) Context() context.Context {
	return s.ctx
}

// Wait for it! Shutdown can be forced, cancelled by timeout, finished correctly.
// Required to use this method before application process termination.
func (s *GracefulShutdown) Wait() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, signals...)

	select {
	case <-s.done:
		return nil
	case <-time.After(Timeout()):
		return ErrTimeout
	case <-stop:
		return ErrForceStop
	}
}
