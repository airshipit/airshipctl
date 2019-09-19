package error

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
