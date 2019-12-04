package errors

// AirshipError is the base error type
// used to create extended error types
// in other airshipctl packages.
type AirshipError struct {
	Message string
}

// Error function implments the golang
// error interface
func (ae *AirshipError) Error() string {
	return ae.Message
}

// ErrNotImplemented returned for not implemented features
type ErrNotImplemented struct {
}

func (e ErrNotImplemented) Error() string {
	return "Not implemented"
}

// ErrWrongConfig returned in case of incorrect configuration
type ErrWrongConfig struct {
}

func (e ErrWrongConfig) Error() string {
	return "Wrong configuration"
}

// ErrMissingConfig returned in case of missing configuration
type ErrMissingConfig struct {
}

func (e ErrMissingConfig) Error() string {
	return "Missing configuration"
}

// ErrConfigFailed returned in case of failure during configuration
type ErrConfigFailed struct {
}

func (e ErrConfigFailed) Error() string {
	return "Configuration failed to complete."
}
