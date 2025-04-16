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
	if errors.Is(err, ErrConflict) {
		return http.StatusConflict
	}
	if errors.Is(err, ErrBadRequest) {
		return http.StatusBadRequest
	}
	if errors.Is(err, ErrUnauthorized) {
		return http.StatusUnauthorized
	}
	return http.StatusInternalServerError
}
