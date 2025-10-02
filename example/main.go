package main

import (
	"fmt"
	"time"

	"github.com/baretech/lens"
)

// User represents a simple user
type User struct {
	ID   int
	Name string
	Age  int
}

// Save simulates saving a user
func (u *User) Save() error {
	fmt.Printf("Saving user: %s (ID: %d)\n", u.Name, u.ID)
	time.Sleep(2 * time.Millisecond)
	return nil
}

// UpdateAge updates the user's age
func (u *User) UpdateAge(newAge int) {
	oldAge := u.Age
	u.Age = newAge
	fmt.Printf("Updated age: %d -> %d\n", oldAge, newAge)
}

// CalculateDiscount calculates discount based on age
func CalculateDiscount(age int) float64 {
	if age >= 65 {
		return 0.2 // 20% senior discount
	}
	if age >= 18 {
		return 0.1 // 10% adult discount
	}
	return 0.0 // No discount for minors
}

func main() {
	fmt.Println("=== Lens Tracing Demo ===")
	fmt.Println("Zero-code-change tracing for any Go application!")

	// Create tracer with both console and JSON output
	jsonWriter, _ := lens.NewJSONFileWriter("./traces/simple.json")
	tracer := lens.New(
		lens.WithLevel(lens.LevelTrace),
		lens.WithWriter(lens.NewConsoleWriter(true)),
		lens.WithWriter(jsonWriter),
	)

	fmt.Println("\n🎯 Step 1: Trace utility functions!")

	// Wrap utility functions - automatic tracing
	CalculateAge := tracer.Wrap(CalculateAge).(func(int) int)
	IsAdult := tracer.Wrap(IsAdult).(func(int) bool)
	ValidateEmail := tracer.Wrap(ValidateEmail).(func(string) bool)
	FormatCurrency := tracer.Wrap(FormatCurrency).(func(float64, string) string)

	// Use utility functions - all calls are automatically traced
	age := CalculateAge(1990)
	fmt.Printf("Age for birth year 1990: %d\n", age)

	isAdult := IsAdult(age)
	fmt.Printf("Is adult: %t\n", isAdult)

	emailValid := ValidateEmail("user@example.com")
	fmt.Printf("Email valid: %t\n", emailValid)

	currency := FormatCurrency(123.456, "USD")
	fmt.Printf("Formatted currency: %s\n", currency)

	// Test cross-file function calls
	processedData := ProcessUserData("john doe", tracer)
	fmt.Printf("Processed user data: %s\n", processedData)

	fmt.Println("\n🎯 Step 2: Trace struct methods!")

	// Create and wrap calculator - just wrap it!
	calc := tracer.WrapWithName(NewCalculator(2), "Calculator").(*Calculator)

	// Use calculator methods - automatic tracing
	sum := calc.Add(10.5, 20.3)
	fmt.Printf("Addition result: %.2f\n", sum)

	product := calc.Multiply(5.5, 3.2)
	fmt.Printf("Multiplication result: %.2f\n", product)

	tax := calc.CalculateTax(100.0, 0.08)
	fmt.Printf("Tax calculation: %.2f\n", tax)

	fmt.Println("\n🎯 Step 3: Trace string processor!")

	// Create and wrap string processor - just wrap it!
	processor := tracer.WrapWithName(NewStringProcessor(), "StringProcessor").(*StringProcessor)

	// Use string processor methods - automatic tracing
	processed1 := processor.ProcessString("hello world")
	processed2 := processor.ProcessString("golang tracing")

	fmt.Printf("Processed strings: %s, %s\n", processed1, processed2)
	fmt.Printf("Total operations: %d\n", processor.GetOperationCount())

	fmt.Println("\n🎯 Step 4: Trace user operations!")

	// Create and wrap user - just wrap it!
	user := tracer.WrapWithName(&User{ID: 1, Name: "Alice", Age: 30}, "User").(*User)

	// Use user methods - automatic tracing
	user.UpdateAge(31)
	user.Save()

	// Manual variable tracing
	oldName := user.Name
	user.Name = "Alice Smith"
	tracer.TraceVariable("User.Name", oldName, user.Name)
}
