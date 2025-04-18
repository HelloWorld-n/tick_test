package errDefs

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrInternalServerError = errors.New("internal server error")
var ErrConflict = errors.New("conflict")
var ErrDatabaseOffline = fmt.Errorf("%w: database offline", ErrInternalServerError)
var ErrDoesExist = fmt.Errorf("%w: item already exists", ErrConflict)
var ErrBadRequest = errors.New("bad request")
var ErrMissingField = fmt.Errorf("%w: field missing", ErrBadRequest)
var ErrUnauthorized = errors.New("unauthorized")

func DetermineStatus(err error) (status int) {
	switch {
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
