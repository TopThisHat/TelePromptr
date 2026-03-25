package validate

import (
	"strings"
	"testing"
)

func TestRequired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{name: "non-empty", value: "hello", valid: true},
		{name: "empty", value: "", valid: false},
		{name: "whitespace only", value: "   ", valid: false},
		{name: "with whitespace", value: " hello ", valid: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New()
			v.Required("field", tt.value)
			if v.Valid() != tt.valid {
				t.Errorf("Valid() = %v, want %v for value %q", v.Valid(), tt.valid, tt.value)
			}
			if !tt.valid && !strings.Contains(v.Errors()[0], "is required") {
				t.Errorf("expected 'is required' in error, got %q", v.Errors()[0])
			}
		})
	}
}

func TestMinLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		min   int
		valid bool
	}{
		{name: "at minimum", value: "abc", min: 3, valid: true},
		{name: "above minimum", value: "abcdef", min: 3, valid: true},
		{name: "below minimum", value: "ab", min: 3, valid: false},
		{name: "empty skipped", value: "", min: 3, valid: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New()
			v.MinLength("field", tt.value, tt.min)
			if v.Valid() != tt.valid {
				t.Errorf("Valid() = %v, want %v", v.Valid(), tt.valid)
			}
		})
	}
}

func TestMaxLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		max   int
		valid bool
	}{
		{name: "at maximum", value: "abc", max: 3, valid: true},
		{name: "below maximum", value: "ab", max: 3, valid: true},
		{name: "above maximum", value: "abcd", max: 3, valid: false},
		{name: "empty", value: "", max: 3, valid: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New()
			v.MaxLength("field", tt.value, tt.max)
			if v.Valid() != tt.valid {
				t.Errorf("Valid() = %v, want %v", v.Valid(), tt.valid)
			}
		})
	}
}

func TestMinInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value int
		min   int
		valid bool
	}{
		{name: "at minimum", value: 5, min: 5, valid: true},
		{name: "above minimum", value: 10, min: 5, valid: true},
		{name: "below minimum", value: 3, min: 5, valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New()
			v.MinInt("field", tt.value, tt.min)
			if v.Valid() != tt.valid {
				t.Errorf("Valid() = %v, want %v", v.Valid(), tt.valid)
			}
		})
	}
}

func TestMaxInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value int
		max   int
		valid bool
	}{
		{name: "at maximum", value: 100, max: 100, valid: true},
		{name: "below maximum", value: 50, max: 100, valid: true},
		{name: "above maximum", value: 101, max: 100, valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New()
			v.MaxInt("field", tt.value, tt.max)
			if v.Valid() != tt.valid {
				t.Errorf("Valid() = %v, want %v", v.Valid(), tt.valid)
			}
		})
	}
}

func TestIntRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value int
		min   int
		max   int
		valid bool
	}{
		{name: "in range", value: 50, min: 1, max: 100, valid: true},
		{name: "at min", value: 1, min: 1, max: 100, valid: true},
		{name: "at max", value: 100, min: 1, max: 100, valid: true},
		{name: "below min", value: 0, min: 1, max: 100, valid: false},
		{name: "above max", value: 101, min: 1, max: 100, valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New()
			v.IntRange("field", tt.value, tt.min, tt.max)
			if v.Valid() != tt.valid {
				t.Errorf("Valid() = %v, want %v", v.Valid(), tt.valid)
			}
		})
	}
}

func TestMinFloat(t *testing.T) {
	t.Parallel()

	v := New()
	v.MinFloat("temperature", 0.5, 0.0)
	if !v.Valid() {
		t.Error("expected valid for 0.5 >= 0.0")
	}

	v2 := New()
	v2.MinFloat("temperature", -0.1, 0.0)
	if v2.Valid() {
		t.Error("expected invalid for -0.1 < 0.0")
	}
}

func TestMaxFloat(t *testing.T) {
	t.Parallel()

	v := New()
	v.MaxFloat("temperature", 1.5, 2.0)
	if !v.Valid() {
		t.Error("expected valid for 1.5 <= 2.0")
	}

	v2 := New()
	v2.MaxFloat("temperature", 2.1, 2.0)
	if v2.Valid() {
		t.Error("expected invalid for 2.1 > 2.0")
	}
}

func TestOneOf(t *testing.T) {
	t.Parallel()

	allowed := []string{"gpt-4", "gpt-3.5", "claude-3"}

	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{name: "valid value", value: "gpt-4", valid: true},
		{name: "another valid", value: "claude-3", valid: true},
		{name: "invalid value", value: "bard", valid: false},
		{name: "empty skipped", value: "", valid: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New()
			v.OneOf("model", tt.value, allowed)
			if v.Valid() != tt.valid {
				t.Errorf("Valid() = %v, want %v", v.Valid(), tt.valid)
			}
		})
	}
}

func TestCustom(t *testing.T) {
	t.Parallel()

	v := New()
	v.Custom(true, "should not appear")
	if !v.Valid() {
		t.Error("expected valid when condition is true")
	}

	v2 := New()
	v2.Custom(false, "custom error message")
	if v2.Valid() {
		t.Error("expected invalid when condition is false")
	}
	if v2.Errors()[0] != "custom error message" {
		t.Errorf("error = %q, want %q", v2.Errors()[0], "custom error message")
	}
}

func TestMultipleErrors(t *testing.T) {
	t.Parallel()

	v := New()
	v.Required("name", "")
	v.MinInt("age", -1, 0)
	v.MaxLength("bio", "this is too long", 5)

	if v.Valid() {
		t.Fatal("expected invalid with multiple errors")
	}
	if len(v.Errors()) != 3 {
		t.Errorf("error count = %d, want 3", len(v.Errors()))
	}
}

func TestError_Valid(t *testing.T) {
	t.Parallel()

	v := New()
	v.Required("name", "present")

	if err := v.Error(); err != nil {
		t.Errorf("expected nil error for valid input, got %v", err)
	}
}

func TestError_Invalid(t *testing.T) {
	t.Parallel()

	v := New()
	v.Required("name", "")
	v.MinInt("count", 0, 1)

	err := v.Error()
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("error = %q, want to contain 'validation failed'", err.Error())
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Errorf("error = %q, want to contain 'name is required'", err.Error())
	}
}
