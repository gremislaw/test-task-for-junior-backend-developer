package task

import "errors"

var ErrNotFound = errors.New("task not found")
var ErrCannotUnmarshalRecurrenceConfig = errors.New("cannot unmarshal recurrence config")
