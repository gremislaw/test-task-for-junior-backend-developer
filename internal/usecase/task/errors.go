package task

import "errors"

var ErrInvalidInput = errors.New("invalid task input")
var ErrMaterialize = errors.New("error when materializing")
var ErrRangeTooLarge = errors.New("date range too large (max 365 days)")
