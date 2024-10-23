package common

import "net/http"

var ErrInternalServerError = NewCustomError(http.StatusInternalServerError, "An internal server error occurred")
var ErrPostNotFound = NewCustomError(http.StatusNotFound, "The requested post was not found")
var ErrPostVersionNotFound = NewCustomError(http.StatusNotFound, "The requested post version was not found")

type CustomError struct {
	StatusCode int
	Message    string
}

func (e *CustomError) Error() string {
	return e.Message
}

func NewCustomError(statusCode int, message string) *CustomError {
	if message == "" {
		message = http.StatusText(statusCode)
	}
	if statusCode < 100 || statusCode > 599 {
		statusCode = http.StatusInternalServerError
	}
	return &CustomError{StatusCode: statusCode, Message: message}
}
