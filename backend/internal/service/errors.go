package service

import "net/http"

type ClientError interface {
	error
	StatusCode() int
}

type appError struct {
	status  int
	message string
}

func (e *appError) Error() string {
	return e.message
}

func (e *appError) StatusCode() int {
	return e.status
}

func NewErrValidacion(message string) error {
	return &appError{status: http.StatusBadRequest, message: message}
}

func NewErrNoAutorizado(message string) error {
	return &appError{status: http.StatusUnauthorized, message: message}
}

func NewErrProhibido(message string) error {
	return &appError{status: http.StatusForbidden, message: message}
}

func NewErrNoEncontrado(message string) error {
	return &appError{status: http.StatusNotFound, message: message}
}

func NewErrConflicto(message string) error {
	return &appError{status: http.StatusConflict, message: message}
}

func NewErrBloqueado(message string) error {
	return &appError{status: http.StatusTooManyRequests, message: message}
}
