package explorer

import "errors"

// errors
var (
	ErrInvalidRequest         = errors.New("invalid request")
	ErrInvalidTransactionHash = errors.New("invalid transaction hash")
	ErrTransactionTimeout     = errors.New("transaction timeout")
	ErrTransactionFailed      = errors.New("transaction failed")
	ErrInvalidHeight          = errors.New("invalid height")
	ErrNotExist               = errors.New("not exist")
	ErrIsNotFormulator        = errors.New("Is not formulator")
	ErrInvalidAuthorization   = errors.New("invalid authorization")
)
