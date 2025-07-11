package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oharai/short-url/internal/shorturl/app"
	"github.com/oharai/short-url/internal/shorturl/infra"
	httpHandler "github.com/oharai/short-url/internal/shorturl/interfaces/http"
)

// setupTestServer creates a test server with all dependencies
func setupTestServer() *httptest.Server {
	repo := infra.NewMemoryShortURLRepository()
	kgs := infra.NewBase62KeyGenerationService()
	analytics := infra.NewMockAnalyticsService()
	baseURL := "http://test.com"

	service := app.NewShortURLService(repo, kgs, analytics, baseURL)
	handler := httpHandler.NewShortURLHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/createShortUrl", handler.CreateShortURL)
	mux.HandleFunc("/v1/getLongUrl", handler.GetLongURL)
	mux.HandleFunc("/admin/shorturls", handler.GetAllShortURLs)
	mux.HandleFunc("/admin/deactivate", handler.DeactivateShortURL)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Short URL not found"))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/v1/") || strings.HasPrefix(r.URL.Path, "/admin/") {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Endpoint not found"))
			return
		}
		handler.RedirectShortURL(w, r)
	})

	return httptest.NewServer(mux)
}

func TestIntegration_CreateAndAccessShortURL(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Test 1: Create a short URL
	createReq := app.CreateShortURLRequest{
		LongURL: "https://example.com/very/long/path",
		UserMetadata: map[string]interface{}{
			"userId": "user123",
			"source": "integration_test",
		},
	}

	body, _ := json.Marshal(createReq)
	resp, err := http.Post(server.URL+"/v1/createShortUrl", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create short URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var createResp app.CreateShortURLResponse
	if decodeErr := json.NewDecoder(resp.Body).Decode(&createResp); decodeErr != nil {
		t.Fatalf("failed to decode create response: %v", decodeErr)
	}

	if createResp.ShortURL == "" {
		t.Error("expected short URL to be returned")
	}

	// Test 2: Access the short URL via API
	getLongReq := app.GetLongURLRequest{
		ShortURL: createResp.ShortURL,
		UserMetadata: map[string]interface{}{
			"ip": "192.168.1.1",
		},
	}

	body, _ = json.Marshal(getLongReq)
	requestURL := server.URL + "/v1/getLongUrl"
	t.Logf("Making request to: %s", requestURL)
	t.Logf("Request body: %s", string(body))
	req, _ := http.NewRequest("GET", requestURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("failed to get long URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status 302, got %d, body: %s", resp.StatusCode, string(body))
	}

	location := resp.Header.Get("Location")
	if location != "https://example.com/very/long/path" {
		t.Errorf("expected location 'https://example.com/very/long/path', got %q", location)
	}

	// Test 3: Access via direct redirect
	shortID := strings.TrimPrefix(createResp.ShortURL, "http://test.com/")
	redirectResp, err := http.Get(server.URL + "/" + shortID)
	if err != nil {
		t.Fatalf("failed to access redirect: %v", err)
	}
	defer redirectResp.Body.Close()

	// Note: http.Get follows redirects by default, so we need to use a custom client
	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	redirectResp, err = client.Get(server.URL + "/" + shortID)
	if err != nil {
		t.Fatalf("failed to access redirect: %v", err)
	}
	defer redirectResp.Body.Close()

	if redirectResp.StatusCode != http.StatusFound {
		t.Errorf("expected status 302, got %d", redirectResp.StatusCode)
	}

	redirectLocation := redirectResp.Header.Get("Location")
	if redirectLocation != "https://example.com/very/long/path" {
		t.Errorf("expected redirect location 'https://example.com/very/long/path', got %q", redirectLocation)
	}
}

func TestIntegration_CustomURL(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Test 1: Create a custom short URL
	createReq := app.CreateShortURLRequest{
		LongURL:   "https://example.com/custom",
		CustomURL: "my-custom-url",
		UserMetadata: map[string]interface{}{
			"campaign": "summer2024",
		},
	}

	body, _ := json.Marshal(createReq)
	resp, err := http.Post(server.URL+"/v1/createShortUrl", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create custom short URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var createResp app.CreateShortURLResponse
	if decodeErr := json.NewDecoder(resp.Body).Decode(&createResp); decodeErr != nil {
		t.Fatalf("failed to decode create response: %v", decodeErr)
	}

	if createResp.ShortURL != "my-custom-url" {
		t.Errorf("expected short URL 'my-custom-url', got %q", createResp.ShortURL)
	}

	// Test 2: Try to create the same custom URL again (should fail)
	resp, err = http.Post(server.URL+"/v1/createShortUrl", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to make second request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected status 409, got %d", resp.StatusCode)
	}
}

func TestIntegration_AdminOperations(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Create several URLs
	urls := []string{
		"https://example1.com",
		"https://example2.com",
		"https://example3.com",
	}

	for _, url := range urls {
		createReq := app.CreateShortURLRequest{
			LongURL: url,
		}

		body, _ := json.Marshal(createReq)
		resp, err := http.Post(server.URL+"/v1/createShortUrl", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("failed to create short URL: %v", err)
		}

		var createResp app.CreateShortURLResponse
		json.NewDecoder(resp.Body).Decode(&createResp)
		resp.Body.Close()
	}

	// Test 1: Get all URLs
	resp, err := http.Get(server.URL + "/admin/shorturls")
	if err != nil {
		t.Fatalf("failed to get all URLs: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var allURLs []*app.ShortURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&allURLs); err != nil {
		t.Fatalf("failed to decode all URLs response: %v", err)
	}

	if len(allURLs) != 3 {
		t.Errorf("expected 3 URLs, got %d", len(allURLs))
	}

	// Test 2: Deactivate a URL
	if len(allURLs) > 0 {
		urlID := allURLs[0].ID
		req, _ := http.NewRequest("DELETE", server.URL+"/admin/deactivate?id="+urlID, nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to deactivate URL: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", resp.StatusCode)
		}

		// Verify URL is deactivated by trying to access it
		shortID := strings.TrimPrefix(allURLs[0].ShortURL, "http://test.com/")
		client = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		accessResp, err := client.Get(server.URL + "/" + shortID)
		if err != nil {
			t.Fatalf("failed to access deactivated URL: %v", err)
		}
		defer accessResp.Body.Close()

		// Should return error since URL is deactivated
		if accessResp.StatusCode == http.StatusFound {
			t.Error("expected deactivated URL to not redirect")
		}
	}
}

func TestIntegration_URLExpiry(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Create a URL with expiry in the past
	pastTime := time.Now().Add(-1 * time.Hour)
	createReq := app.CreateShortURLRequest{
		LongURL: "https://example.com/expired",
		Expiry:  &pastTime,
	}

	body, _ := json.Marshal(createReq)
	resp, err := http.Post(server.URL+"/v1/createShortUrl", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create expired short URL: %v", err)
	}
	defer resp.Body.Close()

	var createResp app.CreateShortURLResponse
	json.NewDecoder(resp.Body).Decode(&createResp)

	// Try to access the expired URL
	shortID := strings.TrimPrefix(createResp.ShortURL, "http://test.com/")
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	accessResp, err := client.Get(server.URL + "/" + shortID)
	if err != nil {
		t.Fatalf("failed to access expired URL: %v", err)
	}
	defer accessResp.Body.Close()

	// Should return error since URL is expired
	if accessResp.StatusCode == http.StatusFound {
		t.Error("expected expired URL to not redirect")
	}
}

func TestIntegration_ErrorCases(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Test 1: Invalid JSON
	resp, err := http.Post(server.URL+"/v1/createShortUrl", "application/json", bytes.NewReader([]byte("invalid json")))
	if err != nil {
		t.Fatalf("failed to make invalid JSON request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid JSON, got %d", resp.StatusCode)
	}

	// Test 2: Empty long URL
	createReq := app.CreateShortURLRequest{
		LongURL: "",
	}

	body, _ := json.Marshal(createReq)
	resp, err = http.Post(server.URL+"/v1/createShortUrl", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to make empty URL request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 for empty URL, got %d", resp.StatusCode)
	}

	// Test 3: Non-existent short URL
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err = client.Get(server.URL + "/nonexistent")
	if err != nil {
		t.Fatalf("failed to access non-existent URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 404 or 400 for non-existent URL, got %d", resp.StatusCode)
	}

	// Test 4: Wrong HTTP method
	req, _ := http.NewRequest("GET", server.URL+"/v1/createShortUrl", nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make wrong method request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405 for wrong method, got %d", resp.StatusCode)
	}
}

func TestIntegration_ConcurrentAccess(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Create a URL first
	createReq := app.CreateShortURLRequest{
		LongURL: "https://example.com/concurrent",
	}

	body, _ := json.Marshal(createReq)
	resp, err := http.Post(server.URL+"/v1/createShortUrl", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create short URL: %v", err)
	}
	defer resp.Body.Close()

	var createResp app.CreateShortURLResponse
	json.NewDecoder(resp.Body).Decode(&createResp)

	shortID := strings.TrimPrefix(createResp.ShortURL, "http://test.com/")

	// Concurrent access test
	done := make(chan bool, 10)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			resp, err := client.Get(server.URL + "/" + shortID)
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusFound {
				errors <- fmt.Errorf("expected status 302, got %d", resp.StatusCode)
				return
			}

			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Success
		case err := <-errors:
			t.Errorf("concurrent access error: %v", err)
		case <-time.After(5 * time.Second):
			t.Error("concurrent access test timed out")
		}
	}
}
