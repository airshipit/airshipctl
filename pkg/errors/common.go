package errors

// ErrNotImplemented returned for not implemented features
type ErrNotImplemented struct {
}

func (e ErrNotImplemented) Error() string {
	return "Error. Not implemented"
}

// ErrWrongConfig returned in case of incorrect configuration
type ErrWrongConfig struct {
}

func (e ErrWrongConfig) Error() string {
	return "Error. Wrong configuration"
}
