// Package main is the entry point for the URL shortening service API server.
// It configures and starts the HTTP server with all necessary dependencies
// using dependency injection principles for clean architecture.
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/oharai/short-url/internal/shorturl/app"
	"github.com/oharai/short-url/internal/shorturl/infra"
	httpHandler "github.com/oharai/short-url/internal/shorturl/interfaces/http"
)

// main initializes and starts the URL shortening service HTTP server.
// It sets up the dependency injection container, configures routing,
// and starts the server on the specified port.
func main() {
	// Configuration - In production, these would come from environment variables
	baseURL := "http://localhost:8080"

	// Dependency Injection Setup
	// Create infrastructure layer implementations
	repo := infra.NewMemoryShortURLRepository()  // Data persistence layer
	kgs := infra.NewBase62KeyGenerationService() // Unique ID generation service
	analytics := infra.NewMockAnalyticsService() // Analytics event processing

	// Create application layer service with injected dependencies
	service := app.NewShortURLService(repo, kgs, analytics, baseURL)

	// Create presentation layer handler
	handler := httpHandler.NewShortURLHandler(service)

	// Route Configuration
	// API endpoints following REST conventions
	http.HandleFunc("/v1/createShortUrl", handler.CreateShortURL)
	http.HandleFunc("/v1/getLongUrl", handler.GetLongURL)
	http.HandleFunc("/admin/shorturls", handler.GetAllShortURLs)
	http.HandleFunc("/admin/deactivate", handler.DeactivateShortURL)

	// Catch-all handler for short URL redirection
	// This handles GET /<shortId> requests and redirects to original URLs
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Handle root path requests
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusNotFound)
			if _, err := w.Write([]byte("Short URL not found")); err != nil {
				log.Printf("Failed to write response: %v", err)
			}
			return
		}

		// Handle unmatched API/admin paths
		if strings.HasPrefix(r.URL.Path, "/v1/") || strings.HasPrefix(r.URL.Path, "/admin/") {
			w.WriteHeader(http.StatusNotFound)
			if _, err := w.Write([]byte("Endpoint not found")); err != nil {
				log.Printf("Failed to write response: %v", err)
			}
			return
		}

		// Delegate to short URL redirection handler
		handler.RedirectShortURL(w, r)
	})

	// Server Configuration and Startup
	port := ":8080"

	// Display startup information for development convenience
	fmt.Printf("Server starting on port %s\n", port)
	fmt.Printf("API Endpoints:\n")
	fmt.Printf("  POST %s/v1/createShortUrl - Create short URL\n", baseURL)
	fmt.Printf("  GET  %s/v1/getLongUrl - Get long URL\n", baseURL)
	fmt.Printf("  GET  %s/admin/shorturls - List all URLs\n", baseURL)
	fmt.Printf("  DELETE %s/admin/deactivate?id=<id> - Deactivate URL\n", baseURL)
	fmt.Printf("  GET  %s/<shortId> - Redirect to long URL\n", baseURL)

	// Configure HTTP server with appropriate timeouts for security
	server := &http.Server{
		Addr:           port,
		Handler:        nil, // Use default ServeMux
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start HTTP server - this call blocks until the server shuts down
	log.Fatal(server.ListenAndServe())
}
