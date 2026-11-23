package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"testing/iotest"
)

const apiPath = "/api/calculate"
const validBody = `{"expression":"1 + 2"}`

// API tests
func TestAPI(t *testing.T) {
	// A valid JSON request returns the calculated result.
	t.Run("ValidRequest", func(t *testing.T) {
		body := " \n" + `{"expression": "7 / 2"}` + "\r\n"
		w := callAPI("POST", apiPath, "application/json; charset=utf-8", strings.NewReader(body))
		checkResult(t, w, 3.5)
	})

	// Only POST is allowed on the calculation endpoint.
	t.Run("UnsupportedMethods", func(t *testing.T) {
		for _, method := range []string{"GET", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"} {
			t.Run(method, func(t *testing.T) {
				w := callAPI(method, apiPath, "application/json", strings.NewReader(validBody))
				if w.Code != 405 || w.Header().Get("Allow") != "POST" {
					t.Fatalf("expected HTTP 405 with Allow: POST, got %d %v", w.Code, w.Header())
				}
				if method != "HEAD" {
					checkError(t, w, 405, "method_not_allowed")
				}
			})
		}
	})

	// Requests to unknown paths return 404.
	t.Run("UnknownPaths", func(t *testing.T) {
		for _, path := range []string{"/", "/missing", apiPath + "/", "/api/Calculate"} {
			t.Run(path, func(t *testing.T) {
				w := callAPI("POST", path, "application/json", strings.NewReader(validBody))
				checkError(t, w, 404, "not_found")
			})
		}
	})

	// A missing or unsupported content type returns 415.
	t.Run("UnsupportedContentTypes", func(t *testing.T) {
		for _, media := range []string{"", "text/plain", "application/x-www-form-urlencoded", "application/json; charset"} {
			t.Run("media/"+media, func(t *testing.T) {
				w := callAPI("POST", apiPath, media, strings.NewReader(validBody))
				checkError(t, w, 415, "unsupported_media_type")
			})
		}
	})

	// The API accepts bodies up to 16 KiB and rejects larger ones.
	t.Run("RequestBodyLimit", func(t *testing.T) {
		const limit = 16 * 1024
		checkError(t, postJSON(strings.Repeat(" ", limit+1)), 413, "body_too_large")
		for _, size := range []int{limit - 1, limit, limit + 1} {
			t.Run(fmt.Sprint(size), func(t *testing.T) {
				body := validBody + strings.Repeat(" ", size-len(validBody))
				// MultiReader gives the handler a streamed body of unknown length.
				w := callAPI("POST", apiPath, "application/json", io.MultiReader(strings.NewReader(body)))
				if size > limit {
					checkError(t, w, 413, "body_too_large")
				} else {
					checkResult(t, w, 3)
				}
			})
		}
	})

	// A read error while receiving the request body returns 400.
	t.Run("InterruptedBody", func(t *testing.T) {
		body := io.MultiReader(strings.NewReader(validBody), iotest.ErrReader(io.ErrUnexpectedEOF))
		checkError(t, callAPI("POST", apiPath, "application/json", body), 400, "invalid_json")
	})

	// An expression can contain up to 128 operations.
	t.Run("OperationCountLimit", func(t *testing.T) {
		for _, n := range []int{0, 1, 127, 128, 129} {
			t.Run(fmt.Sprint(n), func(t *testing.T) {
				body := fmt.Sprintf(`{"expression":%q}`, "1"+strings.Repeat(" + 1", n))
				if n > 128 {
					checkError(t, postJSON(body), 400, "invalid_expression")
				} else {
					checkResult(t, postJSON(body), float64(n+1))
				}
			})
		}
	})

	// Concurrent requests keep their results and errors independent.
	t.Run("ConcurrentRequestsAreIndependent", func(t *testing.T) {
		for i := 1; i <= 24; i++ {
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				t.Parallel()
				body := fmt.Sprintf(`{"expression":"%d + 2 * 3"}`, i)
				checkResult(t, postJSON(body), float64(i+6))
				checkError(t, postJSON(`{"expression":"1 / 0"}`), 422, "division_by_zero")
				checkResult(t, postJSON(body), float64(i+6))
			})
		}
	})

	// Invalid JSON or an invalid expression returns the expected error code.
	t.Run("RequestValidation", func(t *testing.T) {
		for _, tc := range []struct{ name, body, code string }{
			{"unknown property", `{"extra":0}`, "invalid_json"},
			{"multiple JSON values", `{"expression":"1 + 2"} {}`, "invalid_json"},
			{"malformed JSON", `{"expression":`, "invalid_json"},
			{"empty body", "", "invalid_json"},
			{"only whitespace", " \r\n\t", "invalid_json"},
			{"array at root", "[1,2]", "invalid_json"},
			{"scalar at root", "5", "invalid_json"},
			{"numeric expression", `{"expression":12}`, "invalid_json"},
			{"array expression", `{"expression":["1+2"]}`, "invalid_json"},
			{"boolean expression", `{"expression":false}`, "invalid_json"},
			{"object expression", `{"expression":{}}`, "invalid_json"},
			{"trailing garbage", validBody + "!", "invalid_json"},
			{"trailing comma", `{"expression":"1+2",}`, "invalid_json"},
			{"unexpected field with expression", `{"expression":"2","extra":null}`, "invalid_json"},
			{"null request", "null", "invalid_expression"},
			{"missing expression", "{}", "invalid_expression"},
			{"null expression", `{"expression":null}`, "invalid_expression"},
			{"blank expression", `{"expression":""}`, "invalid_expression"},
			{"incomplete expression", `{"expression":"1 + (2"}`, "invalid_expression"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				checkError(t, postJSON(tc.body), 400, tc.code)
			})
		}
	})

	// Advanced expressions use the same JSON result and arithmetic error responses.
	t.Run("ExpressionRequests", func(t *testing.T) {
		checkResult(t, postJSON(`{"expression":"(200 + 10%) / 2"}`), 110)
		checkResult(t, postJSON(`{"expression":"√(9 + 7) ^ 2"}`), 16)
		checkError(t, postJSON(`{"expression":"1 / (3 - 3)"}`), 422, "division_by_zero")
	})
}

// Arbitrary request bodies must return a valid result or a documented error.
func FuzzCalculateJSON(f *testing.F) {
	for _, seed := range []string{
		validBody, "", "null", "[]", `{"expression":"1 / 0"}`,
		`{"expression":"(200 + 10%) / 2"}`, `{"expression":"sqrt(-1)"}`,
		`{"expression":"1e301 * 1e100"}`, validBody + validBody,
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, body string) {
		w := postJSON(body)
		switch {
		case len(body) > 16*1024:
			checkError(t, w, 413, "body_too_large")
		case !json.Valid([]byte(body)):
			checkError(t, w, 400, "invalid_json")
		case w.Code == 200:
			resultValue(t, w)
		case w.Code == 400:
			checkError(t, w, 400, "invalid_json", "invalid_expression")
		case w.Code == 422:
			checkError(t, w, 422, "division_by_zero", "out_of_range", "domain_error")
		default:
			t.Fatalf("unexpected status %d: %s", w.Code, w.Body)
		}
	})
}

// HTTP test helpers
func callAPI(method, path, media string, body io.Reader) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, body)
	if media != "" {
		r.Header.Set("Content-Type", media)
	}
	w := httptest.NewRecorder()
	ServeAPI(w, r)
	return w
}

func postJSON(body string) *httptest.ResponseRecorder {
	return callAPI(http.MethodPost, apiPath, "application/json", strings.NewReader(body))
}

// Verify the HTTP status and the single top-level JSON field for every response.
func responseField(t *testing.T, w *httptest.ResponseRecorder, status int, field string) json.RawMessage {
	t.Helper()
	if w.Code != status {
		t.Fatalf("HTTP %d, want %d: %s", w.Code, status, w.Body)
	}
	if media := w.Header().Get("Content-Type"); media != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", media)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &fields); err != nil || len(fields) != 1 || fields[field] == nil {
		t.Fatalf("expected one JSON field %q, got %s", field, w.Body)
	}
	return fields[field]
}

func resultValue(t *testing.T, w *httptest.ResponseRecorder) float64 {
	t.Helper()
	raw := responseField(t, w, 200, "result")
	var value *float64
	if err := json.Unmarshal(raw, &value); err != nil || value == nil {
		t.Fatalf("expected a JSON number, got %s", raw)
	}
	return *value
}

func checkResult(t *testing.T, w *httptest.ResponseRecorder, want float64) {
	t.Helper()
	if got := resultValue(t, w); got != want {
		t.Errorf("result = %g, want %g", got, want)
	}
}

func checkError(t *testing.T, w *httptest.ResponseRecorder, status int, codes ...string) {
	t.Helper()
	raw := responseField(t, w, status, "error")
	var detail struct{ Code, Message string }
	if err := json.Unmarshal(raw, &detail); err != nil || detail.Message == "" || !slices.Contains(codes, detail.Code) {
		t.Fatalf("expected a message and error code in %v, got %s", codes, raw)
	}
}
