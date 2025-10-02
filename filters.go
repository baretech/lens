package lens

import (
	"path/filepath"
	"time"
)

// PackageFilter filters events based on package patterns
type PackageFilter struct {
	includePatterns []string
	excludePatterns []string
}

// NewPackageFilter creates a new package filter
func NewPackageFilter() *PackageFilter {
	return &PackageFilter{
		includePatterns: make([]string, 0),
		excludePatterns: make([]string, 0),
	}
}

// IncludePackages adds include patterns
func (f *PackageFilter) IncludePackages(patterns ...string) *PackageFilter {
	f.includePatterns = append(f.includePatterns, patterns...)
	return f
}

// ExcludePackages adds exclude patterns
func (f *PackageFilter) ExcludePackages(patterns ...string) *PackageFilter {
	f.excludePatterns = append(f.excludePatterns, patterns...)
	return f
}

// ShouldTrace determines if an event should be traced
func (f *PackageFilter) ShouldTrace(event Event) bool {
	component := event.Component
	if component == "" {
		return true
	}

	// Check exclude patterns first
	for _, pattern := range f.excludePatterns {
		if matched, _ := filepath.Match(pattern, component); matched {
			return false
		}
	}

	// If no include patterns, allow all (that weren't excluded)
	if len(f.includePatterns) == 0 {
		return true
	}

	// Check include patterns
	for _, pattern := range f.includePatterns {
		if matched, _ := filepath.Match(pattern, component); matched {
			return true
		}
	}

	return false
}

// FunctionFilter filters events based on function patterns
type FunctionFilter struct {
	includePatterns []string
	excludePatterns []string
}

// NewFunctionFilter creates a new function filter
func NewFunctionFilter() *FunctionFilter {
	return &FunctionFilter{
		includePatterns: make([]string, 0),
		excludePatterns: make([]string, 0),
	}
}

// IncludeFunctions adds include patterns
func (f *FunctionFilter) IncludeFunctions(patterns ...string) *FunctionFilter {
	f.includePatterns = append(f.includePatterns, patterns...)
	return f
}

// ExcludeFunctions adds exclude patterns
func (f *FunctionFilter) ExcludeFunctions(patterns ...string) *FunctionFilter {
	f.excludePatterns = append(f.excludePatterns, patterns...)
	return f
}

// ShouldTrace determines if an event should be traced
func (f *FunctionFilter) ShouldTrace(event Event) bool {
	function := event.Function
	if function == "" {
		return true
	}

	// Check exclude patterns first
	for _, pattern := range f.excludePatterns {
		if matched, _ := filepath.Match(pattern, function); matched {
			return false
		}
	}

	// If no include patterns, allow all (that weren't excluded)
	if len(f.includePatterns) == 0 {
		return true
	}

	// Check include patterns
	for _, pattern := range f.includePatterns {
		if matched, _ := filepath.Match(pattern, function); matched {
			return true
		}
	}

	return false
}

// DurationFilter filters events based on minimum duration
type DurationFilter struct {
	minDuration time.Duration
}

// NewDurationFilter creates a new duration filter
func NewDurationFilter(minDuration time.Duration) *DurationFilter {
	return &DurationFilter{
		minDuration: minDuration,
	}
}

// ShouldTrace determines if an event should be traced
func (f *DurationFilter) ShouldTrace(event Event) bool {
	// Only filter events that have duration
	if event.Duration == 0 {
		return true
	}

	return event.Duration >= f.minDuration
}

// EventTypeFilter filters events based on event types
type EventTypeFilter struct {
	allowedTypes map[EventType]bool
}

// NewEventTypeFilter creates a new event type filter
func NewEventTypeFilter(types ...EventType) *EventTypeFilter {
	allowedTypes := make(map[EventType]bool)
	for _, t := range types {
		allowedTypes[t] = true
	}

	return &EventTypeFilter{
		allowedTypes: allowedTypes,
	}
}

// ShouldTrace determines if an event should be traced
func (f *EventTypeFilter) ShouldTrace(event Event) bool {
	if len(f.allowedTypes) == 0 {
		return true
	}

	return f.allowedTypes[event.Type]
}

// CompositeFilter combines multiple filters with AND logic
type CompositeFilter struct {
	filters []Filter
}

// NewCompositeFilter creates a new composite filter
func NewCompositeFilter(filters ...Filter) *CompositeFilter {
	return &CompositeFilter{
		filters: filters,
	}
}

// AddFilter adds a filter to the composite
func (f *CompositeFilter) AddFilter(filter Filter) *CompositeFilter {
	f.filters = append(f.filters, filter)
	return f
}

// ShouldTrace determines if an event should be traced (all filters must pass)
func (f *CompositeFilter) ShouldTrace(event Event) bool {
	for _, filter := range f.filters {
		if !filter.ShouldTrace(event) {
			return false
		}
	}
	return true
}

// Convenience functions for creating common filters

// IncludePackages creates a filter that includes only specified packages
func IncludePackages(patterns ...string) Filter {
	return NewPackageFilter().IncludePackages(patterns...)
}

// ExcludePackages creates a filter that excludes specified packages
func ExcludePackages(patterns ...string) Filter {
	return NewPackageFilter().ExcludePackages(patterns...)
}

// IncludeFunctions creates a filter that includes only specified functions
func IncludeFunctions(patterns ...string) Filter {
	return NewFunctionFilter().IncludeFunctions(patterns...)
}

// ExcludeFunctions creates a filter that excludes specified functions
func ExcludeFunctions(patterns ...string) Filter {
	return NewFunctionFilter().ExcludeFunctions(patterns...)
}

// MinDuration creates a filter that only traces events with minimum duration
func MinDuration(duration time.Duration) Filter {
	return NewDurationFilter(duration)
}

// OnlyEventTypes creates a filter that only traces specified event types
func OnlyEventTypes(types ...EventType) Filter {
	return NewEventTypeFilter(types...)
}

// ExcludeCommonNoise creates a filter that excludes common noisy functions
func ExcludeCommonNoise() Filter {
	return ExcludeFunctions(
		"*.String",
		"*.GoString",
		"*.Error",
		"runtime.*",
		"reflect.*",
		"fmt.*",
	)
}
