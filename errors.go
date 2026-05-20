package qac

// ErrorKind categorises a QacError so consumers can distinguish failure modes.
type ErrorKind int

const (
	// KindAssertionFailure means an assertion's condition was not met.
	KindAssertionFailure ErrorKind = iota
	// KindInfrastructure means a system-level operation failed (file I/O, path resolution).
	KindInfrastructure
	// KindConfiguration means the test plan itself is invalid (YAML, unknown fields, wrong field usage).
	KindConfiguration
)

// QacError is a structured error that carries a kind for programmatic classification.
// It wraps the underlying cause so errors.As / errors.Is work through the chain.
type QacError struct {
	Kind  ErrorKind
	Cause error
	msg   string
}

func (e *QacError) Error() string { return e.msg }
func (e *QacError) Unwrap() error { return e.Cause }

// asInfraError wraps an existing error as KindInfrastructure, preserving its message and chain.
func asInfraError(err error) *QacError {
	return &QacError{Kind: KindInfrastructure, Cause: err, msg: err.Error()}
}

// asConfigError wraps an existing error as KindConfiguration, preserving its message and chain.
func asConfigError(err error) *QacError {
	return &QacError{Kind: KindConfiguration, Cause: err, msg: err.Error()}
}
