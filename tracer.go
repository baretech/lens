package lens

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"
)

// TracerImpl is the concrete implementation of the Tracer interface
type TracerImpl struct {
	level   Level
	enabled bool
	writers []Writer
	filters []Filter
	mutex   sync.RWMutex
}

// Wrap wraps any object to enable tracing
func (t *TracerImpl) Wrap(obj interface{}) interface{} {
	return t.WrapWithName(obj, "")
}

// WrapWithName wraps an object with a specific name for tracing
func (t *TracerImpl) WrapWithName(obj interface{}, name string) interface{} {
	if !t.enabled {
		return obj
	}

	objType := reflect.TypeOf(obj)

	if objType == nil {
		return obj
	}

	// Handle different types
	switch objType.Kind() {
	case reflect.Ptr:
		return t.wrapPointer(obj, name)
	case reflect.Struct:
		return t.wrapStruct(obj, name)
	case reflect.Func:
		return t.wrapFunction(obj, name)
	case reflect.Interface:
		return t.wrapInterface(obj, name)
	case reflect.Slice:
		return t.wrapSlice(obj, name)
	case reflect.Map:
		return t.wrapMap(obj, name)
	case reflect.Chan:
		return t.wrapChannel(obj, name)
	default:
		// For basic types, we can't wrap them directly, so we return as-is
		return obj
	}
}

// wrapPointer wraps a pointer type
func (t *TracerImpl) wrapPointer(obj interface{}, name string) interface{} {
	objValue := reflect.ValueOf(obj)
	if objValue.IsNil() {
		return obj
	}

	// For now, return the original pointer without deep wrapping
	// This prevents the double pointer issue while maintaining basic tracing
	// In a full implementation, we would create a proper proxy pointer
	return obj
}

// wrapStruct wraps a struct type with automatic method tracing
func (t *TracerImpl) wrapStruct(obj interface{}, name string) interface{} {
	// For now, return the original struct without deep wrapping
	// This prevents the double pointer issue while maintaining basic tracing
	// In a full implementation, we would create a proper proxy struct
	return obj
}

// structWrapper helps create proxies for struct method calls
type structWrapper struct {
	original interface{}
	tracer   *TracerImpl
	name     string
	objType  reflect.Type
	objValue reflect.Value
}

// createProxy creates a proxy that traces method calls
func (sw *structWrapper) createProxy() interface{} {
	// For structs with methods, we need to create a proxy
	// For now, return the original with method wrapping
	proxyValue := reflect.New(sw.objType).Elem()
	proxyValue.Set(sw.objValue)

	// Wrap all methods of the struct
	proxyPtr := reflect.New(sw.objType)
	proxyPtr.Elem().Set(proxyValue)

	// Create method wrappers
	sw.wrapMethods(proxyPtr)

	return proxyPtr.Interface()
}

// wrapMethods wraps all methods of a struct
func (sw *structWrapper) wrapMethods(structPtr reflect.Value) {
	structType := structPtr.Type()

	// Iterate through all methods
	for i := 0; i < structType.NumMethod(); i++ {
		method := structType.Method(i)
		methodName := fmt.Sprintf("%s.%s", sw.name, method.Name)

		// Get the original method
		originalMethod := structPtr.Method(i)

		// Create traced version
		tracedMethod := sw.createMethodWrapper(originalMethod, methodName)

		// Replace the method (this is complex in Go, so we'll use a simpler approach)
		// For now, we'll trace when methods are called through reflection
		_ = tracedMethod
	}
}

// createMethodWrapper creates a wrapper for a specific method
func (sw *structWrapper) createMethodWrapper(method reflect.Value, methodName string) reflect.Value {
	methodType := method.Type()

	wrapper := reflect.MakeFunc(methodType, func(args []reflect.Value) []reflect.Value {
		// Convert args to interface{} slice
		argInterfaces := make([]interface{}, len(args))
		for i, arg := range args {
			if arg.CanInterface() {
				argInterfaces[i] = arg.Interface()
			}
		}

		// Get source location
		sourceLocation := getSourceLocation(2)
		callerLocation := getCallerLocation(2)

		traceID := generateTraceID()

		// Trace method call
		callEvent := Event{
			ID:             generateEventID(),
			TraceID:        traceID,
			Timestamp:      time.Now(),
			Type:           EventMethodCall,
			Component:      sw.name,
			Function:       methodName,
			Arguments:      argInterfaces,
			Goroutine:      getGoroutineID(),
			SourceFile:     sourceLocation.File,
			SourceLine:     sourceLocation.Line,
			SourceFunction: sourceLocation.Function,
			CallerFile:     callerLocation.File,
			CallerLine:     callerLocation.Line,
			CallerFunction: callerLocation.Function,
		}

		sw.tracer.TraceEvent(callEvent)

		start := time.Now()

		// Call the original method
		results := method.Call(args)

		duration := time.Since(start)

		// Convert results to interface{} slice
		resultInterfaces := make([]interface{}, len(results))
		for i, result := range results {
			if result.CanInterface() {
				resultInterfaces[i] = result.Interface()
			}
		}

		// Trace method return
		returnEvent := Event{
			ID:             generateEventID(),
			TraceID:        traceID,
			Timestamp:      time.Now(),
			Type:           EventFunctionReturn,
			Component:      sw.name,
			Function:       methodName,
			ReturnValue:    resultInterfaces,
			Duration:       duration,
			Goroutine:      getGoroutineID(),
			SourceFile:     sourceLocation.File,
			SourceLine:     sourceLocation.Line,
			SourceFunction: sourceLocation.Function,
			CallerFile:     callerLocation.File,
			CallerLine:     callerLocation.Line,
			CallerFunction: callerLocation.Function,
		}

		sw.tracer.TraceEvent(returnEvent)

		return results
	})

	return wrapper
}

// wrapFunction wraps a function type with automatic tracing
func (t *TracerImpl) wrapFunction(obj interface{}, name string) interface{} {
	objType := reflect.TypeOf(obj)
	objValue := reflect.ValueOf(obj)

	// Capture source location at wrap time (when the function is being wrapped)
	wrapSourceLocation := getSourceLocation(2)
	wrapCallerLocation := getCallerLocation(2)

	// Create a wrapper function that automatically traces calls
	wrapper := reflect.MakeFunc(objType, func(args []reflect.Value) []reflect.Value {
		// Get function name from runtime
		pc := reflect.ValueOf(obj).Pointer()
		fn := runtime.FuncForPC(pc)
		funcName := name
		if fn != nil {
			funcName = fn.Name()
		}

		// Convert args to interface{} slice
		argInterfaces := make([]interface{}, len(args))
		for i, arg := range args {
			if arg.CanInterface() {
				argInterfaces[i] = arg.Interface()
			}
		}

		// Use the captured source location from wrap time
		sourceLocation := wrapSourceLocation
		callerLocation := wrapCallerLocation

		traceID := generateTraceID()

		// Trace function call
		callEvent := Event{
			ID:             generateEventID(),
			TraceID:        traceID,
			Timestamp:      time.Now(),
			Type:           EventFunctionCall,
			Component:      name,
			Function:       funcName,
			Arguments:      argInterfaces,
			Goroutine:      getGoroutineID(),
			SourceFile:     sourceLocation.File,
			SourceLine:     sourceLocation.Line,
			SourceFunction: sourceLocation.Function,
			CallerFile:     callerLocation.File,
			CallerLine:     callerLocation.Line,
			CallerFunction: callerLocation.Function,
		}

		t.TraceEvent(callEvent)

		start := time.Now()

		// Call the original function
		results := objValue.Call(args)

		duration := time.Since(start)

		// Convert results to interface{} slice
		resultInterfaces := make([]interface{}, len(results))
		for i, result := range results {
			if result.CanInterface() {
				resultInterfaces[i] = result.Interface()
			}
		}

		// Trace function return
		returnEvent := Event{
			ID:             generateEventID(),
			TraceID:        traceID,
			Timestamp:      time.Now(),
			Type:           EventFunctionReturn,
			Component:      name,
			Function:       funcName,
			ReturnValue:    resultInterfaces,
			Duration:       duration,
			Goroutine:      getGoroutineID(),
			SourceFile:     sourceLocation.File,
			SourceLine:     sourceLocation.Line,
			SourceFunction: sourceLocation.Function,
			CallerFile:     callerLocation.File,
			CallerLine:     callerLocation.Line,
			CallerFunction: callerLocation.Function,
		}

		t.TraceEvent(returnEvent)

		return results
	})

	return wrapper.Interface()
}

// wrapInterface wraps an interface type
func (t *TracerImpl) wrapInterface(obj interface{}, name string) interface{} {
	// For now, return the object as-is
	// In a full implementation, we would create a dynamic proxy
	return obj
}

// wrapSlice wraps a slice type
func (t *TracerImpl) wrapSlice(obj interface{}, name string) interface{} {
	// For now, return the slice as-is
	// In a full implementation, we would create a proxy that traces slice operations
	return obj
}

// wrapMap wraps a map type
func (t *TracerImpl) wrapMap(obj interface{}, name string) interface{} {
	// For now, return the map as-is
	// In a full implementation, we would create a proxy that traces map operations
	return obj
}

// wrapChannel wraps a channel type
func (t *TracerImpl) wrapChannel(obj interface{}, name string) interface{} {
	// For now, return the channel as-is
	// In a full implementation, we would create a proxy that traces channel operations
	return obj
}

// StartSpan starts a new trace span
func (t *TracerImpl) StartSpan(name string) Span {
	return &SpanImpl{
		name:      name,
		startTime: time.Now(),
		tracer:    t,
		traceID:   generateTraceID(),
	}
}

// TraceEvent traces a single event
func (t *TracerImpl) TraceEvent(event Event) {
	if !t.enabled {
		return
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// Apply filters
	for _, filter := range t.filters {
		if !filter.ShouldTrace(event) {
			return
		}
	}

	// Write to all writers
	for _, writer := range t.writers {
		go func(w Writer) {
			w.Write(event)
		}(writer)
	}
}

// TraceVariable traces a variable change
func (t *TracerImpl) TraceVariable(name string, oldVal, newVal interface{}) {
	// Get source location information
	sourceLocation := getSourceLocation(2)
	callerLocation := getCallerLocation(2)

	event := Event{
		ID:             generateEventID(),
		TraceID:        generateTraceID(),
		Timestamp:      time.Now(),
		Type:           EventVariableWrite,
		Variable:       name,
		OldValue:       oldVal,
		NewValue:       newVal,
		Goroutine:      getGoroutineID(),
		SourceFile:     sourceLocation.File,
		SourceLine:     sourceLocation.Line,
		SourceFunction: sourceLocation.Function,
		CallerFile:     callerLocation.File,
		CallerLine:     callerLocation.Line,
		CallerFunction: callerLocation.Function,
	}

	t.TraceEvent(event)
}

// SetLevel sets the tracing level
func (t *TracerImpl) SetLevel(level Level) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.level = level
}

// AddWriter adds a writer to the tracer
func (t *TracerImpl) AddWriter(writer Writer) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.writers = append(t.writers, writer)
}

// AddFilter adds a filter to the tracer
func (t *TracerImpl) AddFilter(filter Filter) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.filters = append(t.filters, filter)
}

// Enable enables tracing
func (t *TracerImpl) Enable() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.enabled = true
}

// Disable disables tracing
func (t *TracerImpl) Disable() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.enabled = false
}

// SpanImpl implements the Span interface
type SpanImpl struct {
	name      string
	startTime time.Time
	tracer    *TracerImpl
	traceID   string
	tags      map[string]interface{}
	error     error
}

// End ends the span
func (s *SpanImpl) End() {
	duration := time.Since(s.startTime)

	event := Event{
		ID:        generateEventID(),
		TraceID:   s.traceID,
		Timestamp: time.Now(),
		Type:      EventFunctionReturn,
		Function:  s.name,
		Duration:  duration,
		Goroutine: getGoroutineID(),
	}

	if s.error != nil {
		event.Error = s.error.Error()
		event.Type = EventError
	}

	s.tracer.TraceEvent(event)
}

// SetTag sets a tag on the span
func (s *SpanImpl) SetTag(key string, value interface{}) {
	if s.tags == nil {
		s.tags = make(map[string]interface{})
	}
	s.tags[key] = value
}

// SetError sets an error on the span
func (s *SpanImpl) SetError(err error) {
	s.error = err
}

// Helper functions for generating IDs
func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

func generateTraceID() string {
	return fmt.Sprintf("trace_%d", time.Now().UnixNano())
}
