package infra

import (
	"strings"
	"testing"
)

func TestNewBase62KeyGenerationService(t *testing.T) {
	kgs := NewBase62KeyGenerationService()
	if kgs == nil {
		t.Error("expected KGS to be created")
	}
	
	// Test that it implements the interface
	_, ok := kgs.(*Base62KeyGenerationService)
	if !ok {
		t.Error("expected Base62KeyGenerationService implementation")
	}
}

func TestBase62KeyGenerationService_GenerateUniqueID(t *testing.T) {
	kgs := NewBase62KeyGenerationService().(*Base62KeyGenerationService)
	
	// Generate multiple IDs to test uniqueness
	ids := make(map[string]bool)
	duplicateCount := 0
	for i := 0; i < 100; i++ {
		id, err := kgs.GenerateUniqueID()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		
		if len(id) < 7 {
			t.Errorf("expected ID length at least 7, got %d", len(id))
		}
		
		// Check for valid Base62 characters
		if !isValidBase62(id) {
			t.Errorf("ID contains invalid Base62 characters: %q", id)
		}
		
		// Check uniqueness (allow rare duplicates due to randomization)
		if ids[id] {
			duplicateCount++
		}
		ids[id] = true
	}
	
	// Allow a very small number of duplicates due to randomization
	if duplicateCount > 2 {
		t.Errorf("too many duplicate IDs generated: %d", duplicateCount)
	}
}

func TestBase62KeyGenerationService_GetMultipleIDs(t *testing.T) {
	kgs := NewBase62KeyGenerationService().(*Base62KeyGenerationService)
	
	tests := []struct {
		name  string
		count int
	}{
		{
			name:  "single ID",
			count: 1,
		},
		{
			name:  "multiple IDs",
			count: 10,
		},
		{
			name:  "large batch",
			count: 100,
		},
		{
			name:  "zero IDs",
			count: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, err := kgs.GetMultipleIDs(tt.count)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			
			if len(ids) != tt.count {
				t.Errorf("expected %d IDs, got %d", tt.count, len(ids))
			}
			
			// Check uniqueness within the batch
			uniqueIDs := make(map[string]bool)
			for _, id := range ids {
				if len(id) < 7 {
					t.Errorf("expected ID length at least 7, got %d for ID %q", len(id), id)
				}
				
				if !isValidBase62(id) {
					t.Errorf("ID contains invalid Base62 characters: %q", id)
				}
				
				if uniqueIDs[id] {
					t.Errorf("duplicate ID in batch: %q", id)
				}
				uniqueIDs[id] = true
			}
		})
	}
}

func TestBase62KeyGenerationService_NonSequential(t *testing.T) {
	kgs := NewBase62KeyGenerationService().(*Base62KeyGenerationService)
	
	// Generate a sequence of IDs
	ids := make([]string, 20)
	for i := 0; i < 20; i++ {
		id, err := kgs.GenerateUniqueID()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		ids[i] = id
	}
	
	// Check that IDs are not sequential
	sequentialCount := 0
	for i := 1; i < len(ids); i++ {
		// Decode both IDs and check if they are sequential
		val1, err1 := DecodeBase62(strings.TrimLeft(ids[i-1], "0"))
		val2, err2 := DecodeBase62(strings.TrimLeft(ids[i], "0"))
		
		if err1 == nil && err2 == nil && val2 == val1+1 {
			sequentialCount++
		}
	}
	
	// Allow some sequential pairs due to randomization, but not too many
	if sequentialCount > len(ids)/4 {
		t.Errorf("too many sequential IDs detected: %d out of %d pairs", sequentialCount, len(ids)-1)
	}
}

func TestBase62KeyGenerationService_RefillBuffer(t *testing.T) {
	kgs := NewBase62KeyGenerationService().(*Base62KeyGenerationService)
	
	err := kgs.RefillBuffer()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	// Buffer should be filled
	if len(kgs.buffer) == 0 {
		t.Error("expected buffer to be filled")
	}
	
	// Test that buffer is used
	initialBufferSize := len(kgs.buffer)
	id, err := kgs.GenerateUniqueID()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	if len(id) < 7 {
		t.Errorf("expected ID length at least 7, got %d", len(id))
	}
	
	// Buffer should be smaller after generating an ID
	if len(kgs.buffer) != initialBufferSize-1 {
		t.Errorf("expected buffer size %d, got %d", initialBufferSize-1, len(kgs.buffer))
	}
}

func TestBase62KeyGenerationService_generateNonSequentialValue(t *testing.T) {
	kgs := NewBase62KeyGenerationService().(*Base62KeyGenerationService)
	
	// Generate multiple values and check they are different
	values := make(map[int64]bool)
	for i := 0; i < 100; i++ {
		value := kgs.generateNonSequentialValue()
		
		// Value should be positive
		if value < 0 {
			t.Errorf("expected positive value, got %d", value)
		}
		
		// Values should be unique (mostly)
		if values[value] {
			t.Logf("duplicate value generated: %d (this is rare but possible)", value)
		}
		values[value] = true
		
		// Increment counter for next iteration
		kgs.counter++
	}
	
	// Should have generated mostly unique values
	if len(values) < 95 { // Allow some duplicates due to randomization
		t.Errorf("expected at least 95 unique values, got %d", len(values))
	}
}

func TestEncodeBase62(t *testing.T) {
	kgs := &Base62KeyGenerationService{}
	
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{
			name:     "zero",
			input:    0,
			expected: "0",
		},
		{
			name:     "single digit",
			input:    5,
			expected: "5",
		},
		{
			name:     "ten",
			input:    10,
			expected: "a",
		},
		{
			name:     "sixty-one",
			input:    61,
			expected: "Z",
		},
		{
			name:     "sixty-two",
			input:    62,
			expected: "10",
		},
		{
			name:     "large number",
			input:    1000000,
			expected: "4c92",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kgs.encodeBase62(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPadToLength(t *testing.T) {
	kgs := &Base62KeyGenerationService{}
	
	tests := []struct {
		name     string
		input    string
		length   int
		expected string
	}{
		{
			name:     "already correct length",
			input:    "abc123x",
			length:   7,
			expected: "abc123x",
		},
		{
			name:     "pad with zeros",
			input:    "abc",
			length:   7,
			expected: "0000abc",
		},
		{
			name:     "longer than required",
			input:    "abcdefgh",
			length:   7,
			expected: "abcdefgh",
		},
		{
			name:     "empty string",
			input:    "",
			length:   7,
			expected: "0000000",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kgs.padToLength(tt.input, tt.length)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDecodeBase62(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int64
		expectError bool
	}{
		{
			name:     "zero",
			input:    "0",
			expected: 0,
		},
		{
			name:     "single digit",
			input:    "5",
			expected: 5,
		},
		{
			name:     "ten",
			input:    "a",
			expected: 10,
		},
		{
			name:     "sixty-one",
			input:    "Z",
			expected: 61,
		},
		{
			name:     "sixty-two",
			input:    "10",
			expected: 62,
		},
		{
			name:     "large number",
			input:    "4c92",
			expected: 1000000,
		},
		{
			name:        "invalid character",
			input:       "abc@",
			expectError: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeBase62(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestBase62KeyGenerationService_EncodeDecode_RoundTrip(t *testing.T) {
	kgs := &Base62KeyGenerationService{}
	
	testValues := []int64{0, 1, 10, 61, 62, 100, 1000, 1000000, 9223372036854775807} // max int64
	
	for _, val := range testValues {
		encoded := kgs.encodeBase62(val)
		decoded, err := DecodeBase62(encoded)
		
		if err != nil {
			t.Errorf("decode error for value %d: %v", val, err)
			continue
		}
		
		if decoded != val {
			t.Errorf("round trip failed for %d: encoded as %q, decoded as %d", val, encoded, decoded)
		}
	}
}

// Helper function to check if string contains only valid Base62 characters
func isValidBase62(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
	}
	return true
}