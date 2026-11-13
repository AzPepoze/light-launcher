package contract

import "errors"

// ErrAuthRequired means the source needs a sign-in or an API key before it can answer.
var ErrAuthRequired = errors.New("the source needs a sign-in or an API key")
