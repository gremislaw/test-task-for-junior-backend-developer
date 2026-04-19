package task

import "errors"

var (
	ErrInvalidInput             = errors.New("invalid task input")
	ErrMaterialize              = errors.New("error when materializing")
	ErrRangeTooLarge            = errors.New("date range too large (max 365 days)")
	ErrDeleteOldInstance        = errors.New("cannot delete old instance")
	ErrInstanceFieldsImmutable  = errors.New("recurrence fields are immutable for instances")
	ErrInstanceCannotBeTemplate = errors.New("instances cannot be converted to recurrence templates")
)
