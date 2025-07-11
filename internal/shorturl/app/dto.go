// Package app contains the application layer which orchestrates domain objects
// and coordinates application services to fulfill business use cases.
package app

import "time"

// CreateShortURLRequest represents the input data for creating a new short URL.
// It supports both automatic ID generation and custom URL specification.
type CreateShortURLRequest struct {
	LongURL      string                 `json:"longUrl"`                // The original URL to be shortened (required)
	CustomURL    string                 `json:"customUrl,omitempty"`    // Optional custom identifier for the short URL
	Expiry       *time.Time             `json:"expiry,omitempty"`       // Optional expiration time for the URL
	UserMetadata map[string]interface{} `json:"userMetadata,omitempty"` // Additional metadata for analytics
}

// CreateShortURLResponse contains the result of a successful short URL creation.
type CreateShortURLResponse struct {
	ShortURL string `json:"shortUrl"` // The complete short URL that was created
}

// GetLongURLRequest represents the input data for retrieving the original URL.
// Used in the redirect flow to resolve short URLs to their destinations.
type GetLongURLRequest struct {
	ShortURL     string                 `json:"shortUrl"`               // The short URL to resolve (required)
	UserMetadata map[string]interface{} `json:"userMetadata,omitempty"` // Context data for analytics tracking
}

// ShortURLResponse represents the complete information about a short URL.
// Used in administrative operations and listing endpoints.
type ShortURLResponse struct {
	ID           string                 `json:"id"`                     // Unique identifier
	LongURL      string                 `json:"longUrl"`                // Original URL
	ShortURL     string                 `json:"shortUrl"`               // Complete short URL
	CreatedAt    time.Time              `json:"createdAt"`              // Creation timestamp
	Expiry       *time.Time             `json:"expiry,omitempty"`       // Optional expiration time
	IsActive     bool                   `json:"isActive"`               // Current status
	UserMetadata map[string]interface{} `json:"userMetadata,omitempty"` // Associated metadata
}
