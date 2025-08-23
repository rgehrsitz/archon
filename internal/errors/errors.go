package errors

import (
	"fmt"
)

// Common error codes following a structured approach
const (
	// Storage errors
	ErrProjectNotFound    = "PROJECT_NOT_FOUND"
	ErrProjectExists      = "PROJECT_EXISTS" 
	ErrNodeNotFound       = "NODE_NOT_FOUND"
	ErrInvalidPath        = "INVALID_PATH"
	ErrStorageFailure     = "STORAGE_FAILURE"
	
	// Validation errors
	ErrInvalidInput       = "INVALID_INPUT"
	ErrDuplicateName      = "DUPLICATE_NAME" 
	ErrInvalidParent      = "INVALID_PARENT"
	ErrCircularReference  = "CIRCULAR_REFERENCE"
	ErrNameRequired       = "NAME_REQUIRED"
	ErrInvalidUUID        = "INVALID_UUID"
	
	// Git errors
	ErrGitFailure         = "GIT_FAILURE"
	ErrNotRepository      = "NOT_REPOSITORY"
	ErrRemoteFailure      = "REMOTE_FAILURE"
	ErrNotFound           = "NOT_FOUND"
	
	// Schema errors
	ErrSchemaVersion      = "SCHEMA_VERSION_MISMATCH"
	ErrMigrationFailure   = "MIGRATION_FAILURE"
	
	// Search errors
	ErrSearchFailure     = "SEARCH_FAILURE"
	ErrNoProject         = "NO_PROJECT"
	
	// General errors
	ErrUnknown           = "UNKNOWN_ERROR"
	ErrNotImplemented    = "NOT_IMPLEMENTED"
)

// Envelope represents a structured error response
type Envelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Error implements the error interface
func (e Envelope) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("%s: %s (details: %+v)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// New creates a new error envelope
func New(code, message string) Envelope {
	return Envelope{
		Code:    code,
		Message: message,
	}
}

// Wrap creates a new error envelope with details
func Wrap(code, message string, details any) Envelope {
	return Envelope{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// WrapError wraps a standard Go error with an error envelope
func WrapError(code, message string, err error) Envelope {
	return Envelope{
		Code:    code,
		Message: message,
		Details: map[string]any{
			"original_error": err.Error(),
		},
	}
}

// FromValidationErrors creates an error envelope from validation errors
func FromValidationErrors(validationErrors []ValidationError) Envelope {
	return Envelope{
		Code:    ErrInvalidInput,
		Message: "Validation failed",
		Details: map[string]any{
			"validation_errors": validationErrors,
		},
	}
}

// ValidationError represents a single validation failure
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// IsErrorCode checks if an error has a specific error code
func IsErrorCode(err error, code string) bool {
	if envelope, ok := err.(Envelope); ok {
		return envelope.Code == code
	}
	return false
}
