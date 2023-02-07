package blog

import (
	"errors"
	"net/http"
)

var ErrHttpUnAuthorized = errors.New("StatusUnauthorized")

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

type respErr struct {
	err  error
	code int
}

func (e *respErr) Code() int {
	return e.code
}

var respErrorMap = map[int]string{
	http.StatusUnauthorized:        "please log in first",
	http.StatusInternalServerError: "server internal error",
}

func (e *respErr) Error() string {
	var inner string
	if e.err != nil {
		inner = e.err.Error()
	}
	s, ok := respErrorMap[e.code]
	if ok {
		return s + ": " + inner
	}
	return inner
}

func (e *respErr) Unwrap() error {
	return e.err
}
