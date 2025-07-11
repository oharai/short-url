package app

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/oharai/short-url/internal/shorturl/domain"
)

// Mock implementations for testing

type mockRepository struct {
	data    map[string]*domain.ShortURL
	saveErr error
	findErr error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		data: make(map[string]*domain.ShortURL),
	}
}

func (m *mockRepository) Save(shortURL *domain.ShortURL) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.data[shortURL.ID()] = shortURL
	return nil
}

func (m *mockRepository) FindByID(id string) (*domain.ShortURL, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	url, exists := m.data[id]
	if !exists {
		return nil, nil
	}
	return url, nil
}

func (m *mockRepository) FindAll() ([]*domain.ShortURL, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var urls []*domain.ShortURL
	for _, url := range m.data {
		urls = append(urls, url)
	}
	return urls, nil
}

func (m *mockRepository) Delete(id string) error {
	delete(m.data, id)
	return nil
}

type mockKGS struct {
	counter int
	genErr  error
}

func newMockKGS() *mockKGS {
	return &mockKGS{counter: 1}
}

func (m *mockKGS) GenerateUniqueID() (string, error) {
	if m.genErr != nil {
		return "", m.genErr
	}
	id := "test" + strings.Repeat("0", 3) + string(rune('0'+m.counter))
	m.counter++
	return id, nil
}

func (m *mockKGS) GetMultipleIDs(count int) ([]string, error) {
	if m.genErr != nil {
		return nil, m.genErr
	}
	var ids []string
	for i := 0; i < count; i++ {
		id, _ := m.GenerateUniqueID()
		ids = append(ids, id)
	}
	return ids, nil
}

type mockAnalytics struct {
	events []domain.AnalyticsEvent
	sendErr error
}

func newMockAnalytics() *mockAnalytics {
	return &mockAnalytics{
		events: make([]domain.AnalyticsEvent, 0),
	}
}

func (m *mockAnalytics) SendEvent(event domain.AnalyticsEvent) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.events = append(m.events, event)
	return nil
}

func TestNewShortURLService(t *testing.T) {
	repo := newMockRepository()
	kgs := newMockKGS()
	analytics := newMockAnalytics()
	baseURL := "http://test.com"

	service := NewShortURLService(repo, kgs, analytics, baseURL)

	if service == nil {
		t.Error("expected service to be created")
	}

	if service.baseURL != baseURL {
		t.Errorf("expected baseURL %q, got %q", baseURL, service.baseURL)
	}
}

func TestShortURLService_CreateShortURL(t *testing.T) {
	tests := []struct {
		name        string
		request     CreateShortURLRequest
		setupMocks  func(*mockRepository, *mockKGS, *mockAnalytics)
		expectError bool
		errorMsg    string
		validate    func(*testing.T, *CreateShortURLResponse, *mockRepository, *mockAnalytics)
	}{
		{
			name: "successful creation with auto-generated ID",
			request: CreateShortURLRequest{
				LongURL: "https://example.com",
				UserMetadata: map[string]interface{}{
					"userId": "user123",
				},
			},
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {},
			expectError: false,
			validate: func(t *testing.T, resp *CreateShortURLResponse, repo *mockRepository, analytics *mockAnalytics) {
				if resp == nil {
					t.Error("expected response")
					return
				}
				if !strings.Contains(resp.ShortURL, "http://test.com/") {
					t.Errorf("expected short URL to contain base URL, got %q", resp.ShortURL)
				}
				if len(analytics.events) != 1 {
					t.Errorf("expected 1 analytics event, got %d", len(analytics.events))
				}
				if analytics.events[0].EventType != "url_created" {
					t.Errorf("expected event type 'url_created', got %q", analytics.events[0].EventType)
				}
			},
		},
		{
			name: "successful creation with custom URL",
			request: CreateShortURLRequest{
				LongURL:   "https://example.com",
				CustomURL: "my-custom-url",
			},
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {},
			expectError: false,
			validate: func(t *testing.T, resp *CreateShortURLResponse, repo *mockRepository, analytics *mockAnalytics) {
				if resp.ShortURL != "my-custom-url" {
					t.Errorf("expected short URL 'my-custom-url', got %q", resp.ShortURL)
				}
			},
		},
		{
			name: "empty long URL",
			request: CreateShortURLRequest{
				LongURL: "",
			},
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {},
			expectError: true,
			errorMsg:   "longUrl is required",
		},
		{
			name: "custom URL already exists",
			request: CreateShortURLRequest{
				LongURL:   "https://example.com",
				CustomURL: "existing-url",
			},
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {
				existingURL, _ := domain.NewCustomShortURL("existing-url", "https://old.com", nil, nil)
				repo.Save(existingURL)
			},
			expectError: true,
			errorMsg:   "custom URL already exists",
		},
		{
			name: "KGS generation error",
			request: CreateShortURLRequest{
				LongURL: "https://example.com",
			},
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {
				kgs.genErr = errors.New("generation failed")
			},
			expectError: true,
			errorMsg:   "failed to generate unique ID: generation failed",
		},
		{
			name: "repository save error",
			request: CreateShortURLRequest{
				LongURL: "https://example.com",
			},
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {
				repo.saveErr = errors.New("save failed")
			},
			expectError: true,
			errorMsg:   "save failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			kgs := newMockKGS()
			analytics := newMockAnalytics()
			service := NewShortURLService(repo, kgs, analytics, "http://test.com")

			tt.setupMocks(repo, kgs, analytics)

			resp, err := service.CreateShortURL(tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
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

			if tt.validate != nil {
				tt.validate(t, resp, repo, analytics)
			}
		})
	}
}

func TestShortURLService_GetLongURL(t *testing.T) {
	tests := []struct {
		name        string
		request     GetLongURLRequest
		setupMocks  func(*mockRepository, *mockKGS, *mockAnalytics)
		expectError bool
		errorMsg    string
		expectedURL string
	}{
		{
			name: "successful retrieval",
			request: GetLongURLRequest{
				ShortURL: "http://test.com/abc123",
				UserMetadata: map[string]interface{}{
					"ip": "192.168.1.1",
				},
			},
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {
				shortURL, _ := domain.NewShortURL("abc123", "https://example.com", "http://test.com/abc123", nil, nil)
				repo.Save(shortURL)
			},
			expectError: false,
			expectedURL: "https://example.com",
		},
		{
			name: "empty short URL",
			request: GetLongURLRequest{
				ShortURL: "",
			},
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {},
			expectError: true,
			errorMsg:   "shortUrl is required",
		},
		{
			name: "URL not found",
			request: GetLongURLRequest{
				ShortURL: "http://test.com/notfound",
			},
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {},
			expectError: true,
			errorMsg:   "short URL not found",
		},
		{
			name: "inactive URL",
			request: GetLongURLRequest{
				ShortURL: "http://test.com/inactive",
			},
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {
				shortURL, _ := domain.NewShortURL("inactive", "https://example.com", "http://test.com/inactive", nil, nil)
				shortURL.Deactivate()
				repo.Save(shortURL)
			},
			expectError: true,
			errorMsg:   "short URL is not active or expired",
		},
		{
			name: "expired URL",
			request: GetLongURLRequest{
				ShortURL: "http://test.com/expired",
			},
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {
				expiry := time.Now().Add(-24 * time.Hour)
				shortURL, _ := domain.NewShortURL("expired", "https://example.com", "http://test.com/expired", &expiry, nil)
				repo.Save(shortURL)
			},
			expectError: true,
			errorMsg:   "short URL is not active or expired",
		},
		{
			name: "repository error",
			request: GetLongURLRequest{
				ShortURL: "http://test.com/error",
			},
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {
				repo.findErr = errors.New("database error")
			},
			expectError: true,
			errorMsg:   "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			kgs := newMockKGS()
			analytics := newMockAnalytics()
			service := NewShortURLService(repo, kgs, analytics, "http://test.com")

			tt.setupMocks(repo, kgs, analytics)

			longURL, err := service.GetLongURL(tt.request)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
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

			if longURL != tt.expectedURL {
				t.Errorf("expected URL %q, got %q", tt.expectedURL, longURL)
			}

			// Check analytics event was sent
			if len(analytics.events) != 1 {
				t.Errorf("expected 1 analytics event, got %d", len(analytics.events))
			} else if analytics.events[0].EventType != "url_accessed" {
				t.Errorf("expected event type 'url_accessed', got %q", analytics.events[0].EventType)
			}
		})
	}
}

func TestShortURLService_GetAllShortURLs(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mockRepository, *mockKGS, *mockAnalytics)
		expectError bool
		expectCount int
	}{
		{
			name: "successful retrieval",
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {
				url1, _ := domain.NewShortURL("abc123", "https://example1.com", "http://test.com/abc123", nil, nil)
				url2, _ := domain.NewShortURL("def456", "https://example2.com", "http://test.com/def456", nil, nil)
				repo.Save(url1)
				repo.Save(url2)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name: "empty repository",
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {},
			expectError: false,
			expectCount: 0,
		},
		{
			name: "repository error",
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {
				repo.findErr = errors.New("database error")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			kgs := newMockKGS()
			analytics := newMockAnalytics()
			service := NewShortURLService(repo, kgs, analytics, "http://test.com")

			tt.setupMocks(repo, kgs, analytics)

			urls, err := service.GetAllShortURLs()

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(urls) != tt.expectCount {
				t.Errorf("expected %d URLs, got %d", tt.expectCount, len(urls))
			}
		})
	}
}

func TestShortURLService_DeactivateShortURL(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		setupMocks  func(*mockRepository, *mockKGS, *mockAnalytics)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful deactivation",
			id:   "abc123",
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {
				shortURL, _ := domain.NewShortURL("abc123", "https://example.com", "http://test.com/abc123", nil, nil)
				repo.Save(shortURL)
			},
			expectError: false,
		},
		{
			name:        "empty ID",
			id:          "",
			setupMocks:  func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {},
			expectError: true,
			errorMsg:    "ID is required",
		},
		{
			name:        "URL not found",
			id:          "notfound",
			setupMocks:  func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {},
			expectError: true,
			errorMsg:    "short URL not found",
		},
		{
			name: "repository find error",
			id:   "abc123",
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {
				repo.findErr = errors.New("database error")
			},
			expectError: true,
			errorMsg:    "database error",
		},
		{
			name: "repository save error",
			id:   "abc123",
			setupMocks: func(repo *mockRepository, kgs *mockKGS, analytics *mockAnalytics) {
				shortURL, _ := domain.NewShortURL("abc123", "https://example.com", "http://test.com/abc123", nil, nil)
				repo.data["abc123"] = shortURL
				repo.findErr = nil // Allow find to succeed
				repo.saveErr = errors.New("save failed")
			},
			expectError: true,
			errorMsg:    "save failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			kgs := newMockKGS()
			analytics := newMockAnalytics()
			service := NewShortURLService(repo, kgs, analytics, "http://test.com")

			tt.setupMocks(repo, kgs, analytics)

			err := service.DeactivateShortURL(tt.id)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
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

			// Check that URL was deactivated
			if url, exists := repo.data[tt.id]; exists && url.IsActive() {
				t.Error("expected URL to be deactivated")
			}

			// Check analytics event was sent
			if len(analytics.events) != 1 {
				t.Errorf("expected 1 analytics event, got %d", len(analytics.events))
			} else if analytics.events[0].EventType != "url_deactivated" {
				t.Errorf("expected event type 'url_deactivated', got %q", analytics.events[0].EventType)
			}
		})
	}
}

func TestShortURLService_buildShortURL(t *testing.T) {
	service := &ShortURLService{baseURL: "http://test.com"}

	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "simple ID",
			id:       "abc123",
			expected: "http://test.com/abc123",
		},
		{
			name:     "base URL with trailing slash",
			id:       "def456",
			expected: "http://test.com/def456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "base URL with trailing slash" {
				service.baseURL = "http://test.com/"
			}

			result := service.buildShortURL(tt.id)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestShortURLService_extractIDFromShortURL(t *testing.T) {
	service := &ShortURLService{}

	tests := []struct {
		name     string
		shortURL string
		expected string
	}{
		{
			name:     "full URL",
			shortURL: "http://test.com/abc123",
			expected: "abc123",
		},
		{
			name:     "just ID",
			shortURL: "abc123",
			expected: "abc123",
		},
		{
			name:     "URL with path",
			shortURL: "http://test.com/path/def456",
			expected: "def456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.extractIDFromShortURL(tt.shortURL)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}