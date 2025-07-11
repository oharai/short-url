package app

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/oharai/short-url/internal/shorturl/domain"
)

// ShortURLService is the primary application service that orchestrates
// the URL shortening business use cases. It coordinates between domain entities,
// repositories, and external services to implement the application's core functionality.
type ShortURLService struct {
	repo      domain.ShortURLRepository   // Repository for persisting short URL entities
	kgs       domain.KeyGenerationService // Service for generating unique identifiers
	analytics domain.AnalyticsService     // Service for sending analytics events
	baseURL   string                      // Base URL for constructing complete short URLs
}

// NewShortURLService creates a new instance of the ShortURLService with all required dependencies.
// This constructor follows the dependency injection pattern to ensure testability and flexibility.
//
// Parameters:
//   - repo: Repository implementation for data persistence
//   - kgs: Key generation service for creating unique identifiers
//   - analytics: Analytics service for event tracking
//   - baseURL: Base URL used for constructing complete short URLs
//
// Returns:
//   - *ShortURLService: Configured service instance ready for use
func NewShortURLService(repo domain.ShortURLRepository, kgs domain.KeyGenerationService, analytics domain.AnalyticsService, baseURL string) *ShortURLService {
	return &ShortURLService{
		repo:      repo,
		kgs:       kgs,
		analytics: analytics,
		baseURL:   baseURL,
	}
}

// CreateShortURL implements the URL shortening use case.
// It supports both automatic ID generation and custom URL specification.
// The method validates input, generates or validates the identifier, creates the domain entity,
// persists it, and sends analytics events.
//
// Parameters:
//   - req: Request containing the long URL and optional parameters
//
// Returns:
//   - *CreateShortURLResponse: Contains the generated short URL
//   - error: Validation or system error
func (s *ShortURLService) CreateShortURL(req CreateShortURLRequest) (*CreateShortURLResponse, error) {
	// Validate required input
	if req.LongURL == "" {
		return nil, errors.New("longUrl is required")
	}

	var shortURL *domain.ShortURL
	var err error

	// Handle custom URL path vs. automatic generation
	if req.CustomURL != "" {
		// Check if custom URL is already taken
		existing, _ := s.repo.FindByID(req.CustomURL)
		if existing != nil {
			return nil, errors.New("custom URL already exists")
		}

		// Create entity with custom identifier
		shortURL, err = domain.NewCustomShortURL(req.CustomURL, req.LongURL, req.Expiry, req.UserMetadata)
		if err != nil {
			return nil, err
		}
	} else {
		// Generate unique identifier using KGS
		id, err := s.kgs.GenerateUniqueID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate unique ID: %w", err)
		}

		// Build complete short URL and create entity
		shortURLStr := s.buildShortURL(id)
		shortURL, err = domain.NewShortURL(id, req.LongURL, shortURLStr, req.Expiry, req.UserMetadata)
		if err != nil {
			return nil, err
		}
	}

	// Persist the entity
	if err := s.repo.Save(shortURL); err != nil {
		return nil, err
	}

	// Send analytics event for tracking
	event := domain.AnalyticsEvent{
		EventType:    "url_created",
		ShortURL:     shortURL.ShortURL(),
		LongURL:      shortURL.LongURL(),
		UserMetadata: shortURL.UserMetadata(),
		Timestamp:    time.Now(),
	}
	s.analytics.SendEvent(event)

	return &CreateShortURLResponse{
		ShortURL: shortURL.ShortURL(),
	}, nil
}

// GetLongURL implements the URL resolution use case for redirection.
// It extracts the identifier from the short URL, validates the entity's status,
// tracks the access event, and returns the original URL for redirection.
//
// Parameters:
//   - req: Request containing the short URL to resolve and context metadata
//
// Returns:
//   - string: The original long URL for redirection
//   - error: Error if URL not found, expired, or inactive
func (s *ShortURLService) GetLongURL(req GetLongURLRequest) (string, error) {
	// Validate required input
	if req.ShortURL == "" {
		return "", errors.New("shortUrl is required")
	}

	// Extract the identifier from the complete short URL
	id := s.extractIDFromShortURL(req.ShortURL)
	shortURL, err := s.repo.FindByID(id)
	if err != nil {
		return "", err
	}

	// Check if URL exists
	if shortURL == nil {
		return "", errors.New("short URL not found")
	}

	// Validate URL status and expiration
	if !shortURL.IsActive() || shortURL.IsExpired() {
		return "", errors.New("short URL is not active or expired")
	}

	// Track access event for analytics
	event := domain.AnalyticsEvent{
		EventType:    "url_accessed",
		ShortURL:     shortURL.ShortURL(),
		LongURL:      shortURL.LongURL(),
		UserMetadata: req.UserMetadata,
		Timestamp:    time.Now(),
	}
	s.analytics.SendEvent(event)

	return shortURL.LongURL(), nil
}

// GetAllShortURLs retrieves all short URLs for administrative purposes.
// This method is typically used by admin interfaces to display URL statistics
// and management information.
//
// Returns:
//   - []*ShortURLResponse: List of all short URLs with complete information
//   - error: Repository error if data access fails
func (s *ShortURLService) GetAllShortURLs() ([]*ShortURLResponse, error) {
	// Retrieve all entities from repository
	shortURLs, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	// Convert domain entities to response DTOs
	var responses []*ShortURLResponse
	for _, shortURL := range shortURLs {
		responses = append(responses, &ShortURLResponse{
			ID:           shortURL.ID(),
			LongURL:      shortURL.LongURL(),
			ShortURL:     shortURL.ShortURL(),
			CreatedAt:    shortURL.CreatedAt(),
			Expiry:       shortURL.Expiry(),
			IsActive:     shortURL.IsActive(),
			UserMetadata: shortURL.UserMetadata(),
		})
	}

	return responses, nil
}

// DeactivateShortURL implements the URL deactivation use case.
// This prevents the URL from being used for redirection while preserving
// the record for analytics and audit purposes.
//
// Parameters:
//   - id: The unique identifier of the URL to deactivate
//
// Returns:
//   - error: Error if URL not found or persistence fails
func (s *ShortURLService) DeactivateShortURL(id string) error {
	// Validate required input
	if id == "" {
		return errors.New("ID is required")
	}

	// Retrieve the entity
	shortURL, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if shortURL == nil {
		return errors.New("short URL not found")
	}

	// Apply business operation
	shortURL.Deactivate()
	
	// Persist the change
	err = s.repo.Save(shortURL)
	if err != nil {
		return err
	}

	// Track deactivation event
	event := domain.AnalyticsEvent{
		EventType:    "url_deactivated",
		ShortURL:     shortURL.ShortURL(),
		LongURL:      shortURL.LongURL(),
		UserMetadata: shortURL.UserMetadata(),
		Timestamp:    time.Now(),
	}
	s.analytics.SendEvent(event)

	return nil
}

// buildShortURL constructs the complete short URL by combining the base URL with the identifier.
// This ensures consistent URL format across the application.
func (s *ShortURLService) buildShortURL(id string) string {
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.baseURL, "/"), id)
}

// extractIDFromShortURL extracts the unique identifier from a complete short URL.
// This is used to resolve short URLs back to their stored identifiers.
func (s *ShortURLService) extractIDFromShortURL(shortURL string) string {
	parts := strings.Split(shortURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return shortURL
}