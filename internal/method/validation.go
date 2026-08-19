package method

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"material-lab-platform/internal/domain"
)

var fieldNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

func ValidateDefinition(definition domain.Method) []error {
	validationErrors := make([]error, 0)
	if err := definition.Validate(); err != nil {
		validationErrors = append(validationErrors, err)
	}
	for _, field := range definition.Fields {
		if !fieldNamePattern.MatchString(field.Name) {
			validationErrors = append(validationErrors, &domain.MethodFieldError{Field: field.Name, Err: fmt.Errorf("%w: must use lower snake case", domain.ErrValidation)})
		}
		if strings.TrimSpace(field.Label) == "" {
			validationErrors = append(validationErrors, &domain.MethodFieldError{Field: field.Name, Err: fmt.Errorf("%w: label is required", domain.ErrValidation)})
		}
	}
	return validationErrors
}

func JoinValidationErrors(definition domain.Method) error {
	validationErrors := ValidateDefinition(definition)
	if len(validationErrors) == 0 {
		return nil
	}
	return errors.Join(validationErrors...)
}

func IsImmutable(definition domain.Method) bool {
	return definition.Status == domain.MethodPublished || definition.Status == domain.MethodArchived
}
