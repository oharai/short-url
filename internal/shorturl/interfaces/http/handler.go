// Package http contains the HTTP presentation layer that handles REST API requests
// and translates them into application service calls. This layer is responsible
// for HTTP protocol concerns, request validation, and response formatting.
package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/oharai/short-url/internal/shorturl/app"
)

// ShortURLServiceInterface defines the interface for the application service
type ShortURLServiceInterface interface {
	CreateShortURL(req app.CreateShortURLRequest) (*app.CreateShortURLResponse, error)
	GetLongURL(req app.GetLongURLRequest) (string, error)
	GetAllShortURLs() ([]*app.ShortURLResponse, error)
	DeactivateShortURL(id string) error
}

// ShortURLHandler handles HTTP requests for the URL shortening service.
// It acts as the presentation layer, converting HTTP requests into application
// service calls and formatting responses according to REST API conventions.
type ShortURLHandler struct {
	service ShortURLServiceInterface // Application service for business logic operations
}

// NewShortURLHandler creates a new HTTP handler with the provided application service.
// This constructor follows the dependency injection pattern to ensure testability.
//
// Parameters:
//   - service: The application service that handles business logic
//
// Returns:
//   - *ShortURLHandler: Configured HTTP handler ready to process requests
func NewShortURLHandler(service ShortURLServiceInterface) *ShortURLHandler {
	return &ShortURLHandler{
		service: service,
	}
}

// CreateShortURL handles POST /v1/createShortUrl requests.
// This endpoint accepts a JSON payload with the long URL and optional parameters,
// creates a short URL, and returns the result according to the API specification.
//
// Request Format:
//   - Method: POST
//   - Content-Type: application/json
//   - Body: CreateShortURLRequest JSON
//
// Response Format:
//   - Success: 200 OK with CreateShortURLResponse JSON
//   - Error: 400/409/500 with error message JSON
func (h *ShortURLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	// Validate HTTP method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse and validate request body
	var req app.CreateShortURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"}); encodeErr != nil {
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	// Process request through application service
	resp, err := h.service.CreateShortURL(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		// Determine appropriate HTTP status code based on error type
		if strings.Contains(err.Error(), "already exists") {
			w.WriteHeader(http.StatusConflict)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); encodeErr != nil {
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	// Return successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetLongURL handles GET /v1/getLongUrl requests.
// This endpoint accepts a JSON payload with the short URL to resolve,
// retrieves the original URL, and returns a 302 redirect response.
//
// Request Format:
//   - Method: GET
//   - Content-Type: application/json
//   - Body: GetLongURLRequest JSON
//
// Response Format:
//   - Success: 302 Found with Location header
//   - Error: 400/404 with error message JSON
func (h *ShortURLHandler) GetLongURL(w http.ResponseWriter, r *http.Request) {
	// Validate HTTP method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse and validate request body
	var req app.GetLongURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"}); encodeErr != nil {
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	// Resolve short URL through application service
	longURL, err := h.service.GetLongURL(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		// Determine appropriate HTTP status code based on error type
		if strings.Contains(err.Error(), "not found") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); encodeErr != nil {
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	// Return redirect response as specified in the API
	w.Header().Set("Location", longURL)
	w.WriteHeader(http.StatusFound)
}

// RedirectShortURL handles GET /<shortId> requests for direct URL redirection.
// This endpoint extracts the short URL identifier from the path, resolves it
// to the original URL, tracks the access event, and performs an HTTP redirect.
//
// Request Format:
//   - Method: GET
//   - Path: /<shortId>
//   - No body required
//
// Response Format:
//   - Success: 302 Found redirect to original URL
//   - Error: 404/400 with error message JSON
func (h *ShortURLHandler) RedirectShortURL(w http.ResponseWriter, r *http.Request) {
	// Validate HTTP method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Reconstruct the complete short URL from the request
	shortURL := r.Host + r.URL.Path
	if r.URL.Scheme != "" {
		shortURL = r.URL.Scheme + "://" + shortURL
	} else {
		shortURL = "http://" + shortURL
	}

	// Prepare request with user context for analytics tracking
	req := app.GetLongURLRequest{
		ShortURL: shortURL,
		UserMetadata: map[string]interface{}{
			"ip":         r.RemoteAddr,  // Client IP address
			"user_agent": r.UserAgent(), // Browser/client information
			"referer":    r.Referer(),   // Referring page
		},
	}

	// Resolve short URL through application service
	longURL, err := h.service.GetLongURL(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		// Determine appropriate HTTP status code based on error type
		if strings.Contains(err.Error(), "not found") {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		if encodeErr := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); encodeErr != nil {
			http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		}
		return
	}

	// Perform HTTP redirect to the original URL
	http.Redirect(w, r, longURL, http.StatusFound)
}

// GetAllShortURLs handles GET /admin/shorturls requests.
// This administrative endpoint retrieves all short URLs for management purposes.
// It returns a JSON array containing complete information about all URLs.
//
// Request Format:
//   - Method: GET
//   - Path: /admin/shorturls
//   - No body required
//
// Response Format:
//   - Success: 200 OK with ShortURLResponse array JSON
//   - Error: 500 with error message
func (h *ShortURLHandler) GetAllShortURLs(w http.ResponseWriter, r *http.Request) {
	// Validate HTTP method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve all URLs through application service
	shortURLs, err := h.service.GetAllShortURLs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return JSON response with all URLs
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(shortURLs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// DeactivateShortURL handles DELETE /admin/deactivate requests.
// This administrative endpoint deactivates a short URL by its identifier.
// The ID is provided as a query parameter.
//
// Request Format:
//   - Method: DELETE
//   - Path: /admin/deactivate?id=<url_id>
//   - No body required
//
// Response Format:
//   - Success: 204 No Content
//   - Error: 400/404 with error message
func (h *ShortURLHandler) DeactivateShortURL(w http.ResponseWriter, r *http.Request) {
	// Validate HTTP method
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract and validate ID parameter
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	// Deactivate URL through application service
	if err := h.service.DeactivateShortURL(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Return successful response with no content
	w.WriteHeader(http.StatusNoContent)
}
