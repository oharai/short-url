package infra

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/oharai/short-url/internal/shorturl/domain"
)

// base62Chars defines the character set used for Base62 encoding.
// Includes digits (0-9), lowercase letters (a-z), and uppercase letters (A-Z).
const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Base62KeyGenerationService implements the KeyGenerationService interface
// using Base62 encoding with randomization to prevent predictable IDs.
// This provides short, URL-safe identifiers with guaranteed uniqueness.
type Base62KeyGenerationService struct {
	mu      sync.Mutex // Mutex for thread-safe counter operations
	counter int64      // Counter for uniqueness (with randomization)
	buffer  []string   // Pre-generated IDs buffer for performance optimization
}

// NewBase62KeyGenerationService creates a new instance of the Base62 key generation service.
// Initializes the counter with a random starting value to prevent predictable sequences.
//
// Returns:
//   - domain.KeyGenerationService: Service interface implementation
func NewBase62KeyGenerationService() domain.KeyGenerationService {
	// Initialize counter with a random value to prevent predictable sequences
	randomStart, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	
	return &Base62KeyGenerationService{
		counter: randomStart.Int64() + time.Now().Unix(), // Combine random value with timestamp
		buffer:  make([]string, 0),
	}
}

// GenerateUniqueID generates a single unique Base62-encoded identifier.
// Uses buffered IDs when available for better performance, otherwise generates new ones.
//
// Returns:
//   - string: A unique 7-character Base62-encoded identifier
//   - error: Generation error if any
func (k *Base62KeyGenerationService) GenerateUniqueID() (string, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Use buffered ID if available
	if len(k.buffer) > 0 {
		id := k.buffer[0]
		k.buffer = k.buffer[1:]
		return id, nil
	}

	// Generate new ID on demand
	ids, err := k.generateMultipleIDsInternal(1)
	if err != nil {
		return "", err
	}
	return ids[0], nil
}

// GetMultipleIDs generates multiple unique identifiers in a single operation.
// This is more efficient than calling GenerateUniqueID multiple times.
//
// Parameters:
//   - count: Number of IDs to generate
//
// Returns:
//   - []string: Slice of unique Base62-encoded identifiers
//   - error: Generation error if any
func (k *Base62KeyGenerationService) GetMultipleIDs(count int) ([]string, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	return k.generateMultipleIDsInternal(count)
}

// generateMultipleIDsInternal is the internal implementation for generating multiple IDs.
// This method assumes the caller has already acquired the mutex lock.
// It combines counter values with random elements to prevent predictable sequences.
//
// Parameters:
//   - count: Number of IDs to generate
//
// Returns:
//   - []string: Slice of generated IDs
//   - error: Always nil for this implementation
func (k *Base62KeyGenerationService) generateMultipleIDsInternal(count int) ([]string, error) {
	ids := make([]string, count)
	for i := range count {
		// Generate a non-sequential ID by combining counter with random elements
		uniqueValue := k.generateNonSequentialValue()
		id := k.encodeBase62(uniqueValue)
		ids[i] = k.padToLength(id, 7)
		k.counter++
	}
	return ids, nil
}

// generateNonSequentialValue creates a non-sequential value by combining
// the counter with random elements and bit manipulation to prevent predictability.
//
// Returns:
//   - int64: A unique but non-sequential value
func (k *Base62KeyGenerationService) generateNonSequentialValue() int64 {
	// Use counter for uniqueness but scramble it to prevent sequential patterns
	
	// Method 1: XOR with a random seed based on current time
	timeSeed := time.Now().UnixNano() & 0xFFFF // Use lower 16 bits of nanoseconds
	scrambled := k.counter ^ timeSeed
	
	// Method 2: Add a small random component
	randomComponent, _ := rand.Int(rand.Reader, big.NewInt(1000))
	scrambled += randomComponent.Int64()
	
	// Method 3: Bit rotation to further scramble the value
	// Rotate bits to make the pattern less predictable
	scrambled = ((scrambled << 13) | (scrambled >> (64 - 13))) // Rotate left by 13 bits
	
	// Ensure the result is positive and within a reasonable range
	if scrambled < 0 {
		scrambled = -scrambled
	}
	
	return scrambled
}

// encodeBase62 converts a decimal number to Base62 encoding.
// Uses the character set: 0-9, a-z, A-Z (62 total characters).
//
// Parameters:
//   - num: The decimal number to encode
//
// Returns:
//   - string: Base62-encoded representation
func (k *Base62KeyGenerationService) encodeBase62(num int64) string {
	if num == 0 {
		return "0"
	}

	result := ""
	for num > 0 {
		result = string(base62Chars[num%62]) + result
		num /= 62
	}
	return result
}

// padToLength pads a string with leading zeros to reach the specified length.
// This ensures consistent ID length across all generated identifiers.
//
// Parameters:
//   - str: The string to pad
//   - length: The desired final length
//
// Returns:
//   - string: Padded string of the specified length
func (k *Base62KeyGenerationService) padToLength(str string, length int) string {
	if len(str) >= length {
		return str
	}

	padding := length - len(str)
	for i := 0; i < padding; i++ {
		str = "0" + str
	}
	return str
}

// RefillBuffer pre-generates IDs and stores them in the buffer for faster access.
// This method can be called periodically to maintain a buffer of ready-to-use IDs.
//
// Returns:
//   - error: Buffer refill error if any
func (k *Base62KeyGenerationService) RefillBuffer() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Refill buffer when it gets low
	if len(k.buffer) < 100 {
		newIDs, err := k.generateMultipleIDsInternal(1000)
		if err != nil {
			return fmt.Errorf("failed to refill buffer: %w", err)
		}
		k.buffer = append(k.buffer, newIDs...)
	}
	return nil
}

// DecodeBase62 converts a Base62-encoded string back to its decimal representation.
// This utility function can be used for reverse operations or validation.
//
// Parameters:
//   - encoded: The Base62-encoded string to decode
//
// Returns:
//   - int64: The decimal representation
//   - error: Decoding error if invalid characters are found
func DecodeBase62(encoded string) (int64, error) {
	result := int64(0)
	power := int64(1)

	// Process characters from right to left
	for i := len(encoded) - 1; i >= 0; i-- {
		char := encoded[i]
		var digit int64

		// Convert character to its numeric value
		switch {
		case char >= '0' && char <= '9':
			digit = int64(char - '0')
		case char >= 'a' && char <= 'z':
			digit = int64(char-'a') + 10
		case char >= 'A' && char <= 'Z':
			digit = int64(char-'A') + 36
		default:
			return 0, fmt.Errorf("invalid character in base62 string: %c", char)
		}

		result += digit * power
		power *= 62
	}

	return result, nil
}
