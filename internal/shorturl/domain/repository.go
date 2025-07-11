package domain

// ShortURLRepository defines the contract for persisting ShortURL entities.
// This interface abstracts the data access layer and enables dependency inversion,
// allowing different implementations (memory, database, etc.) to be used.
type ShortURLRepository interface {
	// Save persists a ShortURL entity to the data store.
	// If the entity already exists, it will be updated.
	Save(shortURL *ShortURL) error

	// FindByID retrieves a ShortURL entity by its unique identifier.
	// Returns nil if the entity is not found.
	FindByID(id string) (*ShortURL, error)

	// FindAll retrieves all ShortURL entities from the data store.
	// Returns an empty slice if no entities are found.
	FindAll() ([]*ShortURL, error)

	// Delete removes a ShortURL entity from the data store by its ID.
	// Returns an error if the entity is not found.
	Delete(id string) error
}
