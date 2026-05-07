package domain

import "testing"

func TestIsValidTableType(t *testing.T) {
	tests := []struct {
		name string
		t    string
		want bool
	}{
		{"encounter", "encounter", true},
		{"rumor", "rumor", true},
		{"weather", "weather", true},
		{"treasure", "treasure", true},
		{"invalid", "combat", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTableType(tt.t)
			if got != tt.want {
				t.Fatalf("IsValidTableType(%q) = %v, want %v", tt.t, got, tt.want)
			}
		})
	}
}
