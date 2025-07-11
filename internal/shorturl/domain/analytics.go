package domain

import "time"

// AnalyticsEvent represents an event that occurred in the URL shortening system.
// These events are used for tracking user behavior, system performance, and business metrics.
type AnalyticsEvent struct {
	EventType    string                 `json:"eventType"`    // Type of event (url_created, url_accessed, etc.)
	ShortURL     string                 `json:"shortUrl"`     // The short URL involved in the event
	LongURL      string                 `json:"longUrl"`      // The corresponding long URL
	UserMetadata map[string]interface{} `json:"userMetadata,omitempty"` // Additional context data
	Timestamp    time.Time              `json:"timestamp"`    // When the event occurred
}

// AnalyticsService defines the contract for sending analytics events to external systems.
// This interface abstracts the analytics pipeline and enables different implementations
// (Kafka, HTTP endpoints, local files, etc.) to be used.
type AnalyticsService interface {
	// SendEvent transmits an analytics event to the configured analytics system.
	// The implementation should handle event formatting, batching, and error recovery.
	SendEvent(event AnalyticsEvent) error
}