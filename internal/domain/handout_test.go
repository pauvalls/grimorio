package domain

import "testing"

func TestIsValidHandoutType(t *testing.T) {
	tests := []struct {
		name string
		t    string
		want bool
	}{
		{"summary", "summary", true},
		{"encounter", "encounter", true},
		{"quest", "quest", true},
		{"lore", "lore", true},
		{"faction", "faction", true},
		{"invalid", "item", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidHandoutType(tt.t)
			if got != tt.want {
				t.Fatalf("IsValidHandoutType(%q) = %v, want %v", tt.t, got, tt.want)
			}
		})
	}
}
