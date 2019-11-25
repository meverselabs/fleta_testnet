package subject

import "errors"

// errors
var (
	ErrInvalidSubjectID = errors.New("invalid subject id")
	ErrExistSubject     = errors.New("exist subject")
	ErrNotExistSubject  = errors.New("not exist subject")
)
