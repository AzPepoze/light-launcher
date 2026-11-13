package api

import "errors"

// CodedError carries a stable machine-readable code next to the human message. The UI can act on
// the code instead of matching error text, which changes with wording and provider details.
type CodedError struct {
	Code string
	Err  error
}

func (e *CodedError) Error() string { return e.Err.Error() }
func (e *CodedError) Unwrap() error { return e.Err }

// WithCode wraps err with a stable code. A nil err stays nil.
func WithCode(err error, code string) error {
	if err == nil {
		return nil
	}
	return &CodedError{Code: code, Err: err}
}

// ErrorCodeOf returns the innermost stable code on err, or an empty string when it has none.
func ErrorCodeOf(err error) string {
	var coded *CodedError
	if errors.As(err, &coded) {
		return coded.Code
	}
	var named interface{ ErrorCode() string }
	if errors.As(err, &named) {
		return named.ErrorCode()
	}
	return ""
}
