package shutdown

import "errors"

var (
	ErrForceStop          = errors.New("shutdown force stopped")
	ErrTimeout            = errors.New("shutdown timed out")
	ErrCyclicDependencies = errors.New("cyclic dependencies are not allowed")
)
