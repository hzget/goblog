package blog

import (
	"errors"
	"net/http"
)

var ErrHttpUnAuthorized = errors.New("StatusUnauthorized")

var ErrCacheTokenUnmatch = errors.New("Cache token unmatch")
var ErrCredentialFailed = errors.New("fail to validate credential")

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
	msg  string
}

func NewRespErr(err error, code int, msg ...string) error {
	r := &respErr{err: err, code: code}
	for _, v := range msg {
		r.msg += v
	}
	return r
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
	if e.msg != "" {
		return e.msg + " " + inner
	}
	s, ok := respErrorMap[e.code]
	if ok {
		return s + " " + inner
	}
	return inner
}

func (e *respErr) Unwrap() error {
	return e.err
}
