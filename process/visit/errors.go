package visit

import "errors"

// errors
var (
	ErrInvalidVisitID = errors.New("invalid visit id")
	ErrInvalidVisit   = errors.New("invalid visit")
	ErrExistVisit     = errors.New("exist visit")
	ErrNotExistVisit  = errors.New("not exist visit")
)
