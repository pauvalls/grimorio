package domain

import "testing"

func TestIsValidExportFormat(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"pdf", true},
		{"markdown", true},
		{"epub", true},
		{"html", false},
		{"", false},
		{"PDF", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsValidExportFormat(tt.input); got != tt.want {
				t.Errorf("IsValidExportFormat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
