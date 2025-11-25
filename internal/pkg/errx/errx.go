package errx

import "fmt"

type Error struct {
	code  int
	msg   string
	cause error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.cause == nil {
		return e.msg
	}
	if e.msg == "" {
		return e.cause.Error()
	}
	return e.msg + ": " + e.cause.Error()
}

func (e *Error) Unwrap() error {
	return e.cause
}

func (e *Error) Code() int {
	if e == nil {
		return 0
	}
	return e.code
}

func New(code int, msg string) error {
	return &Error{code: code, msg: msg}
}

func Newf(code int, format string, args ...any) error {
	return &Error{code: code, msg: fmt.Sprintf(format, args...)}
}

func Warp(code int, err error, msg string) error {
	if err == nil {
		return nil
	}
	return &Error{code: code, msg: msg, cause: err}
}

func Warpf(code int, err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &Error{code: code, msg: fmt.Sprintf(format, args...), cause: err}
}
