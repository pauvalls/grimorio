package handlers

import (
	"errors"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// HandlerError represents a classified error with a user-safe message.
type HandlerError struct {
	Code    string // not_found, validation_failed, internal_error, already_exists
	Message string // user-safe message
	Cause   error  // internal cause, logged but never sent to client
}

// Error implements the error interface.
func (he *HandlerError) Error() string {
	if he.Cause != nil {
		return he.Message + ": " + he.Cause.Error()
	}
	return he.Message
}

// NewHandlerError creates a new HandlerError.
func NewHandlerError(code, message string, cause error) *HandlerError {
	return &HandlerError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// ToToolResult converts an error to an MCP tool result.
// If the error is a HandlerError, it uses its user-safe message.
// Otherwise, it classifies the error and returns a generic message.
func ToToolResult(err error) *mcp.CallToolResult {
	if err == nil {
		return mcp.NewToolResultError("an internal error occurred")
	}

	var he *HandlerError
	if errors.As(err, &he) {
		return mcp.NewToolResultError(he.Message)
	}

	he = classifyError(err)
	return mcp.NewToolResultError(he.Message)
}

// classifyError categorizes an error into a HandlerError.
func classifyError(err error) *HandlerError {
	if err == nil {
		return NewHandlerError("internal_error", "an internal error occurred", nil)
	}

	errStr := err.Error()

	// Not found errors
	if errors.Is(err, os.ErrNotExist) ||
		strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "not exist") ||
		strings.Contains(errStr, "no such file") {
		return NewHandlerError("not_found", "campaign not found", err)
	}

	// Validation errors
	if strings.Contains(errStr, "required") ||
		strings.Contains(errStr, "empty") ||
		strings.Contains(errStr, "invalid") ||
		strings.Contains(errStr, "must be") {
		return NewHandlerError("validation_failed", "invalid input: "+errStr, err)
	}

	// Already exists
	if strings.Contains(errStr, "already exists") ||
		strings.Contains(errStr, "duplicate") {
		return NewHandlerError("already_exists", "resource already exists", err)
	}

	// Default: internal error
	return NewHandlerError("internal_error", "an internal error occurred", err)
}
