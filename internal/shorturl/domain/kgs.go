package domain

// KeyGenerationService defines the contract for generating unique identifiers
// for short URLs. This interface enables different ID generation strategies
// to be implemented (Base62, UUID, etc.) while maintaining consistency.
type KeyGenerationService interface {
	// GenerateUniqueID creates a single unique identifier for a short URL.
	// The generated ID should be URL-safe and follow the configured format.
	GenerateUniqueID() (string, error)

	// GetMultipleIDs generates multiple unique identifiers in a single operation.
	// This is useful for performance optimization when many IDs are needed.
	// The count parameter specifies how many IDs to generate.
	GetMultipleIDs(count int) ([]string, error)
}