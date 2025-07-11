// Package infra contains the infrastructure layer implementations
// that handle external concerns like data persistence, key generation, and analytics.
package infra

import (
	"errors"
	"sync"

	"github.com/oharai/short-url/internal/shorturl/domain"
)

// MemoryShortURLRepository is an in-memory implementation of the ShortURLRepository interface.
// It provides a simple, thread-safe storage solution suitable for development, testing,
// and small-scale deployments. Data is lost when the application restarts.
type MemoryShortURLRepository struct {
	mu   sync.RWMutex                // Read-write mutex for concurrent access safety
	data map[string]*domain.ShortURL // In-memory storage keyed by URL ID
}

// NewMemoryShortURLRepository creates a new instance of the in-memory repository.
// It initializes the internal data structures and returns the interface implementation.
//
// Returns:
//   - domain.ShortURLRepository: Repository interface implementation
func NewMemoryShortURLRepository() domain.ShortURLRepository {
	return &MemoryShortURLRepository{
		data: make(map[string]*domain.ShortURL),
	}
}

// Save persists a ShortURL entity to memory storage.
// Uses write lock to ensure thread safety during updates.
//
// Parameters:
//   - shortURL: The entity to save or update
//
// Returns:
//   - error: Always nil for this implementation
func (r *MemoryShortURLRepository) Save(shortURL *domain.ShortURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[shortURL.ID()] = shortURL
	return nil
}

// FindByID retrieves a ShortURL entity by its unique identifier.
// Uses read lock to allow concurrent reads while maintaining data integrity.
//
// Parameters:
//   - id: The unique identifier to search for
//
// Returns:
//   - *domain.ShortURL: The found entity or nil if not found
//   - error: Always nil for this implementation
func (r *MemoryShortURLRepository) FindByID(id string) (*domain.ShortURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shortURL, exists := r.data[id]
	if !exists {
		return nil, nil
	}

	return shortURL, nil
}

// FindAll retrieves all ShortURL entities from memory storage.
// Returns a slice containing all stored entities. Order is not guaranteed.
//
// Returns:
//   - []*domain.ShortURL: Slice of all stored entities
//   - error: Always nil for this implementation
func (r *MemoryShortURLRepository) FindAll() ([]*domain.ShortURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var shortURLs []*domain.ShortURL
	for _, shortURL := range r.data {
		shortURLs = append(shortURLs, shortURL)
	}

	return shortURLs, nil
}

// Delete removes a ShortURL entity from memory storage by its identifier.
// Uses write lock to ensure thread safety during deletion.
//
// Parameters:
//   - id: The unique identifier of the entity to delete
//
// Returns:
//   - error: Error if the entity was not found
func (r *MemoryShortURLRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[id]; !exists {
		return errors.New("short URL not found")
	}

	delete(r.data, id)
	return nil
}
