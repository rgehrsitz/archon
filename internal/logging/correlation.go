package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/rgehrsitz/archon/internal/errors"
)

// CorrelationIDKey is the context key for correlation IDs
type CorrelationIDKey struct{}

// TraceContext holds debugging and correlation information
type TraceContext struct {
	CorrelationID string            `json:"correlation_id"`
	RequestID     string            `json:"request_id,omitempty"`
	Operation     string            `json:"operation,omitempty"`
	UserID        string            `json:"user_id,omitempty"`
	StartTime     time.Time         `json:"start_time"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// NewTraceContext creates a new trace context with a correlation ID
func NewTraceContext(operation string) *TraceContext {
	return &TraceContext{
		CorrelationID: generateCorrelationID(),
		Operation:     operation,
		StartTime:     time.Now(),
		Metadata:      make(map[string]string),
	}
}

// WithRequestID adds a request ID to the trace context
func (tc *TraceContext) WithRequestID(requestID string) *TraceContext {
	tc.RequestID = requestID
	return tc
}

// WithUserID adds a user ID to the trace context
func (tc *TraceContext) WithUserID(userID string) *TraceContext {
	tc.UserID = userID
	return tc
}

// WithMetadata adds metadata to the trace context
func (tc *TraceContext) WithMetadata(key, value string) *TraceContext {
	if tc.Metadata == nil {
		tc.Metadata = make(map[string]string)
	}
	tc.Metadata[key] = value
	return tc
}

// Duration returns the elapsed time since the trace context was created
func (tc *TraceContext) Duration() time.Duration {
	return time.Since(tc.StartTime)
}

// ToLogFields converts the trace context to structured log fields
func (tc *TraceContext) ToLogFields() map[string]interface{} {
	fields := map[string]interface{}{
		"correlation_id": tc.CorrelationID,
		"operation":      tc.Operation,
		"duration_ms":    tc.Duration().Milliseconds(),
	}
	
	if tc.RequestID != "" {
		fields["request_id"] = tc.RequestID
	}
	
	if tc.UserID != "" {
		fields["user_id"] = tc.UserID
	}
	
	for k, v := range tc.Metadata {
		fields["meta_"+k] = v
	}
	
	return fields
}

// NewContextWithTrace creates a new context with trace information
func NewContextWithTrace(ctx context.Context, trace *TraceContext) context.Context {
	return context.WithValue(ctx, CorrelationIDKey{}, trace)
}

// TraceFromContext retrieves trace information from context
func TraceFromContext(ctx context.Context) *TraceContext {
	if trace, ok := ctx.Value(CorrelationIDKey{}).(*TraceContext); ok {
		return trace
	}
	return nil
}

// ErrorEvent represents a structured error event for correlation
type ErrorEvent struct {
	*TraceContext
	Error     error     `json:"error"`
	ErrorCode string    `json:"error_code,omitempty"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Stack     []string  `json:"stack,omitempty"`
}

// NewErrorEvent creates a new error event with correlation information
func NewErrorEvent(ctx context.Context, err error, message string) *ErrorEvent {
	trace := TraceFromContext(ctx)
	if trace == nil {
		trace = NewTraceContext("unknown")
	}
	
	event := &ErrorEvent{
		TraceContext: trace,
		Error:        err,
		Message:      message,
		Timestamp:    time.Now(),
	}
	
	// Extract error code if it's an envelope error
	if envelope, ok := err.(errors.Envelope); ok {
		event.ErrorCode = envelope.Code
	}
	
	return event
}

// LogError logs a structured error event with correlation
func (l *Logger) LogError(ctx context.Context, err error, message string) {
	event := NewErrorEvent(ctx, err, message)
	
	logEvent := l.Error()
	
	// Add all trace context fields
	for k, v := range event.ToLogFields() {
		logEvent = logEvent.Interface(k, v)
	}
	
	// Add error-specific fields and log
	logEvent.
		Err(err).
		Str("error_code", event.ErrorCode).
		Msg(message)
}

// LogOperation logs the start and completion of an operation
func (l *Logger) LogOperation(ctx context.Context, operation string, fn func() error) error {
	trace := TraceFromContext(ctx)
	if trace == nil {
		trace = NewTraceContext(operation)
		ctx = NewContextWithTrace(ctx, trace)
	}
	
	// Create operation-specific logger
	opLogger := l.WithContext(trace.ToLogFields())
	
	opLogger.Info().
		Str("phase", "start").
		Msg(fmt.Sprintf("Starting operation: %s", operation))
	
	start := time.Now()
	err := fn()
	duration := time.Since(start)
	
	if err != nil {
		opLogger.Error().
			Err(err).
			Str("phase", "error").
			Dur("duration", duration).
			Msg(fmt.Sprintf("Operation failed: %s", operation))
		
		// Log detailed error event
		l.LogError(ctx, err, fmt.Sprintf("Operation %s failed", operation))
	} else {
		opLogger.Info().
			Str("phase", "complete").
			Dur("duration", duration).
			Msg(fmt.Sprintf("Operation completed: %s", operation))
	}
	
	return err
}

// LogStorageOperation logs storage-specific operations with additional context
func (l *Logger) LogStorageOperation(ctx context.Context, operation, nodeID, path string, fn func() error) error {
	trace := TraceFromContext(ctx)
	if trace == nil {
		trace = NewTraceContext(operation)
	}
	
	trace.WithMetadata("node_id", nodeID).WithMetadata("path", path)
	ctx = NewContextWithTrace(ctx, trace)
	
	return l.LogOperation(ctx, operation, fn)
}

// LogIndexOperation logs search index operations with additional context
func (l *Logger) LogIndexOperation(ctx context.Context, operation string, count int, fn func() error) error {
	trace := TraceFromContext(ctx)
	if trace == nil {
		trace = NewTraceContext(operation)
	}
	
	trace.WithMetadata("record_count", fmt.Sprintf("%d", count))
	ctx = NewContextWithTrace(ctx, trace)
	
	return l.LogOperation(ctx, operation, fn)
}

// generateCorrelationID creates a unique correlation ID
func generateCorrelationID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("corr_%d", time.Now().UnixNano())
	}
	return "corr_" + hex.EncodeToString(bytes)
}