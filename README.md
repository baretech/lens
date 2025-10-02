# Lens - Universal Go Tracing Library

Lens is a Go tracing library that brings the power of observability to your applications with **minimal code changes and maximum flexibility**. Think of it as giving your Go programs X-ray vision - you can see exactly what's happening inside your functions, when they're called, how long they take, and what data flows through them.

## Why Lens?

Most tracing solutions require you to modify your code extensively. You need to add logging statements everywhere, wrap functions manually, or restructure your code to fit the tracing framework. Lens takes a different approach - it uses Go's reflection capabilities to automatically trace your existing code with minimal changes to your business logic, while giving you complete control over what gets traced and how.

The name "Lens" comes from the idea of looking through your code with perfect clarity, seeing the execution flow as it happens. Just like a camera lens brings the world into focus, Lens brings your Go programs into focus.

## How It Works

Lens works by wrapping your functions and objects at runtime using Go's reflection system. When you wrap a function with Lens, it creates a transparent proxy that automatically captures:

- Function calls and their arguments
- Return values and execution time
- Variable changes and state transitions
- Source file locations and line numbers
- Call stack information

The beauty is that your original code remains largely unchanged. You just wrap your functions once, and Lens handles all the tracing automatically while giving you the flexibility to control exactly what gets traced.

## Getting Started

First, install Lens in your Go project:

```bash
go get github.com/baretech/lens
```

Then, in your main function, create a tracer and start wrapping your functions:

```go
package main

import (
    "fmt"
    "github.com/baretech/lens"
)

func main() {
    // Create a tracer with console output
    tracer := lens.New(
        lens.WithLevel(lens.LevelTrace),
        lens.WithWriter(lens.NewConsoleWriter(true)),
    )

    // Wrap your functions - that's it!
    CalculateAge := tracer.Wrap(CalculateAge).(func(int) int)
    IsAdult := tracer.Wrap(IsAdult).(func(int) bool)

    // Use your functions normally - tracing happens automatically
    age := CalculateAge(1990)
    isAdult := IsAdult(age)
    
    fmt.Printf("Age: %d, Adult: %t\n", age, isAdult)
}

func CalculateAge(birthYear int) int {
    return 2024 - birthYear
}

func IsAdult(age int) bool {
    return age >= 18
}
```

## What You'll See

When you run this code, Lens automatically produces detailed tracing output in both console and JSON formats. The console output looks like this:

```
[14:30:15.123] function_call func=main.CalculateAge args=[1990] [main.go:15]
[14:30:15.124] function_return func=main.CalculateAge returns=[34] duration=1.2µs [main.go:15]
[14:30:15.125] function_call func=main.IsAdult args=[34] [main.go:16]
[14:30:15.126] function_return func=main.IsAdult returns=[true] duration=0.8µs [main.go:16]
```

And the JSON output provides structured data perfect for analysis tools. Here are examples of different trace types:

**Function Call:**
```json
{
  "id": "evt_1759436509286785000",
  "trace_id": "trace_1759436509286784000",
  "timestamp": "2025-10-03T01:51:49.286785+05:30",
  "type": "function_call",
  "function": "main.CalculateAge",
  "arguments": [1990],
  "goroutine": 1,
  "source_file": "/path/to/main.go",
  "source_line": 15,
  "source_function": "main.main"
}
```

**Function Return:**
```json
{
  "id": "evt_1759436509286943000",
  "trace_id": "trace_1759436509286941000",
  "timestamp": "2025-10-03T01:51:49.286943+05:30",
  "type": "function_return",
  "function": "main.IsAdult",
  "return_value": [true],
  "duration": 667,
  "goroutine": 7,
  "source_file": "/path/to/main.go",
  "source_line": 16,
  "source_function": "main.main"
}
```

**Variable Change:**
```json
{
  "id": "evt_1759436509287036000",
  "trace_id": "trace_1759436509287037000",
  "timestamp": "2025-10-03T01:51:49.287037+05:30",
  "type": "variable_write",
  "variable": "ProcessUserData.result",
  "old_value": "john doe",
  "new_value": "PROCESSED: JOHN DOE (length: 8)",
  "goroutine": 19,
  "source_file": "/path/to/util.go",
  "source_line": 123,
  "source_function": "main.ProcessUserData"
}
```

Each trace shows you exactly what happened: which function was called, what arguments it received, what it returned, how long it took, and where in your code it happened. The timing information is incredibly precise, measured in microseconds.

## Tracing Objects and Methods

Lens doesn't just work with functions - it can trace entire objects and their methods:

```go
type Calculator struct {
    precision int
}

func (c *Calculator) Add(a, b float64) float64 {
    return a + b
}

func main() {
    tracer := lens.New(lens.WithLevel(lens.LevelTrace))
    
    // Wrap an entire object
    calc := tracer.WrapWithName(NewCalculator(2), "Calculator").(*Calculator)
    
    // All method calls are automatically traced
    result := calc.Add(10.5, 20.3)
    fmt.Printf("Result: %.2f\n", result)
}
```

This will trace every method call on the Calculator object, showing you the complete interaction with your objects.

## Cross-File Tracing

One of Lens's most powerful features is its ability to trace function calls across multiple files. You can have functions in different packages calling each other, and Lens will trace the entire call chain:

```go
// In util.go
func ProcessUserData(name string, tracer lens.Tracer) string {
    // Wrap functions from other files
    ValidateInput := tracer.Wrap(ValidateInput).(func(string) bool)
    ProcessText := tracer.Wrap(ProcessText).(func(string) string)
    
    if !ValidateInput(name) {
        return "Invalid input"
    }
    
    return ProcessText(name)
}
```

Lens will trace the entire flow, showing you how data moves through your application across file boundaries.

## Output Formats

Lens supports multiple output formats. You can write to the console, JSON files, or any custom writer you create:

```go
// Console output with colors
tracer := lens.New(
    lens.WithWriter(lens.NewConsoleWriter(true)),
)

// JSON file output
jsonWriter, _ := lens.NewJSONFileWriter("./traces/app.json")
tracer := lens.New(
    lens.WithWriter(jsonWriter),
)

// Multiple outputs simultaneously
tracer := lens.New(
    lens.WithWriter(lens.NewConsoleWriter(true)),
    lens.WithWriter(jsonWriter),
)
```

The JSON output is particularly useful for analysis tools, allowing you to build custom dashboards and monitoring solutions.

## Variable Tracing

Sometimes you want to trace specific variable changes. Lens provides a simple way to do this:

```go
oldValue := user.Name
user.Name = "New Name"
tracer.TraceVariable("User.Name", oldValue, user.Name)
```

This will log the variable change with the old and new values, helping you track state transitions in your application.

## Performance Considerations

Lens is designed to be lightweight and fast. The reflection overhead is minimal, and you can control the tracing level to balance observability with performance:

```go
// Trace everything (most detailed)
tracer := lens.New(lens.WithLevel(lens.LevelTrace))

// Trace only errors and warnings
tracer := lens.New(lens.WithLevel(lens.LevelError))

// Disable tracing entirely
tracer := lens.New(lens.WithLevel(lens.LevelOff))
```

In production, you might want to use a higher level to reduce overhead while still capturing important information.

## Real-World Use Cases

Lens shines in several scenarios. When you're debugging a complex function that's not behaving as expected, Lens shows you exactly what's happening at each step. When you're optimizing performance, the timing information helps you identify bottlenecks. When you're onboarding new developers, the traces serve as living documentation of how your code actually works.

It's particularly useful in microservices architectures where you need to understand how requests flow through your system. You can trace a request from the API layer down to the database, seeing every function call and data transformation along the way.

## The Future of Lens

We're constantly working to make Lens more powerful and easier to use. One of our most exciting upcoming features is a **web-based UI** that will transform your traces into interactive visualizations. Instead of scrolling through JSON files or console output, you'll be able to see your application's execution flow as an interactive timeline, complete with:

- **Visual call graphs** showing how functions call each other

Beyond the UI, we're also working on **automatic instrumentation** that will make Lens even easier to use. Instead of manually wrapping functions, you'll be able to instrument your entire codebase with a single command:

```bash
lens instrument ./your-project
```

This command will automatically analyze your Go code and add the necessary tracing calls, making Lens truly zero-effort to adopt. The tool will intelligently identify functions that should be traced, add the appropriate wrapper calls, and even suggest optimal tracing configurations based on your code structure.

## Contributing

Lens is an open-source project, and we welcome contributions from the community. Whether you want to add new features, fix bugs, or improve documentation, your help makes Lens better for everyone.

The codebase is designed to be approachable, with clear separation between the core tracing logic and the output formatting. If you're interested in contributing, start by exploring the example code and understanding how the reflection-based wrapping works.

## Getting Help

If you run into issues or have questions, we're here to help. Check out the examples in the repository, and don't hesitate to open an issue if you need assistance. The Go community is known for being helpful and welcoming, and we strive to maintain that tradition.

Lens is more than just a tracing library - it's a tool that helps you understand your code better, debug faster, and build more reliable applications. With minimal code changes and maximum flexibility, Lens adapts to your workflow rather than forcing you to adapt to it. Give it a try, and see how it changes the way you think about observability in Go.
