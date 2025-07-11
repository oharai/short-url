package infra

import (
	"testing"

	"github.com/oharai/short-url/internal/shorturl/domain"
)

func TestNewMemoryShortURLRepository(t *testing.T) {
	repo := NewMemoryShortURLRepository()
	if repo == nil {
		t.Error("expected repository to be created")
	}
}

func TestMemoryShortURLRepository_Save(t *testing.T) {
	repo := NewMemoryShortURLRepository().(*MemoryShortURLRepository)

	shortURL, _ := domain.NewShortURL("abc123", "https://example.com", "http://short.ly/abc123", nil, nil)

	err := repo.Save(shortURL)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify URL was saved
	if len(repo.data) != 1 {
		t.Errorf("expected 1 URL in repository, got %d", len(repo.data))
	}

	saved, exists := repo.data["abc123"]
	if !exists {
		t.Error("expected URL to be saved")
	}

	if saved.ID() != "abc123" {
		t.Errorf("expected ID 'abc123', got %q", saved.ID())
	}
}

func TestMemoryShortURLRepository_Save_Update(t *testing.T) {
	repo := NewMemoryShortURLRepository().(*MemoryShortURLRepository)

	// Save initial URL
	shortURL, _ := domain.NewShortURL("abc123", "https://example.com", "http://short.ly/abc123", nil, nil)
	repo.Save(shortURL)

	// Update the URL (deactivate it)
	shortURL.Deactivate()
	err := repo.Save(shortURL)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify update
	saved := repo.data["abc123"]
	if saved.IsActive() {
		t.Error("expected URL to be deactivated")
	}
}

func TestMemoryShortURLRepository_FindByID(t *testing.T) {
	repo := NewMemoryShortURLRepository().(*MemoryShortURLRepository)

	// Save a URL
	shortURL, _ := domain.NewShortURL("abc123", "https://example.com", "http://short.ly/abc123", nil, nil)
	repo.Save(shortURL)

	tests := []struct {
		name      string
		id        string
		expectNil bool
		expectErr bool
	}{
		{
			name:      "existing URL",
			id:        "abc123",
			expectNil: false,
			expectErr: false,
		},
		{
			name:      "non-existing URL",
			id:        "notfound",
			expectNil: true,
			expectErr: false,
		},
		{
			name:      "empty ID",
			id:        "",
			expectNil: true,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, err := repo.FindByID(tt.id)

			if tt.expectErr && err == nil {
				t.Error("expected error but got none")
				return
			}

			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectNil && found != nil {
				t.Error("expected nil but got URL")
				return
			}

			if !tt.expectNil && found == nil {
				t.Error("expected URL but got nil")
				return
			}

			if found != nil && found.ID() != tt.id {
				t.Errorf("expected ID %q, got %q", tt.id, found.ID())
			}
		})
	}
}

func TestMemoryShortURLRepository_FindAll(t *testing.T) {
	repo := NewMemoryShortURLRepository().(*MemoryShortURLRepository)

	tests := []struct {
		name        string
		setupURLs   int
		expectCount int
	}{
		{
			name:        "empty repository",
			setupURLs:   0,
			expectCount: 0,
		},
		{
			name:        "single URL",
			setupURLs:   1,
			expectCount: 1,
		},
		{
			name:        "multiple URLs",
			setupURLs:   3,
			expectCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear repository
			repo.data = make(map[string]*domain.ShortURL)

			// Setup URLs
			for i := 0; i < tt.setupURLs; i++ {
				id := "url" + string(rune('0'+i))
				url := "https://example" + string(rune('0'+i)) + ".com"
				shortURL, _ := domain.NewShortURL(id, url, "http://short.ly/"+id, nil, nil)
				repo.Save(shortURL)
			}

			urls, err := repo.FindAll()
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

func TestMemoryShortURLRepository_Delete(t *testing.T) {
	repo := NewMemoryShortURLRepository().(*MemoryShortURLRepository)

	// Save a URL
	shortURL, _ := domain.NewShortURL("abc123", "https://example.com", "http://short.ly/abc123", nil, nil)
	repo.Save(shortURL)

	tests := []struct {
		name        string
		id          string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "delete existing URL",
			id:          "abc123",
			expectError: false,
		},
		{
			name:        "delete non-existing URL",
			id:          "notfound",
			expectError: true,
			errorMsg:    "short URL not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset repository state for each test
			if tt.name == "delete existing URL" {
				repo.data = make(map[string]*domain.ShortURL)
				shortURL, _ := domain.NewShortURL("abc123", "https://example.com", "http://short.ly/abc123", nil, nil)
				repo.Save(shortURL)
			}

			err := repo.Delete(tt.id)

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

			// Verify URL was deleted
			if _, exists := repo.data[tt.id]; exists {
				t.Error("expected URL to be deleted")
			}
		})
	}
}

func TestMemoryShortURLRepository_ConcurrentAccess(t *testing.T) {
	repo := NewMemoryShortURLRepository().(*MemoryShortURLRepository)

	// Test concurrent saves
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			id := "url" + string(rune('0'+index))
			url := "https://example" + string(rune('0'+index)) + ".com"
			shortURL, _ := domain.NewShortURL(id, url, "http://short.ly/"+id, nil, nil)
			repo.Save(shortURL)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all URLs were saved
	urls, _ := repo.FindAll()
	if len(urls) != 10 {
		t.Errorf("expected 10 URLs after concurrent saves, got %d", len(urls))
	}

	// Test concurrent reads
	for i := 0; i < 10; i++ {
		go func(index int) {
			id := "url" + string(rune('0'+index))
			found, _ := repo.FindByID(id)
			if found == nil {
				t.Errorf("expected to find URL with ID %q", id)
			}
			done <- true
		}(i)
	}

	// Wait for all read operations to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
