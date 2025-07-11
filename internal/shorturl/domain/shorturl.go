// Package domain contains the core business logic and entities for the URL shortening service.
package domain

import (
	"errors"
	"time"
)

// ShortURL represents the core entity of the URL shortening service.
// It encapsulates all business rules related to URL shortening and management.
type ShortURL struct {
	id           string                 // Unique identifier for the short URL
	longURL      string                 // Original long URL to be shortened
	shortURL     string                 // Complete short URL including domain
	createdAt    time.Time              // Timestamp when the URL was created
	expiry       *time.Time             // Optional expiration time for the URL
	isActive     bool                   // Flag indicating if the URL is currently active
	userMetadata map[string]interface{} // Additional user-defined metadata
}

// NewShortURL creates a new ShortURL entity with automatically generated ID.
// It validates the input parameters and initializes the entity with default values.
//
// Parameters:
//   - id: Unique identifier for the short URL
//   - longURL: Original URL to be shortened
//   - shortURL: Complete short URL string
//   - expiry: Optional expiration time
//   - userMetadata: Additional metadata associated with the URL
//
// Returns:
//   - *ShortURL: The created entity
//   - error: Validation error if any required field is empty
func NewShortURL(id, longURL, shortURL string, expiry *time.Time, userMetadata map[string]interface{}) (*ShortURL, error) {
	if longURL == "" {
		return nil, errors.New("long URL cannot be empty")
	}
	if shortURL == "" {
		return nil, errors.New("short URL cannot be empty")
	}

	return &ShortURL{
		id:           id,
		longURL:      longURL,
		shortURL:     shortURL,
		createdAt:    time.Now(),
		expiry:       expiry,
		isActive:     true,
		userMetadata: userMetadata,
	}, nil
}

// NewCustomShortURL creates a new ShortURL entity with a custom identifier.
// This is used when users want to specify their own custom short URL.
//
// Parameters:
//   - customURL: User-defined custom identifier
//   - longURL: Original URL to be shortened
//   - expiry: Optional expiration time
//   - userMetadata: Additional metadata associated with the URL
//
// Returns:
//   - *ShortURL: The created entity
//   - error: Validation error if any required field is empty
func NewCustomShortURL(customURL, longURL string, expiry *time.Time, userMetadata map[string]interface{}) (*ShortURL, error) {
	if longURL == "" {
		return nil, errors.New("long URL cannot be empty")
	}
	if customURL == "" {
		return nil, errors.New("custom URL cannot be empty")
	}

	return &ShortURL{
		id:           customURL,
		longURL:      longURL,
		shortURL:     customURL,
		createdAt:    time.Now(),
		expiry:       expiry,
		isActive:     true,
		userMetadata: userMetadata,
	}, nil
}

// ReconstructShortURL recreates a ShortURL entity from stored data.
// This is typically used when loading entities from a data store.
//
// Parameters:
//   - id: Unique identifier
//   - longURL: Original URL
//   - shortURL: Complete short URL string
//   - createdAt: Creation timestamp
//   - expiry: Optional expiration time
//   - isActive: Current active status
//   - userMetadata: Associated metadata
//
// Returns:
//   - *ShortURL: The reconstructed entity
func ReconstructShortURL(id, longURL, shortURL string, createdAt time.Time, expiry *time.Time, isActive bool, userMetadata map[string]interface{}) *ShortURL {
	return &ShortURL{
		id:           id,
		longURL:      longURL,
		shortURL:     shortURL,
		createdAt:    createdAt,
		expiry:       expiry,
		isActive:     isActive,
		userMetadata: userMetadata,
	}
}

// ID returns the unique identifier of the short URL.
func (s *ShortURL) ID() string {
	return s.id
}

// LongURL returns the original long URL that was shortened.
func (s *ShortURL) LongURL() string {
	return s.longURL
}

// ShortURL returns the complete short URL string including the domain.
func (s *ShortURL) ShortURL() string {
	return s.shortURL
}

// CreatedAt returns the timestamp when the URL was created.
func (s *ShortURL) CreatedAt() time.Time {
	return s.createdAt
}

// Expiry returns the expiration time of the URL, if set.
// Returns nil if no expiration is configured.
func (s *ShortURL) Expiry() *time.Time {
	return s.expiry
}

// IsActive returns whether the URL is currently active and can be used for redirection.
func (s *ShortURL) IsActive() bool {
	return s.isActive
}

// UserMetadata returns the additional metadata associated with the URL.
func (s *ShortURL) UserMetadata() map[string]interface{} {
	return s.userMetadata
}

// IsExpired checks if the URL has passed its expiration time.
// Returns false if no expiration time is set.
func (s *ShortURL) IsExpired() bool {
	if s.expiry == nil {
		return false
	}
	return time.Now().After(*s.expiry)
}

// Deactivate marks the URL as inactive, preventing it from being used for redirection.
// This is a business operation that implements the URL deactivation use case.
func (s *ShortURL) Deactivate() {
	s.isActive = false
}