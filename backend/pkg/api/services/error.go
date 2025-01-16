package services

import (
	"bin-vul-inspector/pkg/api/v1/dto"
)

type Error struct {
	Code    int
	Message string
	Extra   []error
}

func (e *Error) Error() string {
	return e.Message
}

func NewError(code int, message string, errs ...error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Extra:   errs,
	}
}

func NewErrorWithStatus(code int, errs ...error) *Error {
	return &Error{
		Code:    code,
		Message: dto.StatusText(code),
		Extra:   errs,
	}
}
