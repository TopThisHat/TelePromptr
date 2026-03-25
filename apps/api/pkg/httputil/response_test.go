package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	data := map[string]string{"hello": "world"}
	JSON(rec, http.StatusCreated, data)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if ct := rec.Header().Get("Content-Type"); ct != ContentTypeJSON {
		t.Errorf("Content-Type = %q, want %q", ct, ContentTypeJSON)
	}

	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if got["hello"] != "world" {
		t.Errorf("body = %v, want {hello: world}", got)
	}
}

func TestSuccess(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	data := map[string]int{"count": 42}
	Success(rec, data)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp SuccessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Pagination != nil {
		t.Error("expected nil pagination for non-paginated response")
	}

	// Data should be present.
	dataMap, ok := resp.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map data, got %T", resp.Data)
	}
	if count, ok := dataMap["count"].(float64); !ok || int(count) != 42 {
		t.Errorf("data.count = %v, want 42", dataMap["count"])
	}
}

func TestSuccessWithStatus(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	SuccessWithStatus(rec, http.StatusCreated, "created")

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp SuccessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Data != "created" {
		t.Errorf("data = %v, want %q", resp.Data, "created")
	}
}

func TestSuccessWithPagination(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		page           int
		perPage        int
		total          int
		wantTotalPages int
	}{
		{name: "exact division", page: 1, perPage: 10, total: 30, wantTotalPages: 3},
		{name: "with remainder", page: 2, perPage: 10, total: 25, wantTotalPages: 3},
		{name: "single page", page: 1, perPage: 50, total: 5, wantTotalPages: 1},
		{name: "empty result", page: 1, perPage: 10, total: 0, wantTotalPages: 0},
		{name: "zero perPage", page: 1, perPage: 0, total: 10, wantTotalPages: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			items := []string{"a", "b", "c"}
			SuccessWithPagination(rec, items, tt.page, tt.perPage, tt.total)

			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
			}

			var resp SuccessResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if resp.Pagination == nil {
				t.Fatal("expected pagination in response")
			}
			if resp.Pagination.Page != tt.page {
				t.Errorf("page = %d, want %d", resp.Pagination.Page, tt.page)
			}
			if resp.Pagination.PerPage != tt.perPage {
				t.Errorf("per_page = %d, want %d", resp.Pagination.PerPage, tt.perPage)
			}
			if resp.Pagination.Total != tt.total {
				t.Errorf("total = %d, want %d", resp.Pagination.Total, tt.total)
			}
			if resp.Pagination.TotalPages != tt.wantTotalPages {
				t.Errorf("total_pages = %d, want %d", resp.Pagination.TotalPages, tt.wantTotalPages)
			}
		})
	}
}

func TestErr(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	Err(rec, http.StatusBadRequest, "validation_error", "invalid input")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error.Code != "validation_error" {
		t.Errorf("error.code = %q, want %q", resp.Error.Code, "validation_error")
	}
	if resp.Error.Message != "invalid input" {
		t.Errorf("error.message = %q, want %q", resp.Error.Message, "invalid input")
	}
	if resp.Error.Details != nil {
		t.Error("expected nil details when not provided")
	}
}

func TestErrWithDetails(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	details := []string{"name is required", "email must be valid"}
	ErrWithDetails(rec, http.StatusUnprocessableEntity, "validation_error", "validation failed", details)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp.Error.Details) != 2 {
		t.Fatalf("details length = %d, want 2", len(resp.Error.Details))
	}
	if resp.Error.Details[0] != "name is required" {
		t.Errorf("details[0] = %q, want %q", resp.Error.Details[0], "name is required")
	}
	if resp.Error.Details[1] != "email must be valid" {
		t.Errorf("details[1] = %q, want %q", resp.Error.Details[1], "email must be valid")
	}
}

func TestNotFound(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	NotFound(rec, "resource not found")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error.Code != "not_found" {
		t.Errorf("error.code = %q, want %q", resp.Error.Code, "not_found")
	}
}

func TestBadRequest(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	BadRequest(rec, "invalid request")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error.Code != "bad_request" {
		t.Errorf("error.code = %q, want %q", resp.Error.Code, "bad_request")
	}
}

func TestInternalError(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	InternalError(rec, "something broke")

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error.Code != "internal_error" {
		t.Errorf("error.code = %q, want %q", resp.Error.Code, "internal_error")
	}
}

func TestUnauthorized(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	Unauthorized(rec, "not authenticated")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error.Code != "unauthorized" {
		t.Errorf("error.code = %q, want %q", resp.Error.Code, "unauthorized")
	}
}

func TestForbidden(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	Forbidden(rec, "access denied")

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error.Code != "forbidden" {
		t.Errorf("error.code = %q, want %q", resp.Error.Code, "forbidden")
	}
}

func TestSuccessResponse_PaginationOmittedWhenNil(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	Success(rec, "test")

	body := rec.Body.String()
	// The JSON should not contain "pagination" when it is nil.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if _, hasPagination := raw["pagination"]; hasPagination {
		t.Error("pagination should be omitted when nil")
	}
}
