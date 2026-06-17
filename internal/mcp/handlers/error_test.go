package handlers

import (
	"errors"
	"os"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name            string
		input           error
		wantCode        string
		wantUserMessage string
	}{
		{
			name:            "os.ErrNotExist maps to not_found",
			input:           os.ErrNotExist,
			wantCode:        "not_found",
			wantUserMessage: "campaign not found",
		},
		{
			name:            "validation error maps to validation_failed",
			input:           errors.New("campaign name is empty"),
			wantCode:        "validation_failed",
			wantUserMessage: "invalid input: campaign name is empty",
		},
		{
			name:            "wrapped os.ErrNotExist maps to not_found",
			input:           errors.New("file not found: some/path"),
			wantCode:        "not_found",
			wantUserMessage: "campaign not found",
		},
		{
			name:            "generic error preserves original message",
			input:           errors.New("database connection failed"),
			wantCode:        "internal_error",
			wantUserMessage: "error: database connection failed",
		},
		{
			name:            "nil error maps to internal_error",
			input:           nil,
			wantCode:        "internal_error",
			wantUserMessage: "an internal error occurred",
		},
		{
			name:            "PDF generation error maps to pdf_generation_failed",
			input:           errors.New("PDF generation failed after 3 attempts: images missing"),
			wantCode:        "pdf_generation_failed",
			wantUserMessage: "error: PDF generation failed after 3 attempts: images missing",
		},
		{
			name:            "unknown error preserves original text",
			input:           errors.New("something completely unexpected happened"),
			wantCode:        "internal_error",
			wantUserMessage: "error: something completely unexpected happened",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			he := classifyError(tt.input)
			if he.Code != tt.wantCode {
				t.Errorf("classifyError() code = %q, want %q", he.Code, tt.wantCode)
			}
			if he.Message != tt.wantUserMessage {
				t.Errorf("classifyError() message = %q, want %q", he.Message, tt.wantUserMessage)
			}
		})
	}
}

func TestNewHandlerError(t *testing.T) {
	cause := errors.New("underlying error")
	he := NewHandlerError("validation_failed", "name is required", cause)

	if he.Code != "validation_failed" {
		t.Errorf("Code = %q, want %q", he.Code, "validation_failed")
	}
	if he.Message != "name is required" {
		t.Errorf("Message = %q, want %q", he.Message, "name is required")
	}
	if he.Cause != cause {
		t.Error("Cause should be the underlying error")
	}
}

func TestToToolResult_WithHandlerError(t *testing.T) {
	he := NewHandlerError("not_found", "campaign not found", os.ErrNotExist)
	result := ToToolResult(he)

	if result == nil {
		t.Fatal("ToToolResult() returned nil")
	}
	if !result.IsError {
		t.Error("ToToolResult() should return error result")
	}
	if len(result.Content) == 0 {
		t.Fatal("ToToolResult() should have content")
	}
	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if textContent.Text != "campaign not found" {
		t.Errorf("message = %q, want %q", textContent.Text, "campaign not found")
	}
}

func TestToToolResult_WithPlainError(t *testing.T) {
	plainErr := errors.New("something went wrong")
	result := ToToolResult(plainErr)

	if result == nil {
		t.Fatal("ToToolResult() returned nil")
	}
	if !result.IsError {
		t.Error("ToToolResult() should return error result for plain errors")
	}
	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	if textContent.Text != "error: something went wrong" {
		t.Errorf("message = %q, want %q", textContent.Text, "error: something went wrong")
	}
}

func TestToToolResult_WithNil(t *testing.T) {
	result := ToToolResult(nil)

	if result == nil {
		t.Fatal("ToToolResult(nil) returned nil")
	}
	if !result.IsError {
		t.Error("ToToolResult(nil) should return error result")
	}
}
