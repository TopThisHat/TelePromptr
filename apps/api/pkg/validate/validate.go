// Package validate provides request validation utilities for the TelePromptr API.
// It supports building a set of validation rules for struct fields and collecting
// all violations into a structured error suitable for API error responses.
package validate

import (
	"fmt"
	"strings"
)

// Validator accumulates validation errors for a set of fields. It is not
// safe for concurrent use; create a new Validator per request.
type Validator struct {
	errors []string
}

// New creates a new empty Validator.
func New() *Validator {
	return &Validator{}
}

// Required checks that the given string value is non-empty. If empty, a
// validation error is recorded for the named field.
func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.errors = append(v.errors, fmt.Sprintf("%s is required", field))
	}
}

// MinLength checks that the string value has at least min characters.
// An empty string is not checked here (use Required for that).
func (v *Validator) MinLength(field, value string, min int) {
	if value != "" && len(value) < min {
		v.errors = append(v.errors, fmt.Sprintf("%s must be at least %d characters", field, min))
	}
}

// MaxLength checks that the string value has at most max characters.
func (v *Validator) MaxLength(field, value string, max int) {
	if len(value) > max {
		v.errors = append(v.errors, fmt.Sprintf("%s must be at most %d characters", field, max))
	}
}

// MinInt checks that the integer value is at least min.
func (v *Validator) MinInt(field string, value, min int) {
	if value < min {
		v.errors = append(v.errors, fmt.Sprintf("%s must be at least %d", field, min))
	}
}

// MaxInt checks that the integer value is at most max.
func (v *Validator) MaxInt(field string, value, max int) {
	if value > max {
		v.errors = append(v.errors, fmt.Sprintf("%s must be at most %d", field, max))
	}
}

// IntRange checks that the integer value falls within [min, max] inclusive.
func (v *Validator) IntRange(field string, value, min, max int) {
	if value < min || value > max {
		v.errors = append(v.errors, fmt.Sprintf("%s must be between %d and %d", field, min, max))
	}
}

// MinFloat checks that the float value is at least min.
func (v *Validator) MinFloat(field string, value, min float64) {
	if value < min {
		v.errors = append(v.errors, fmt.Sprintf("%s must be at least %g", field, min))
	}
}

// MaxFloat checks that the float value is at most max.
func (v *Validator) MaxFloat(field string, value, max float64) {
	if value > max {
		v.errors = append(v.errors, fmt.Sprintf("%s must be at most %g", field, max))
	}
}

// OneOf checks that value is one of the allowed values. If the value is empty,
// it is skipped (use Required for presence checks).
func (v *Validator) OneOf(field, value string, allowed []string) {
	if value == "" {
		return
	}
	for _, a := range allowed {
		if value == a {
			return
		}
	}
	v.errors = append(v.errors, fmt.Sprintf("%s must be one of: %s", field, strings.Join(allowed, ", ")))
}

// Custom adds a validation error if the given condition is false.
func (v *Validator) Custom(condition bool, message string) {
	if !condition {
		v.errors = append(v.errors, message)
	}
}

// Valid returns true if no validation errors have been recorded.
func (v *Validator) Valid() bool {
	return len(v.errors) == 0
}

// Errors returns the accumulated validation error messages.
func (v *Validator) Errors() []string {
	return v.errors
}

// Error returns a single error joining all validation messages, or nil if valid.
func (v *Validator) Error() error {
	if v.Valid() {
		return nil
	}
	return fmt.Errorf("validation failed: %s", strings.Join(v.errors, "; "))
}
