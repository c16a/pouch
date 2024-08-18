package commands

import "errors"

var (
	ErrorInvalidDataType = errors.New("InvalidDataType")
	ErrorNotFound        = errors.New("NotFound")
	ErrInvalidCommand    = errors.New("InvalidCommand")
	ErrEmptyCommand      = errors.New("EmptyCommand")
)
