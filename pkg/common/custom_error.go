package common

import "net/http"

var ErrInternalServerError = NewCustomError(http.StatusInternalServerError, "internal server error")
var ErrPostNotFound = NewCustomError(http.StatusNotFound, "post not found")
var ErrPostVersionNotFound = NewCustomError(http.StatusNotFound, "post version not found")

type CustomError struct {
	StatusCode int
	Message    string
}

func (e *CustomError) Error() string {
	return e.Message
}

func NewCustomError(statusCode int, message string) *CustomError {
	return &CustomError{StatusCode: statusCode, Message: message}
}
