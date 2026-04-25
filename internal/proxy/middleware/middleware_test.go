package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogger(t *testing.T) {
	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with logger middleware
	loggedHandler := Logger(handler)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Call the handler
	loggedHandler.ServeHTTP(w, req)

	// Check response
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestResponseWriter_Status(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := newResponseWriter(rec)

	rw.WriteHeader(http.StatusNotFound)

	if rw.Status() != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rw.Status())
	}
}

func TestResponseWriter_Written(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := newResponseWriter(rec)

	if rw.Written() {
		t.Error("Expected Written() to be false before writing")
	}

	rw.Write([]byte("test"))

	if !rw.Written() {
		t.Error("Expected Written() to be true after writing")
	}
}

func TestResponseWriter_Body(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := newResponseWriter(rec)

	testBody := []byte("test body content")
	rw.Write(testBody)

	if string(rw.Body()) != string(testBody) {
		t.Errorf("Expected body '%s', got '%s'", string(testBody), string(rw.Body()))
	}
}

func TestResponseWriter_DefaultStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := newResponseWriter(rec)

	// Write without calling WriteHeader should default to 200
	rw.Write([]byte("test"))

	if rw.Status() != http.StatusOK {
		t.Errorf("Expected status 200 (default), got %d", rw.Status())
	}
}

func TestResponseWriter_Before(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := newResponseWriter(rec)

	called := false
	rw.Before(func(w ResponseWriter) {
		called = true
	})

	rw.WriteHeader(http.StatusOK)

	if !called {
		t.Error("Before function was not called")
	}
}

func TestResponseWriter_Flush(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := newResponseWriter(rec)

	// Should not panic
	rw.Flush()

	// After flush without write, status should be OK
	if rw.Status() != http.StatusOK {
		t.Errorf("Expected status 200 after flush, got %d", rw.Status())
	}
}

func TestPrettyJSON(t *testing.T) {
	input := `{"key": "value"}`
	expected := `{'key': 'value'}`

	result := prettyJSON([]byte(input))
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}
