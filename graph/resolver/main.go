package resolver

type InternalError struct {
	err error
}

func (e InternalError) Error() string {
	return e.err.Error()
}

func (e InternalError) Unwrap() error {
	return e.err
}
