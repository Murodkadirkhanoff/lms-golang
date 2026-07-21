package domain

import "errors"

// Domain errors owned by the enrollment context.
var (
	ErrNotFound     = errors.New("enrollment resource not found")
	ErrNotPermitted = errors.New("not permitted")
)
