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
	errors := make([]error, 0)
	if err := definition.Validate(); err != nil {
		errors = append(errors, err)
	}
	for _, field := range definition.Fields {
		if !fieldNamePattern.MatchString(field.Name) {
			errors = append(errors, fmt.Errorf("field %q must use lower snake case", field.Name))
		}
		if strings.TrimSpace(field.Label) == "" {
			errors = append(errors, fmt.Errorf("field %q label is required", field.Name))
		}
	}
	return errors
}

func JoinValidationErrors(definition domain.Method) error {
	validationErrors := ValidateDefinition(definition)
	if len(validationErrors) == 0 {
		return nil
	}
	messages := make([]string, 0, len(validationErrors))
	for _, validationErr := range validationErrors {
		messages = append(messages, validationErr.Error())
	}
	return errors.New(strings.Join(messages, "; "))
}

func IsImmutable(definition domain.Method) bool {
	return definition.Status == domain.MethodPublished || definition.Status == domain.MethodArchived
}
