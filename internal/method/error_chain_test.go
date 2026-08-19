package method

import (
	"errors"
	"testing"

	"material-lab-platform/internal/domain"
)

func TestJoinValidationErrorsKeepsSentinelAndFieldContext(t *testing.T) {
	definition := domain.Method{Code: "method", Name: "Method", Version: 1, Fields: []domain.MethodField{{Name: "Bad-Name", Type: domain.FieldNumber}}, Formula: "unknown + 1"}
	err := JoinValidationErrors(definition)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("validation sentinel lost: %v", err)
	}
	var fieldErr *domain.MethodFieldError
	if !errors.As(err, &fieldErr) {
		t.Fatalf("field context lost: %v", err)
	}
	fields := collectMethodErrorFields(err)
	if fields["formula"] != 1 || fields["Bad-Name"] != 2 {
		t.Fatalf("field errors = %#v, want formula once and Bad-Name twice", fields)
	}
}

func collectMethodErrorFields(err error) map[string]int {
	fields := map[string]int{}
	var walk func(error)
	walk = func(current error) {
		if current == nil {
			return
		}
		if fieldErr, ok := current.(*domain.MethodFieldError); ok {
			fields[fieldErr.Field]++
		}
		if joined, ok := current.(interface{ Unwrap() []error }); ok {
			for _, child := range joined.Unwrap() {
				walk(child)
			}
			return
		}
		if wrapped, ok := current.(interface{ Unwrap() error }); ok {
			walk(wrapped.Unwrap())
		}
	}
	walk(err)
	return fields
}
