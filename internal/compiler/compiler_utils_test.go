package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "hello-world"},
		{"Test-123", "test-123"},
		{"Multiple   Spaces", "multiple---spaces"},
		{"UPPERCASE", "uppercase"},
		{"special!@#chars", "special---chars"},
	}

	for _, tt := range tests {
		result := sanitizeID(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeID(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestIsTableSeparator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"|---|---|", true},
		{"| --- | --- |", true},
		{"|:-:|:---:|", true},
		{"not a separator", false},
		{"| cell | cell |", false},
	}

	for _, tt := range tests {
		result := isTableSeparator(tt.input)
		if result != tt.expected {
			t.Errorf("isTableSeparator(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestParseTableRow(t *testing.T) {
	result := parseTableRow("| cell1 | cell2 | cell3 |")
	expected := []string{"cell1", "cell2", "cell3"}
	if len(result) != len(expected) {
		t.Fatalf("parseTableRow() len = %d, want %d", len(result), len(expected))
	}
	for i, v := range expected {
		if strings.TrimSpace(result[i]) != v {
			t.Errorf("parseTableRow()[%d] = %q, want %q", i, result[i], v)
		}
	}
}

func TestParseTableAlign(t *testing.T) {
	result := parseTableAlign("| --- |:---:| ---:|")
	expected := []string{"left", "center", "right"}
	if len(result) != len(expected) {
		t.Fatalf("parseTableAlign() len = %d, want %d", len(result), len(expected))
	}
	for i, v := range expected {
		if strings.TrimSpace(result[i]) != v {
			t.Errorf("parseTableAlign()[%d] = %q, want %q", i, result[i], v)
		}
	}
}

func TestFindCoverImage(t *testing.T) {
	tmpDir := t.TempDir()
	assetsDir := filepath.Join(tmpDir, "assets")

	// No cover image
	result := findCoverImage(tmpDir)
	if result != "" {
		t.Errorf("findCoverImage() = %q, want empty string", result)
	}

	// Create a cover image in assets dir
	os.MkdirAll(assetsDir, 0755)
	coverPath := filepath.Join(assetsDir, "cover.png")
	os.WriteFile(coverPath, []byte("fake image"), 0644)

	result = findCoverImage(tmpDir)
	if result != coverPath {
		t.Errorf("findCoverImage() = %q, want %q", result, coverPath)
	}
}
