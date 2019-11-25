package query

import "errors"

// errors
var (
	ErrInvalidQueryID = errors.New("invalid query id")
	ErrInvalidQuery   = errors.New("invalid query")
	ErrExistQuery     = errors.New("exist query")
	ErrNotExistQuery  = errors.New("not exist query")
	ErrClosedQuery    = errors.New("closed query")
)
