package errs

import "errors"

var (
	InvalidRequest = errors.New("invalid Request")
	AlreadyExists  = errors.New("user already exists")
	NotFound       = errors.New("user not found")
)
