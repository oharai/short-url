package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oharai/short-url/internal/shorturl/app"
)

// Mock service for testing
type mockShortURLService struct {
	createResponse  *app.CreateShortURLResponse
	createError     error
	longURL         string
	getLongError    error
	allURLs         []*app.ShortURLResponse
	getAllError     error
	deactivateError error
}

func (m *mockShortURLService) CreateShortURL(req app.CreateShortURLRequest) (*app.CreateShortURLResponse, error) {
	if m.createError != nil {
		return nil, m.createError
	}
	return m.createResponse, nil
}

func (m *mockShortURLService) GetLongURL(req app.GetLongURLRequest) (string, error) {
	if m.getLongError != nil {
		return "", m.getLongError
	}
	return m.longURL, nil
}

func (m *mockShortURLService) GetAllShortURLs() ([]*app.ShortURLResponse, error) {
	if m.getAllError != nil {
		return nil, m.getAllError
	}
	return m.allURLs, nil
}

func (m *mockShortURLService) DeactivateShortURL(id string) error {
	return m.deactivateError
}

func TestNewShortURLHandler(t *testing.T) {
	service := &mockShortURLService{}
	handler := NewShortURLHandler(service)

	if handler == nil {
		t.Error("expected handler to be created")
		return
	}

	if handler.service == nil {
		t.Error("expected service to be set")
	}
}

func TestShortURLHandler_CreateShortURL(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           interface{}
		setupService   func(*mockShortURLService)
		expectedStatus int
		expectedBody   string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful creation",
			method: "POST",
			body: app.CreateShortURLRequest{
				LongURL: "https://example.com",
			},
			setupService: func(m *mockShortURLService) {
				m.createResponse = &app.CreateShortURLResponse{
					ShortURL: "http://test.com/abc123",
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp app.CreateShortURLResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
					return
				}
				if resp.ShortURL != "http://test.com/abc123" {
					t.Errorf("expected short URL 'http://test.com/abc123', got %q", resp.ShortURL)
				}
			},
		},
		{
			name:           "wrong method",
			method:         "GET",
			body:           nil,
			setupService:   func(m *mockShortURLService) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid JSON",
			method:         "POST",
			body:           "invalid json",
			setupService:   func(m *mockShortURLService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid JSON"}`,
		},
		{
			name:   "service error - conflict",
			method: "POST",
			body: app.CreateShortURLRequest{
				LongURL:   "https://example.com",
				CustomURL: "existing",
			},
			setupService: func(m *mockShortURLService) {
				m.createError = errors.New("custom URL already exists")
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"custom URL already exists"}`,
		},
		{
			name:   "service error - bad request",
			method: "POST",
			body: app.CreateShortURLRequest{
				LongURL: "",
			},
			setupService: func(m *mockShortURLService) {
				m.createError = errors.New("longUrl is required")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"longUrl is required"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockShortURLService{}
			tt.setupService(service)
			handler := NewShortURLHandler(service)

			var body []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					body = []byte(str)
				} else {
					body, _ = json.Marshal(tt.body)
				}
			}

			req := httptest.NewRequest(tt.method, "/v1/createShortUrl", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateShortURL(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" {
				body := strings.TrimSpace(w.Body.String())
				if body != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, body)
				}
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestShortURLHandler_GetLongURL(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           interface{}
		setupService   func(*mockShortURLService)
		expectedStatus int
		expectedBody   string
		checkHeaders   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful retrieval",
			method: "GET",
			body: app.GetLongURLRequest{
				ShortURL: "http://test.com/abc123",
			},
			setupService: func(m *mockShortURLService) {
				m.longURL = "https://example.com"
			},
			expectedStatus: http.StatusFound,
			checkHeaders: func(t *testing.T, w *httptest.ResponseRecorder) {
				location := w.Header().Get("Location")
				if location != "https://example.com" {
					t.Errorf("expected Location header 'https://example.com', got %q", location)
				}
			},
		},
		{
			name:           "wrong method",
			method:         "POST",
			body:           nil,
			setupService:   func(m *mockShortURLService) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid JSON",
			method:         "GET",
			body:           "invalid json",
			setupService:   func(m *mockShortURLService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid JSON"}`,
		},
		{
			name:   "URL not found",
			method: "GET",
			body: app.GetLongURLRequest{
				ShortURL: "http://test.com/notfound",
			},
			setupService: func(m *mockShortURLService) {
				m.getLongError = errors.New("short URL not found")
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"short URL not found"}`,
		},
		{
			name:   "service error",
			method: "GET",
			body: app.GetLongURLRequest{
				ShortURL: "http://test.com/error",
			},
			setupService: func(m *mockShortURLService) {
				m.getLongError = errors.New("database error")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"database error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockShortURLService{}
			tt.setupService(service)
			handler := NewShortURLHandler(service)

			var body []byte
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					body = []byte(str)
				} else {
					body, _ = json.Marshal(tt.body)
				}
			}

			req := httptest.NewRequest(tt.method, "/v1/getLongUrl", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.GetLongURL(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" {
				body := strings.TrimSpace(w.Body.String())
				if body != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, body)
				}
			}

			if tt.checkHeaders != nil {
				tt.checkHeaders(t, w)
			}
		})
	}
}

func TestShortURLHandler_RedirectShortURL(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		host           string
		setupService   func(*mockShortURLService)
		expectedStatus int
		expectedBody   string
		checkLocation  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful redirect",
			method: "GET",
			path:   "/abc123",
			host:   "test.com",
			setupService: func(m *mockShortURLService) {
				m.longURL = "https://example.com"
			},
			expectedStatus: http.StatusFound,
			checkLocation: func(t *testing.T, w *httptest.ResponseRecorder) {
				location := w.Header().Get("Location")
				if location != "https://example.com" {
					t.Errorf("expected Location header 'https://example.com', got %q", location)
				}
			},
		},
		{
			name:           "wrong method",
			method:         "POST",
			path:           "/abc123",
			host:           "test.com",
			setupService:   func(m *mockShortURLService) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "URL not found",
			method: "GET",
			path:   "/notfound",
			host:   "test.com",
			setupService: func(m *mockShortURLService) {
				m.getLongError = errors.New("short URL not found")
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"short URL not found"}`,
		},
		{
			name:   "service error",
			method: "GET",
			path:   "/error",
			host:   "test.com",
			setupService: func(m *mockShortURLService) {
				m.getLongError = errors.New("database error")
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"database error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockShortURLService{}
			tt.setupService(service)
			handler := NewShortURLHandler(service)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Host = tt.host
			w := httptest.NewRecorder()

			handler.RedirectShortURL(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" {
				body := strings.TrimSpace(w.Body.String())
				if body != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, body)
				}
			}

			if tt.checkLocation != nil {
				tt.checkLocation(t, w)
			}
		})
	}
}

func TestShortURLHandler_GetAllShortURLs(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		setupService   func(*mockShortURLService)
		expectedStatus int
		expectedBody   string
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful retrieval",
			method: "GET",
			setupService: func(m *mockShortURLService) {
				m.allURLs = []*app.ShortURLResponse{
					{
						ID:        "abc123",
						LongURL:   "https://example1.com",
						ShortURL:  "http://test.com/abc123",
						CreatedAt: time.Now(),
						IsActive:  true,
					},
					{
						ID:        "def456",
						LongURL:   "https://example2.com",
						ShortURL:  "http://test.com/def456",
						CreatedAt: time.Now(),
						IsActive:  true,
					},
				}
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp []*app.ShortURLResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
					return
				}
				if len(resp) != 2 {
					t.Errorf("expected 2 URLs, got %d", len(resp))
				}
			},
		},
		{
			name:           "wrong method",
			method:         "POST",
			setupService:   func(m *mockShortURLService) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "service error",
			method: "GET",
			setupService: func(m *mockShortURLService) {
				m.getAllError = errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "empty result",
			method: "GET",
			setupService: func(m *mockShortURLService) {
				m.allURLs = []*app.ShortURLResponse{}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockShortURLService{}
			tt.setupService(service)
			handler := NewShortURLHandler(service)

			req := httptest.NewRequest(tt.method, "/admin/shorturls", nil)
			w := httptest.NewRecorder()

			handler.GetAllShortURLs(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" {
				body := strings.TrimSpace(w.Body.String())
				if body != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, body)
				}
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestShortURLHandler_DeactivateShortURL(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		queryParams    string
		setupService   func(*mockShortURLService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful deactivation",
			method:      "DELETE",
			queryParams: "?id=abc123",
			setupService: func(m *mockShortURLService) {
				m.deactivateError = nil
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "wrong method",
			method:         "GET",
			queryParams:    "?id=abc123",
			setupService:   func(m *mockShortURLService) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "missing ID parameter",
			method:         "DELETE",
			queryParams:    "",
			setupService:   func(m *mockShortURLService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "service error",
			method:      "DELETE",
			queryParams: "?id=notfound",
			setupService: func(m *mockShortURLService) {
				m.deactivateError = errors.New("short URL not found")
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &mockShortURLService{}
			tt.setupService(service)
			handler := NewShortURLHandler(service)

			req := httptest.NewRequest(tt.method, "/admin/deactivate"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.DeactivateShortURL(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" {
				body := strings.TrimSpace(w.Body.String())
				if body != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, body)
				}
			}
		})
	}
}
