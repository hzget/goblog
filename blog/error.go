package blog

type limitErr struct {
	err error
	msg string
}

func (e *limitErr) Error() string {
	return e.msg
}

func (e *limitErr) Unwrap() error {
	return e.err
}
