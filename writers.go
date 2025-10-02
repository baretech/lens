package lens

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// JSONFileWriter writes trace events to JSON files
type JSONFileWriter struct {
	path   string
	file   *os.File
	mutex  sync.Mutex
	buffer []Event
}

// NewJSONFileWriter creates a new JSON file writer
func NewJSONFileWriter(path string) (*JSONFileWriter, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return &JSONFileWriter{
		path:   path,
		file:   file,
		buffer: make([]Event, 0),
	}, nil
}

// Write writes an event to the JSON file
func (w *JSONFileWriter) Write(event Event) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	_, err = w.file.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// Flush flushes any buffered data
func (w *JSONFileWriter) Flush() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}

// Close closes the writer
func (w *JSONFileWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.file != nil {
		err := w.file.Close()
		w.file = nil
		return err
	}
	return nil
}

// ConsoleWriter writes trace events to the console
type ConsoleWriter struct {
	colored bool
	mutex   sync.Mutex
}

// NewConsoleWriter creates a new console writer
func NewConsoleWriter(colored bool) *ConsoleWriter {
	return &ConsoleWriter{
		colored: colored,
	}
}

// Write writes an event to the console
func (w *ConsoleWriter) Write(event Event) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	var output string
	if w.colored {
		output = w.formatColored(event)
	} else {
		output = w.formatPlain(event)
	}

	fmt.Println(output)
	return nil
}

// formatColored formats an event with colors
func (w *ConsoleWriter) formatColored(event Event) string {
	const (
		colorReset  = "\033[0m"
		colorRed    = "\033[31m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorBlue   = "\033[34m"
		colorPurple = "\033[35m"
		colorCyan   = "\033[36m"
	)

	var color string
	switch event.Type {
	case EventError, EventPanic:
		color = colorRed
	case EventFunctionCall, EventMethodCall:
		color = colorBlue
	case EventFunctionReturn:
		color = colorGreen
	case EventVariableRead, EventVariableWrite:
		color = colorYellow
	default:
		color = colorCyan
	}

	timestamp := event.Timestamp.Format("15:04:05.000")
	return fmt.Sprintf("%s[%s] %s %s%s", color, timestamp, event.Type, w.formatEventDetails(event), colorReset)
}

// formatPlain formats an event without colors
func (w *ConsoleWriter) formatPlain(event Event) string {
	timestamp := event.Timestamp.Format("15:04:05.000")
	return fmt.Sprintf("[%s] %s %s", timestamp, event.Type, w.formatEventDetails(event))
}

// formatEventDetails formats the details of an event
func (w *ConsoleWriter) formatEventDetails(event Event) string {
	var details string

	switch event.Type {
	case EventFunctionCall, EventMethodCall:
		if event.Function != "" {
			details = fmt.Sprintf("func=%s args=%v", event.Function, event.Arguments)
		}
	case EventFunctionReturn:
		if event.Function != "" {
			duration := ""
			if event.Duration > 0 {
				duration = fmt.Sprintf(" duration=%v", event.Duration)
			}
			details = fmt.Sprintf("func=%s returns=%v%s", event.Function, event.ReturnValue, duration)
		}
	case EventVariableRead, EventVariableWrite:
		if event.Variable != "" {
			if event.Type == EventVariableWrite {
				details = fmt.Sprintf("var=%s old=%v new=%v", event.Variable, event.OldValue, event.NewValue)
			} else {
				details = fmt.Sprintf("var=%s value=%v", event.Variable, event.NewValue)
			}
		}
	case EventError:
		details = fmt.Sprintf("error=%s", event.Error)
	default:
		if event.Component != "" {
			details = fmt.Sprintf("component=%s", event.Component)
		}
	}

	// Add source location information
	sourceInfo := ""
	if event.CallerFile != "" && event.CallerLine > 0 {
		// Extract just the filename from the full path
		filename := event.CallerFile
		if lastSlash := len(filename) - 1; lastSlash >= 0 {
			for i := lastSlash; i >= 0; i-- {
				if filename[i] == '/' {
					filename = filename[i+1:]
					break
				}
			}
		}
		sourceInfo = fmt.Sprintf(" [%s:%d]", filename, event.CallerLine)
	}

	return details + sourceInfo
}

// Flush flushes any buffered data (no-op for console)
func (w *ConsoleWriter) Flush() error {
	return nil
}

// Close closes the writer (no-op for console)
func (w *ConsoleWriter) Close() error {
	return nil
}
