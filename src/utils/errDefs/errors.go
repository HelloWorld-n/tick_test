package errDefs

import (
	"errors"
	"fmt"
)

var ErrServerInternalError = errors.New("internal server error")
var ErrDatabaseOffline = fmt.Errorf("%w: database offline", ErrServerInternalError)
var ErrDoesExist = errors.New("item already exists")
var ErrBadRequest = errors.New("bad request")
var ErrMissingField = fmt.Errorf("%w: field missing", ErrBadRequest)
var ErrUnauthorized = errors.New("unauthorized")
