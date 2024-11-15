package urlshortenererror

import (
	"errors"
	"fmt"
)

var (
	ErrDBConnection = errors.New("database connection error")
	ErrDBQuery      = errors.New("database query error")
	ErrNotFound     = errors.New("resource not found")
	ErrDuplicate    = errors.New("duplicate entry")
)

type WebError struct {
	ErrType  error
	InnerErr error
	Message  string
	Code     int
}

func (e *WebError) Error() string {
	if e.InnerErr != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.InnerErr.Error())
	}
	return e.Message
}

func (e *WebError) Unwrap() error {
	return e.InnerErr
}

func (e *WebError) Is(target error) bool {
	_, ok := target.(*WebError)
	return ok
}

func New(errType error, innerErr error, message string, code int) *WebError {
	err := &WebError{
		ErrType:  errType,
		InnerErr: innerErr,
		Message:  message,
		Code:     code,
	}
	return err
}

func Wrap(err error, message string, code int, errType error) *WebError {
	if errors.Is(err, &WebError{}) {
		errType = nil
	}
	return &WebError{
		ErrType:  errType,
		InnerErr: err,
		Message:  message,
		Code:     code,
	}
}
