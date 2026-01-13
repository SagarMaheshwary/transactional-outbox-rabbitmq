package constant

import "errors"

var (
	ErrPermanent = errors.New("permanent error")
	ErrTransient = errors.New("transient error")
)
