package infra

import (
	"log"

	"github.com/oharai/short-url/internal/shorturl/domain"
)

// MockAnalyticsService is a mock implementation of the AnalyticsService interface.
// It simulates analytics event processing by logging events to the console.
// This implementation is suitable for development, testing, and demonstration purposes.
type MockAnalyticsService struct{}

// NewMockAnalyticsService creates a new instance of the mock analytics service.
// This factory function returns the interface implementation for dependency injection.
//
// Returns:
//   - domain.AnalyticsService: Mock service interface implementation
func NewMockAnalyticsService() domain.AnalyticsService {
	return &MockAnalyticsService{}
}

// SendEvent processes an analytics event by logging it to the console.
// In a production environment, this would send the event to a real analytics system
// such as Kafka, HTTP endpoints, or cloud analytics services.
//
// Parameters:
//   - event: The analytics event to process
//
// Returns:
//   - error: Always nil for this mock implementation
func (a *MockAnalyticsService) SendEvent(event domain.AnalyticsEvent) error {
	// Log the event for debugging and demonstration purposes
	log.Printf("Analytics Event: %+v", event)
	return nil
}