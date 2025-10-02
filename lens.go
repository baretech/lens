// Package lens provides universal, zero-code-change tracing for any Go application
package lens

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// Tracer is the main interface for the lens tracing system
type Tracer interface {
	// Generic wrappers - work with any type T
	Wrap(obj interface{}) interface{}
	WrapWithName(obj interface{}, name string) interface{}

	// Manual tracing (for advanced use cases)
	StartSpan(name string) Span
	TraceEvent(event Event)
	TraceVariable(name string, oldVal, newVal interface{})

	// Configuration
	SetLevel(level Level)
	AddWriter(writer Writer)
	AddFilter(filter Filter)
	Enable()
	Disable()
}

// Event represents a single trace event
type Event struct {
	ID          string        `json:"id"`
	TraceID     string        `json:"trace_id"`
	Timestamp   time.Time     `json:"timestamp"`
	Type        EventType     `json:"type"`
	Component   string        `json:"component"`
	Function    string        `json:"function,omitempty"`
	Variable    string        `json:"variable,omitempty"`
	OldValue    interface{}   `json:"old_value,omitempty"`
	NewValue    interface{}   `json:"new_value,omitempty"`
	Arguments   []interface{} `json:"arguments,omitempty"`
	ReturnValue []interface{} `json:"return_value,omitempty"`
	Error       string        `json:"error,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
	StackTrace  []string      `json:"stack_trace,omitempty"`
	Goroutine   int           `json:"goroutine"`
	// Enhanced source location information
	SourceFile     string `json:"source_file,omitempty"`
	SourceLine     int    `json:"source_line,omitempty"`
	SourceFunction string `json:"source_function,omitempty"`
	CallerFile     string `json:"caller_file,omitempty"`
	CallerLine     int    `json:"caller_line,omitempty"`
	CallerFunction string `json:"caller_function,omitempty"`
}

// EventType defines the type of trace event
type EventType string

const (
	EventVariableRead     EventType = "variable_read"
	EventVariableWrite    EventType = "variable_write"
	EventFunctionCall     EventType = "function_call"
	EventFunctionReturn   EventType = "function_return"
	EventMethodCall       EventType = "method_call"
	EventFieldAccess      EventType = "field_access"
	EventSliceOperation   EventType = "slice_operation"
	EventMapOperation     EventType = "map_operation"
	EventChannelOperation EventType = "channel_operation"
	EventError            EventType = "error"
	EventPanic            EventType = "panic"
)

// Level defines the tracing level
type Level int

const (
	LevelOff Level = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
	LevelTrace
)

// Span represents a trace span
type Span interface {
	End()
	SetTag(key string, value interface{})
	SetError(err error)
}

// Writer interface for outputting trace events
type Writer interface {
	Write(event Event) error
	Flush() error
	Close() error
}

// Filter interface for filtering trace events
type Filter interface {
	ShouldTrace(event Event) bool
}

// New creates a new tracer instance
func New(options ...Option) *TracerImpl {
	tracer := &TracerImpl{
		level:   LevelInfo,
		enabled: true,
		writers: make([]Writer, 0),
		filters: make([]Filter, 0),
	}

	for _, option := range options {
		option(tracer)
	}

	return tracer
}

// Option is a function that configures a tracer
type Option func(*TracerImpl)

// WithLevel sets the tracing level
func WithLevel(level Level) Option {
	return func(t *TracerImpl) {
		t.level = level
	}
}

// WithWriter adds a writer to the tracer
func WithWriter(writer Writer) Option {
	return func(t *TracerImpl) {
		t.writers = append(t.writers, writer)
	}
}

// SourceLocation represents source code location information
type SourceLocation struct {
	File     string
	Line     int
	Function string
}

// getSourceLocation returns detailed source location information
func getSourceLocation(skip int) SourceLocation {
	// Try multiple skip levels to find the actual source location
	for i := skip; i < skip+10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		// Skip runtime and internal lens files
		if !isInternalFile(file) {
			fn := runtime.FuncForPC(pc)
			var funcName string
			if fn != nil {
				funcName = fn.Name()
			}

			return SourceLocation{
				File:     file,
				Line:     line,
				Function: funcName,
			}
		}
	}

	// Fallback to the original skip level if no good location found
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return SourceLocation{}
	}

	fn := runtime.FuncForPC(pc)
	var funcName string
	if fn != nil {
		funcName = fn.Name()
	}

	return SourceLocation{
		File:     file,
		Line:     line,
		Function: funcName,
	}
}

// isInternalFile checks if a file is an internal runtime or lens file
func isInternalFile(file string) bool {
	// Skip Go runtime files
	if strings.Contains(file, "/usr/local/go/src/") || strings.Contains(file, "/usr/lib/go/src/") {
		return true
	}
	// Skip lens internal files (if any)
	if strings.Contains(file, "lens.go") || strings.Contains(file, "tracer.go") {
		return true
	}
	return false
}

// getCallerLocation returns the caller's source location
func getCallerLocation(skip int) SourceLocation {
	return getSourceLocation(skip + 1)
}

// getStackTrace returns the current stack trace
func getStackTrace(skip int) []string {
	var traces []string
	for i := skip; i < skip+10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			traces = append(traces, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
		}
	}
	return traces
}

// getGoroutineID returns the current goroutine ID
func getGoroutineID() int {
	// Simple implementation - in production, this would be more sophisticated
	return int(runtime.NumGoroutine())
}
