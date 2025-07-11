package domain

import (
	"testing"
	"time"
)

func TestNewShortURL(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		longURL     string
		shortURL    string
		expiry      *time.Time
		userMetadata map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:     "valid short URL creation",
			id:       "abc123",
			longURL:  "https://example.com",
			shortURL: "http://short.ly/abc123",
			expiry:   nil,
			userMetadata: map[string]interface{}{
				"userId": "user123",
			},
			expectError: false,
		},
		{
			name:        "empty long URL",
			id:          "abc123",
			longURL:     "",
			shortURL:    "http://short.ly/abc123",
			expectError: true,
			errorMsg:    "long URL cannot be empty",
		},
		{
			name:        "empty short URL",
			id:          "abc123",
			longURL:     "https://example.com",
			shortURL:    "",
			expectError: true,
			errorMsg:    "short URL cannot be empty",
		},
		{
			name:     "with expiry time",
			id:       "abc123",
			longURL:  "https://example.com",
			shortURL: "http://short.ly/abc123",
			expiry:   timePtr(time.Now().Add(24 * time.Hour)),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortURL, err := NewShortURL(tt.id, tt.longURL, tt.shortURL, tt.expiry, tt.userMetadata)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("expected error message %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if shortURL.ID() != tt.id {
				t.Errorf("expected ID %q, got %q", tt.id, shortURL.ID())
			}

			if shortURL.LongURL() != tt.longURL {
				t.Errorf("expected long URL %q, got %q", tt.longURL, shortURL.LongURL())
			}

			if shortURL.ShortURL() != tt.shortURL {
				t.Errorf("expected short URL %q, got %q", tt.shortURL, shortURL.ShortURL())
			}

			if !shortURL.IsActive() {
				t.Errorf("expected URL to be active")
			}

			if tt.expiry != nil && !shortURL.Expiry().Equal(*tt.expiry) {
				t.Errorf("expected expiry %v, got %v", *tt.expiry, *shortURL.Expiry())
			}

			if tt.expiry == nil && shortURL.Expiry() != nil {
				t.Errorf("expected no expiry, got %v", *shortURL.Expiry())
			}
		})
	}
}

func TestNewCustomShortURL(t *testing.T) {
	tests := []struct {
		name        string
		customURL   string
		longURL     string
		expiry      *time.Time
		userMetadata map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:      "valid custom short URL creation",
			customURL: "my-custom-url",
			longURL:   "https://example.com",
			expiry:    nil,
			userMetadata: map[string]interface{}{
				"campaign": "summer2024",
			},
			expectError: false,
		},
		{
			name:        "empty long URL",
			customURL:   "my-custom-url",
			longURL:     "",
			expectError: true,
			errorMsg:    "long URL cannot be empty",
		},
		{
			name:        "empty custom URL",
			customURL:   "",
			longURL:     "https://example.com",
			expectError: true,
			errorMsg:    "custom URL cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortURL, err := NewCustomShortURL(tt.customURL, tt.longURL, tt.expiry, tt.userMetadata)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("expected error message %q, got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if shortURL.ID() != tt.customURL {
				t.Errorf("expected ID %q, got %q", tt.customURL, shortURL.ID())
			}

			if shortURL.ShortURL() != tt.customURL {
				t.Errorf("expected short URL %q, got %q", tt.customURL, shortURL.ShortURL())
			}
		})
	}
}

func TestReconstructShortURL(t *testing.T) {
	id := "abc123"
	longURL := "https://example.com"
	shortURL := "http://short.ly/abc123"
	createdAt := time.Now()
	expiry := timePtr(time.Now().Add(24 * time.Hour))
	isActive := true
	userMetadata := map[string]interface{}{
		"userId": "user123",
	}

	reconstructed := ReconstructShortURL(id, longURL, shortURL, createdAt, expiry, isActive, userMetadata)

	if reconstructed.ID() != id {
		t.Errorf("expected ID %q, got %q", id, reconstructed.ID())
	}

	if reconstructed.LongURL() != longURL {
		t.Errorf("expected long URL %q, got %q", longURL, reconstructed.LongURL())
	}

	if reconstructed.ShortURL() != shortURL {
		t.Errorf("expected short URL %q, got %q", shortURL, reconstructed.ShortURL())
	}

	if !reconstructed.CreatedAt().Equal(createdAt) {
		t.Errorf("expected created at %v, got %v", createdAt, reconstructed.CreatedAt())
	}

	if !reconstructed.Expiry().Equal(*expiry) {
		t.Errorf("expected expiry %v, got %v", *expiry, *reconstructed.Expiry())
	}

	if reconstructed.IsActive() != isActive {
		t.Errorf("expected active %v, got %v", isActive, reconstructed.IsActive())
	}
}

func TestShortURL_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		expiry   *time.Time
		expected bool
	}{
		{
			name:     "no expiry",
			expiry:   nil,
			expected: false,
		},
		{
			name:     "future expiry",
			expiry:   timePtr(time.Now().Add(24 * time.Hour)),
			expected: false,
		},
		{
			name:     "past expiry",
			expiry:   timePtr(time.Now().Add(-24 * time.Hour)),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortURL, _ := NewShortURL("abc123", "https://example.com", "http://short.ly/abc123", tt.expiry, nil)

			if shortURL.IsExpired() != tt.expected {
				t.Errorf("expected IsExpired() to return %v, got %v", tt.expected, shortURL.IsExpired())
			}
		})
	}
}

func TestShortURL_Deactivate(t *testing.T) {
	shortURL, _ := NewShortURL("abc123", "https://example.com", "http://short.ly/abc123", nil, nil)

	if !shortURL.IsActive() {
		t.Errorf("expected URL to be active initially")
	}

	shortURL.Deactivate()

	if shortURL.IsActive() {
		t.Errorf("expected URL to be inactive after deactivation")
	}
}

func TestShortURL_Getters(t *testing.T) {
	id := "abc123"
	longURL := "https://example.com"
	shortURL := "http://short.ly/abc123"
	expiry := timePtr(time.Now().Add(24 * time.Hour))
	userMetadata := map[string]interface{}{
		"userId": "user123",
		"source": "api",
	}

	url, _ := NewShortURL(id, longURL, shortURL, expiry, userMetadata)

	if url.ID() != id {
		t.Errorf("ID() returned %q, expected %q", url.ID(), id)
	}

	if url.LongURL() != longURL {
		t.Errorf("LongURL() returned %q, expected %q", url.LongURL(), longURL)
	}

	if url.ShortURL() != shortURL {
		t.Errorf("ShortURL() returned %q, expected %q", url.ShortURL(), shortURL)
	}

	if url.Expiry() == nil || !url.Expiry().Equal(*expiry) {
		t.Errorf("Expiry() returned %v, expected %v", url.Expiry(), expiry)
	}

	if !url.IsActive() {
		t.Errorf("IsActive() returned false, expected true")
	}

	metadata := url.UserMetadata()
	if metadata["userId"] != "user123" {
		t.Errorf("UserMetadata()[\"userId\"] returned %v, expected \"user123\"", metadata["userId"])
	}

	if metadata["source"] != "api" {
		t.Errorf("UserMetadata()[\"source\"] returned %v, expected \"api\"", metadata["source"])
	}
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}