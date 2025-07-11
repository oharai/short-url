package infra

import (
	"testing"
	"time"

	"github.com/oharai/short-url/internal/shorturl/domain"
)

func TestNewMockAnalyticsService(t *testing.T) {
	analytics := NewMockAnalyticsService()
	if analytics == nil {
		t.Error("expected analytics service to be created")
	}

	// Test that it implements the interface
	_, ok := analytics.(*MockAnalyticsService)
	if !ok {
		t.Error("expected MockAnalyticsService implementation")
	}
}

func TestMockAnalyticsService_SendEvent(t *testing.T) {
	analytics := NewMockAnalyticsService().(*MockAnalyticsService)

	event := domain.AnalyticsEvent{
		EventType: "url_created",
		ShortURL:  "http://test.com/abc123",
		LongURL:   "https://example.com",
		UserMetadata: map[string]interface{}{
			"userId": "user123",
			"source": "api",
		},
		Timestamp: time.Now(),
	}

	err := analytics.SendEvent(event)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMockAnalyticsService_SendEvent_Multiple(t *testing.T) {
	analytics := NewMockAnalyticsService().(*MockAnalyticsService)

	events := []domain.AnalyticsEvent{
		{
			EventType: "url_created",
			ShortURL:  "http://test.com/abc123",
			LongURL:   "https://example1.com",
			Timestamp: time.Now(),
		},
		{
			EventType: "url_accessed",
			ShortURL:  "http://test.com/abc123",
			LongURL:   "https://example1.com",
			UserMetadata: map[string]interface{}{
				"ip": "192.168.1.1",
			},
			Timestamp: time.Now(),
		},
		{
			EventType: "url_deactivated",
			ShortURL:  "http://test.com/abc123",
			LongURL:   "https://example1.com",
			Timestamp: time.Now(),
		},
	}

	for i, event := range events {
		err := analytics.SendEvent(event)
		if err != nil {
			t.Errorf("unexpected error for event %d: %v", i, err)
		}
	}
}

func TestMockAnalyticsService_SendEvent_EmptyEvent(t *testing.T) {
	analytics := NewMockAnalyticsService().(*MockAnalyticsService)

	emptyEvent := domain.AnalyticsEvent{}

	err := analytics.SendEvent(emptyEvent)
	if err != nil {
		t.Errorf("unexpected error for empty event: %v", err)
	}
}

func TestMockAnalyticsService_SendEvent_NilMetadata(t *testing.T) {
	analytics := NewMockAnalyticsService().(*MockAnalyticsService)

	event := domain.AnalyticsEvent{
		EventType:    "url_created",
		ShortURL:     "http://test.com/abc123",
		LongURL:      "https://example.com",
		UserMetadata: nil, // nil metadata
		Timestamp:    time.Now(),
	}

	err := analytics.SendEvent(event)
	if err != nil {
		t.Errorf("unexpected error for event with nil metadata: %v", err)
	}
}
