package rc

import "errors"

var (
	ErrBaseUrlNotSet = errors.New("base url not set")
	ErrNonNilContext = errors.New("context must be non-nil")
)
