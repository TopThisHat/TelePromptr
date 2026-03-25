package validate

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSON_ValidPayload(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	body := `{"name":"test","count":42}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	var p payload
	if err := DecodeJSON(req, &p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "test" {
		t.Errorf("Name = %q, want %q", p.Name, "test")
	}
	if p.Count != 42 {
		t.Errorf("Count = %d, want 42", p.Count)
	}
}

func TestDecodeJSON_InvalidJSON(t *testing.T) {
	t.Parallel()

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	var dst map[string]any
	err := DecodeJSON(req, &dst)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("error = %q, want to contain 'invalid JSON'", err.Error())
	}
}

func TestDecodeJSON_NilBody(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Body = nil

	var dst map[string]any
	err := DecodeJSON(req, &dst)
	if err == nil {
		t.Fatal("expected error for nil body")
	}
}

func TestDecodeJSON_UnknownFields(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	body := `{"name":"test","unknown":"field"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	var p payload
	err := DecodeJSON(req, &p)
	if err == nil {
		t.Fatal("expected error for unknown fields")
	}
}

func TestDecodeJSON_MultipleValues(t *testing.T) {
	t.Parallel()

	body := `{"name":"first"}{"name":"second"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	var dst map[string]any
	err := DecodeJSON(req, &dst)
	if err == nil {
		t.Fatal("expected error for trailing content")
	}
	if !strings.Contains(err.Error(), "single JSON value") {
		t.Errorf("error = %q, want to mention single JSON value", err.Error())
	}
}

func TestDecodeAndValidate_Success(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	body := `{"name":"Alice","age":30}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()

	var p payload
	ok := DecodeAndValidate(rec, req, &p, func(v *Validator) {
		v.Required("name", p.Name)
		v.MinInt("age", p.Age, 1)
	})

	if !ok {
		t.Fatal("expected DecodeAndValidate to return true")
	}
	if p.Name != "Alice" {
		t.Errorf("Name = %q, want %q", p.Name, "Alice")
	}
}

func TestDecodeAndValidate_InvalidJSON(t *testing.T) {
	t.Parallel()

	body := `{bad json}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()

	var dst map[string]any
	ok := DecodeAndValidate(rec, req, &dst, func(v *Validator) {})

	if ok {
		t.Fatal("expected false for invalid JSON")
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDecodeAndValidate_ValidationFailure(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	body := `{"name":""}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()

	var p payload
	ok := DecodeAndValidate(rec, req, &p, func(v *Validator) {
		v.Required("name", p.Name)
	})

	if ok {
		t.Fatal("expected false for validation failure")
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}

	// Check error response structure.
	var errResp struct {
		Error struct {
			Code    string   `json:"code"`
			Message string   `json:"message"`
			Details []string `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if errResp.Error.Code != "validation_error" {
		t.Errorf("error.code = %q, want %q", errResp.Error.Code, "validation_error")
	}
	if len(errResp.Error.Details) == 0 {
		t.Error("expected non-empty details array")
	}
}

func TestReadBody_Success(t *testing.T) {
	t.Parallel()

	body := "hello world"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	data, err := ReadBody(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != body {
		t.Errorf("body = %q, want %q", string(data), body)
	}
}

func TestReadBody_NilBody(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Body = nil

	_, err := ReadBody(req)
	if err == nil {
		t.Fatal("expected error for nil body")
	}
}
