package main

import (
	"fmt"
	"math"
	"time"

	"github.com/baretech/lens"
)

// Calculator provides mathematical operations
type Calculator struct {
	precision int
}

// NewCalculator creates a new calculator with specified precision
func NewCalculator(precision int) *Calculator {
	return &Calculator{precision: precision}
}

// Add performs addition with precision
func (c *Calculator) Add(a, b float64) float64 {
	result := a + b
	return c.round(result)
}

// Multiply performs multiplication with precision
func (c *Calculator) Multiply(a, b float64) float64 {
	result := a * b
	return c.round(result)
}

// CalculateTax calculates tax based on amount and rate
func (c *Calculator) CalculateTax(amount, rate float64) float64 {
	tax := amount * rate
	return c.round(tax)
}

// round rounds a number to the specified precision
func (c *Calculator) round(value float64) float64 {
	factor := math.Pow(10, float64(c.precision))
	return math.Round(value*factor) / factor
}

// StringProcessor handles string operations
type StringProcessor struct {
	operations int
}

// NewStringProcessor creates a new string processor
func NewStringProcessor() *StringProcessor {
	return &StringProcessor{operations: 0}
}

// ProcessString processes a string with various operations
func (sp *StringProcessor) ProcessString(input string) string {
	sp.operations++

	// Simulate some processing time
	time.Sleep(1 * time.Millisecond)

	// Convert to uppercase
	result := fmt.Sprintf("PROCESSED: %s", input)

	fmt.Printf("Processed string: %s (operation #%d)\n", result, sp.operations)
	return result
}

// GetOperationCount returns the number of operations performed
func (sp *StringProcessor) GetOperationCount() int {
	return sp.operations
}

// ValidateEmail validates an email address (simple validation)
func ValidateEmail(email string) bool {
	if len(email) == 0 {
		return false
	}

	// Simple validation - check for @ symbol
	hasAt := false
	for _, char := range email {
		if char == '@' {
			hasAt = true
			break
		}
	}

	return hasAt
}

// FormatCurrency formats a number as currency
func FormatCurrency(amount float64, currency string) string {
	return fmt.Sprintf("%s %.2f", currency, amount)
}

// CalculateAge calculates age from birth year
func CalculateAge(birthYear int) int {
	currentYear := time.Now().Year()
	return currentYear - birthYear
}

// IsAdult checks if a person is an adult based on age
func IsAdult(age int) bool {
	return age >= 18
}

// ProcessUserData processes user data using lib functions
func ProcessUserData(name string, tracer lens.Tracer) string {
	// Wrap internal functions for tracing using the passed tracer
	ValidateInput := tracer.Wrap(ValidateInput).(func(string) bool)
	ProcessText := tracer.Wrap(ProcessText).(func(string) string)

	// Trace the validation step
	valid := ValidateInput(name)
	if !valid {
		tracer.TraceVariable("ProcessUserData.validation", false, false)
		return "Invalid input"
	}

	// Trace the processing step
	result := ProcessText(name)
	tracer.TraceVariable("ProcessUserData.result", name, result)

	return result
}
