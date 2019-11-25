package study

import "errors"

// errors
var (
	ErrInvalidStudyID  = errors.New("invalid study id")
	ErrExistStudy      = errors.New("exist study")
	ErrNotStudyAccount = errors.New("not study account")
	ErrNotSiteAccount  = errors.New("not site account")
)
