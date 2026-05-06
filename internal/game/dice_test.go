package game

import (
	"testing"
)

func TestParseDice(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *DiceSpec
		wantErr bool
	}{
		{
			name:    "simple d20",
			input:   "d20",
			want:    &DiceSpec{Count: 1, Sides: 20, Modifier: 0},
			wantErr: false,
		},
		{
			name:    "2d6",
			input:   "2d6",
			want:    &DiceSpec{Count: 2, Sides: 6, Modifier: 0},
			wantErr: false,
		},
		{
			name:    "1d8+3",
			input:   "1d8+3",
			want:    &DiceSpec{Count: 1, Sides: 8, Modifier: 3},
			wantErr: false,
		},
		{
			name:    "3d10-2",
			input:   "3d10-2",
			want:    &DiceSpec{Count: 3, Sides: 10, Modifier: -2},
			wantErr: false,
		},
		{
			name:    "4d6+0",
			input:   "4d6+0",
			want:    &DiceSpec{Count: 4, Sides: 6, Modifier: 0},
			wantErr: false,
		},
		{
			name:    "d100",
			input:   "d100",
			want:    &DiceSpec{Count: 1, Sides: 100, Modifier: 0},
			wantErr: false,
		},
		{
			name:    "2d8-5",
			input:   "2d8-5",
			want:    &DiceSpec{Count: 2, Sides: 8, Modifier: -5},
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - no d",
			input:   "20",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format - multiple d",
			input:   "2d6d8",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "zero sides",
			input:   "2d0",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "negative sides",
			input:   "2d-6",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "zero count",
			input:   "0d6",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "negative count",
			input:   "-2d6",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid modifier format",
			input:   "2d6+",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "whitespace",
			input:   "  2d6+3  ",
			want:    &DiceSpec{Count: 2, Sides: 6, Modifier: 3},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDice(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDice(%q) expected error but got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseDice(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got.Count != tt.want.Count {
				t.Errorf("ParseDice(%q) count = %d, want %d", tt.input, got.Count, tt.want.Count)
			}
			if got.Sides != tt.want.Sides {
				t.Errorf("ParseDice(%q) sides = %d, want %d", tt.input, got.Sides, tt.want.Sides)
			}
			if got.Modifier != tt.want.Modifier {
				t.Errorf("ParseDice(%q) modifier = %d, want %d", tt.input, got.Modifier, tt.want.Modifier)
			}
		})
	}
}

func TestRoll(t *testing.T) {
	tests := []struct {
		name        string
		spec        *DiceSpec
		minTotal    int
		maxTotal    int
		minResults  int
		maxResults  int
	}{
		{
			name:       "1d20",
			spec:       &DiceSpec{Count: 1, Sides: 20, Modifier: 0},
			minTotal:   1,
			maxTotal:   20,
			minResults: 1,
			maxResults: 1,
		},
		{
			name:       "2d6+3",
			spec:       &DiceSpec{Count: 2, Sides: 6, Modifier: 3},
			minTotal:   5,  // 1+1+3
			maxTotal:   15, // 6+6+3
			minResults: 2,
			maxResults: 2,
		},
		{
			name:       "3d8-2",
			spec:       &DiceSpec{Count: 3, Sides: 8, Modifier: -2},
			minTotal:   1,  // 1+1+1-2
			maxTotal:   22, // 8+8+8-2
			minResults: 3,
			maxResults: 3,
		},
		{
			name:       "d100",
			spec:       &DiceSpec{Count: 1, Sides: 100, Modifier: 0},
			minTotal:   1,
			maxTotal:   100,
			minResults: 1,
			maxResults: 1,
		},
		{
			name:       "4d6 (drop lowest simulation)",
			spec:       &DiceSpec{Count: 4, Sides: 6, Modifier: 0},
			minTotal:   4,
			maxTotal:   24,
			minResults: 4,
			maxResults: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Roll(tt.spec)

			if result.Total < tt.minTotal || result.Total > tt.maxTotal {
				t.Errorf("Roll() total = %d, want between %d and %d", result.Total, tt.minTotal, tt.maxTotal)
			}

			if len(result.Results) < tt.minResults || len(result.Results) > tt.maxResults {
				t.Errorf("Roll() results count = %d, want between %d and %d", len(result.Results), tt.minResults, tt.maxResults)
			}

			// Verify each die result is within bounds
			for i, r := range result.Results {
				if r < 1 || r > tt.spec.Sides {
					t.Errorf("Roll() result[%d] = %d, want between 1 and %d", i, r, tt.spec.Sides)
				}
			}

			// Verify total equals sum of results + modifier
			sum := 0
			for _, r := range result.Results {
				sum += r
			}
			expectedTotal := sum + tt.spec.Modifier
			if result.Total != expectedTotal {
				t.Errorf("Roll() total = %d, expected %d (sum=%d + modifier=%d)", result.Total, expectedTotal, sum, tt.spec.Modifier)
			}

			if result.Modifier != tt.spec.Modifier {
				t.Errorf("Roll() modifier = %d, want %d", result.Modifier, tt.spec.Modifier)
			}

			if result.Dice != tt.spec.String() {
				t.Errorf("Roll() dice = %s, want %s", result.Dice, tt.spec.String())
			}
		})
	}
}

func TestRoll_ZeroCount(t *testing.T) {
	spec := &DiceSpec{Count: 0, Sides: 6, Modifier: 0}
	result := Roll(spec)
	if result.Total != 0 {
		t.Errorf("Roll() with count=0 should return 0, got %d", result.Total)
	}
	if len(result.Results) != 0 {
		t.Errorf("Roll() with count=0 should return empty results, got %d", len(result.Results))
	}
}

func TestRoll_MultipleRollsDistribution(t *testing.T) {
	// Roll 1d6 1000 times and verify distribution
	spec := &DiceSpec{Count: 1, Sides: 6, Modifier: 0}
	counts := make(map[int]int)
	
	for i := 0; i < 1000; i++ {
		result := Roll(spec)
		counts[result.Total]++
	}

	// All results should be between 1 and 6
	for i := 1; i <= 6; i++ {
		if counts[i] == 0 {
			t.Errorf("Roll distribution: value %d never appeared in 1000 rolls", i)
		}
	}

	// Roughly uniform distribution (each should be ~167)
	for i := 1; i <= 6; i++ {
		if counts[i] < 50 || counts[i] > 300 {
			t.Errorf("Roll distribution: value %d appeared %d times (expected ~167)", i, counts[i])
		}
	}
}

func TestDiceSpec_String(t *testing.T) {
	tests := []struct {
		spec *DiceSpec
		want string
	}{
		{&DiceSpec{Count: 1, Sides: 20, Modifier: 0}, "d20"},
		{&DiceSpec{Count: 2, Sides: 6, Modifier: 0}, "2d6"},
		{&DiceSpec{Count: 1, Sides: 8, Modifier: 3}, "1d8+3"},
		{&DiceSpec{Count: 3, Sides: 10, Modifier: -2}, "3d10-2"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.spec.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseAndRollIntegration(t *testing.T) {
	// Test that parsed dice can be rolled
	notations := []string{"d20", "2d6", "1d8+3", "3d10-2", "d100"}
	
	for _, notation := range notations {
		t.Run(notation, func(t *testing.T) {
			spec, err := ParseDice(notation)
			if err != nil {
				t.Fatalf("ParseDice(%q) error: %v", notation, err)
			}
			
			result := Roll(spec)
			if result.Total == 0 && spec.Count > 0 {
				t.Errorf("Roll(%q) returned 0, expected non-zero", notation)
			}
		})
	}
}
