package urlshortenererror

import (
	"errors"
	"fmt"
)

// Var ...
var (
	// ErrDBConnection ...
	ErrDBConnection = errors.New("database connection error")
	// ErrDBQuery ...
	ErrDBQuery = errors.New("database query error")
	// ErrNotFound ...
	ErrNotFound = errors.New("resource not found")
	// ErrDuplicate ...
	ErrDuplicate = errors.New("duplicate entry")
	// ErrInvalidInput ...
	ErrInvalidInput = errors.New("invalid input")
	// ErrInvalidDBPort ...
	ErrInvalidDBPort = errors.New("invalid database port")
	// ErrInvalidURL ...
	ErrInvalidURL = errors.New("invalid URL format")
	// ErrServerError ...
	ErrServerError = errors.New("internal server error")
)

// WebError struct to hold the error details.
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

// Unwrap returns the inner error.
func (e *WebError) Unwrap() error {
	return e.InnerErr
}

// Is checks if the error is a WebError.
func (*WebError) Is(target error) bool {
	_, ok := target.(*WebError)

	return ok
}

// New creates a new WebError.
func New(errType error, innerErr error, message string, code int) *WebError {
	err := &WebError{
		ErrType:  errType,
		InnerErr: innerErr,
		Message:  message,
		Code:     code,
	}

	return err
}

// Wrap wraps an error with a message and code, and returns a WebError.
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
