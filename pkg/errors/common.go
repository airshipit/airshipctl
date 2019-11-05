package errors

// ErrNotImplemented returned for not implemented features
type ErrNotImplemented struct {
}

func (e ErrNotImplemented) Error() string {
	return "Error. Not implemented"
}
