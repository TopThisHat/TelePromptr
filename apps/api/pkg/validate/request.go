package validate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ralphlozano/telepromptr/apps/api/pkg/httputil"
)

// maxRequestBodySize is the maximum allowed request body size (1 MB).
const maxRequestBodySize = 1 << 20

// DecodeJSON reads the request body, decodes it as JSON into dst, and returns
// an error suitable for an API response if decoding fails. The request body
// is limited to 1 MB to prevent abuse.
func DecodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}

	r.Body = http.MaxBytesReader(nil, r.Body, maxRequestBodySize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Ensure no trailing content after the JSON value.
	if decoder.More() {
		return fmt.Errorf("request body must contain a single JSON value")
	}

	return nil
}

// DecodeAndValidate reads the request body as JSON into dst, then runs the
// provided validation function. If decoding or validation fails, it writes
// an appropriate error response and returns false. On success it returns true
// and the caller can proceed with the decoded and validated data.
func DecodeAndValidate(w http.ResponseWriter, r *http.Request, dst any, validateFn func(*Validator)) bool {
	if err := DecodeJSON(r, dst); err != nil {
		httputil.BadRequest(w, err.Error())
		return false
	}

	v := New()
	validateFn(v)
	if !v.Valid() {
		httputil.ErrWithDetails(w, http.StatusUnprocessableEntity, "validation_error", "validation failed", v.Errors())
		return false
	}

	return true
}

// ReadBody reads the entire request body and returns it as bytes.
// The body is limited to maxRequestBodySize.
func ReadBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, fmt.Errorf("request body is empty")
	}
	r.Body = http.MaxBytesReader(nil, r.Body, maxRequestBodySize)
	return io.ReadAll(r.Body)
}
