package api

import (
	"errors"
	"fmt"
	"testing"
)

func TestFieldErrorsOfFindsWrappedFields(t *testing.T) {
	fieldsErr := &FieldsError{Fields: []FieldError{{Field: "retries", Message: "must be at most 5"}}}
	wrapped := fmt.Errorf("save: %w", fieldsErr)

	got := FieldErrorsOf(wrapped)
	if len(got) != 1 || got[0].Field != "retries" {
		t.Fatalf("FieldErrorsOf = %+v, want the retries field", got)
	}
	if fieldsErr.Error() != "retries must be at most 5" {
		t.Fatalf("Error() = %q", fieldsErr.Error())
	}
}

func TestFieldErrorsOfPlainErrorHasNone(t *testing.T) {
	if got := FieldErrorsOf(errors.New("boom")); got != nil {
		t.Fatalf("FieldErrorsOf = %+v, want nil", got)
	}
}
