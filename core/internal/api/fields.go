package api

import (
	"errors"
	"strings"
)

// FieldError names one input that failed validation and says what is wrong with it.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FieldsError is a failed validation with one entry per bad field. The UI shows each entry next to its input.
type FieldsError struct {
	Fields []FieldError
}

// Error joins the field messages so logs show every problem at once.
func (e *FieldsError) Error() string {
	parts := make([]string, 0, len(e.Fields))
	for _, field := range e.Fields {
		parts = append(parts, field.Field+" "+field.Message)
	}
	return strings.Join(parts, "; ")
}

// FieldErrorsOf returns the field errors inside err, or nil when err has none.
func FieldErrorsOf(err error) []FieldError {
	var fieldsErr *FieldsError
	if errors.As(err, &fieldsErr) {
		return fieldsErr.Fields
	}
	return nil
}
