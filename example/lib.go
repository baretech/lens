package main

import (
	"fmt"
	"strings"
)

// ProcessText processes text with various operations
func ProcessText(text string) string {
	// Convert to uppercase
	result := strings.ToUpper(text)

	// Add prefix
	result = fmt.Sprintf("PROCESSED: %s", result)

	// Add suffix with length
	result = fmt.Sprintf("%s (length: %d)", result, len(text))

	return result
}

// ValidateInput validates input text
func ValidateInput(text string) bool {
	if len(text) == 0 {
		return false
	}

	// Check for minimum length
	if len(text) < 3 {
		return false
	}

	// Check for valid characters (no special chars)
	for _, char := range text {
		if char < 32 || char > 126 {
			return false
		}
	}

	return true
}
