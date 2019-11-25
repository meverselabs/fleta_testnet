package user

import "errors"

// errors
var (
	ErrInvalidUserID = errors.New("invalid user id")
	ErrExistUser     = errors.New("exist user")
	ErrNotExistUser  = errors.New("not exist user")
	ErrInvalidRole   = errors.New("invalid user")
)
