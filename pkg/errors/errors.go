package errors

import (
	stderrors "errors"
	"fmt"
)

// ErrRecordNotFound is a sentinel returned by repository when a record does not exist.
// Services detect this with errors.Is and convert it to NewNotFound.
var ErrRecordNotFound = stderrors.New("record not found")

type ErrorCode string

const (
	ErrNotFound         ErrorCode = "NOT_FOUND"
	ErrInvalidParameter ErrorCode = "INVALID_PARAMETER"
	ErrAuthZ            ErrorCode = "AUTHORIZATION"
	ErrAuthN            ErrorCode = "AUTHENTICATION"
	ErrInternal         ErrorCode = "INTERNAL"
)

type AppError struct {
	Code    ErrorCode
	Message string
	Details map[string]string
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// WithDetail attaches debugging context to the error. Details are NOT exposed to clients.
func (e *AppError) WithDetail(key, value string) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]string)
	}
	e.Details[key] = value
	return e
}

func NewNotFound(message string) *AppError {
	return &AppError{Code: ErrNotFound, Message: message}
}

func NewInvalidParameter(message string) *AppError {
	return &AppError{Code: ErrInvalidParameter, Message: message}
}

func NewAuthZ(message string) *AppError {
	return &AppError{Code: ErrAuthZ, Message: message}
}

func NewAuthN(message string) *AppError {
	return &AppError{Code: ErrAuthN, Message: message}
}

func NewInternal(message string) *AppError {
	return &AppError{Code: ErrInternal, Message: message}
}
