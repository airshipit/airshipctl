package errors

// AirshipError is the base error type
// used to create extended error types
// in other airshipctl packages.
type AirshipError struct {
	Message string
}

// Error function implements the golang
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
