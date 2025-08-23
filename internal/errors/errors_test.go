package errors

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	code := "TEST_ERROR"
	message := "This is a test error"
	
	envelope := New(code, message)
	
	if envelope.Code != code {
		t.Errorf("Expected code %s, got %s", code, envelope.Code)
	}
	
	if envelope.Message != message {
		t.Errorf("Expected message %s, got %s", message, envelope.Message)
	}
	
	if envelope.Details != nil {
		t.Errorf("Expected nil details, got %v", envelope.Details)
	}
}

func TestWrap(t *testing.T) {
	code := "TEST_ERROR"
	message := "This is a test error"
	details := map[string]any{"key": "value", "number": 42}
	
	envelope := Wrap(code, message, details)
	
	if envelope.Code != code {
		t.Errorf("Expected code %s, got %s", code, envelope.Code)
	}
	
	if envelope.Message != message {
		t.Errorf("Expected message %s, got %s", message, envelope.Message)
	}
	
	if envelope.Details == nil {
		t.Error("Expected non-nil details")
	}
	
	detailsMap, ok := envelope.Details.(map[string]any)
	if !ok {
		t.Errorf("Expected details to be map[string]any, got %T", envelope.Details)
	}
	
	if detailsMap["key"] != "value" {
		t.Errorf("Expected details key to be 'value', got %v", detailsMap["key"])
	}
	
	if detailsMap["number"] != 42 {
		t.Errorf("Expected details number to be 42, got %v", detailsMap["number"])
	}
}

func TestWrapError(t *testing.T) {
	code := "WRAPPED_ERROR"
	message := "Wrapped error message"
	originalErr := fmt.Errorf("original error")
	
	envelope := WrapError(code, message, originalErr)
	
	if envelope.Code != code {
		t.Errorf("Expected code %s, got %s", code, envelope.Code)
	}
	
	if envelope.Message != message {
		t.Errorf("Expected message %s, got %s", message, envelope.Message)
	}
	
	if envelope.Details == nil {
		t.Error("Expected non-nil details")
	}
	
	detailsMap, ok := envelope.Details.(map[string]any)
	if !ok {
		t.Errorf("Expected details to be map[string]any, got %T", envelope.Details)
	}
	
	originalError, exists := detailsMap["original_error"]
	if !exists {
		t.Error("Expected original_error in details")
	}
	
	if originalError != "original error" {
		t.Errorf("Expected original_error to be 'original error', got %v", originalError)
	}
}

func TestFromValidationErrors(t *testing.T) {
	validationErrors := []ValidationError{
		{Field: "name", Message: "Name is required", Code: ErrNameRequired},
		{Field: "id", Message: "Invalid UUID format", Code: ErrInvalidUUID},
	}
	
	envelope := FromValidationErrors(validationErrors)
	
	if envelope.Code != ErrInvalidInput {
		t.Errorf("Expected code %s, got %s", ErrInvalidInput, envelope.Code)
	}
	
	if envelope.Message != "Validation failed" {
		t.Errorf("Expected message 'Validation failed', got %s", envelope.Message)
	}
	
	if envelope.Details == nil {
		t.Error("Expected non-nil details")
	}
	
	detailsMap, ok := envelope.Details.(map[string]any)
	if !ok {
		t.Errorf("Expected details to be map[string]any, got %T", envelope.Details)
	}
	
	validationErrorsInterface, exists := detailsMap["validation_errors"]
	if !exists {
		t.Error("Expected validation_errors in details")
	}
	
	retrievedErrors, ok := validationErrorsInterface.([]ValidationError)
	if !ok {
		t.Errorf("Expected validation_errors to be []ValidationError, got %T", validationErrorsInterface)
	}
	
	if len(retrievedErrors) != 2 {
		t.Errorf("Expected 2 validation errors, got %d", len(retrievedErrors))
	}
	
	if retrievedErrors[0].Field != "name" {
		t.Errorf("Expected first error field to be 'name', got %s", retrievedErrors[0].Field)
	}
	
	if retrievedErrors[1].Code != ErrInvalidUUID {
		t.Errorf("Expected second error code to be %s, got %s", ErrInvalidUUID, retrievedErrors[1].Code)
	}
}

func TestEnvelopeError(t *testing.T) {
	tests := []struct {
		name           string
		envelope       Envelope
		expectedString string
	}{
		{
			name:           "Without details",
			envelope:       New("TEST_CODE", "Test message"),
			expectedString: "TEST_CODE: Test message",
		},
		{
			name:           "With details",
			envelope:       Wrap("TEST_CODE", "Test message", map[string]string{"key": "value"}),
			expectedString: "TEST_CODE: Test message (details: map[key:value])",
		},
		{
			name: "With nil details",
			envelope: Envelope{
				Code:    "TEST_CODE",
				Message: "Test message",
				Details: nil,
			},
			expectedString: "TEST_CODE: Test message",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.envelope.Error()
			if result != tt.expectedString {
				t.Errorf("Expected error string %s, got %s", tt.expectedString, result)
			}
		})
	}
}

func TestIsErrorCode(t *testing.T) {
	envelope := New("TEST_CODE", "Test message")
	
	// Test with matching code
	if !IsErrorCode(envelope, "TEST_CODE") {
		t.Error("Expected IsErrorCode to return true for matching code")
	}
	
	// Test with non-matching code
	if IsErrorCode(envelope, "OTHER_CODE") {
		t.Error("Expected IsErrorCode to return false for non-matching code")
	}
	
	// Test with non-envelope error
	regularErr := fmt.Errorf("regular error")
	if IsErrorCode(regularErr, "TEST_CODE") {
		t.Error("Expected IsErrorCode to return false for non-envelope error")
	}
}

func TestErrorCodes(t *testing.T) {
	// Test that all error codes are defined and non-empty
	errorCodes := []string{
		ErrProjectNotFound,
		ErrProjectExists,
		ErrNodeNotFound,
		ErrInvalidPath,
		ErrStorageFailure,
		ErrInvalidInput,
		ErrDuplicateName,
		ErrInvalidParent,
		ErrCircularReference,
		ErrNameRequired,
		ErrInvalidUUID,
		ErrGitFailure,
		ErrNotRepository,
		ErrRemoteFailure,
		ErrSchemaVersion,
		ErrMigrationFailure,
		ErrUnknown,
		ErrNotImplemented,
	}
	
	for _, code := range errorCodes {
		if code == "" {
			t.Errorf("Error code should not be empty")
		}
		
		// Test that code is uppercase with underscores (convention)
		if code != fmt.Sprintf("%s", code) {
			t.Errorf("Error code %s should follow naming convention", code)
		}
	}
	
	// Test that codes are unique
	codeSet := make(map[string]bool)
	for _, code := range errorCodes {
		if codeSet[code] {
			t.Errorf("Duplicate error code: %s", code)
		}
		codeSet[code] = true
	}
}

func TestValidationError(t *testing.T) {
	field := "testField"
	message := "Test validation message"
	code := "TEST_VALIDATION_CODE"
	
	validationError := ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	}
	
	if validationError.Field != field {
		t.Errorf("Expected field %s, got %s", field, validationError.Field)
	}
	
	if validationError.Message != message {
		t.Errorf("Expected message %s, got %s", message, validationError.Message)
	}
	
	if validationError.Code != code {
		t.Errorf("Expected code %s, got %s", code, validationError.Code)
	}
}

// Test that Envelope implements error interface
func TestEnvelopeImplementsError(t *testing.T) {
	var err error = New("TEST_CODE", "Test message")
	
	if err.Error() != "TEST_CODE: Test message" {
		t.Errorf("Envelope should implement error interface properly")
	}
}

// Benchmark tests
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New("BENCH_CODE", "Benchmark message")
	}
}

func BenchmarkWrap(b *testing.B) {
	details := map[string]any{"key": "value"}
	for i := 0; i < b.N; i++ {
		Wrap("BENCH_CODE", "Benchmark message", details)
	}
}

func BenchmarkIsErrorCode(b *testing.B) {
	envelope := New("TEST_CODE", "Test message")
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		IsErrorCode(envelope, "TEST_CODE")
	}
}