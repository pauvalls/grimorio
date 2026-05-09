package validators

import (
	"fmt"
	"github.com/pauvalls/grimorio/internal/domain"
)

// AreaValidator validates area content according to WotC standards.
type AreaValidator struct{}

// NewAreaValidator creates a new AreaValidator.
func NewAreaValidator() *AreaValidator {
	return &AreaValidator{}
}

// ValidateAreaNumber checks if area number is in valid range (1-15).
func (v *AreaValidator) ValidateAreaNumber(num int) error {
	if num < 1 || num > 15 {
		return fmt.Errorf("area number must be between 1 and 15, got %d", num)
	}
	return nil
}

// ValidateLevelRange checks if level range is valid.
func (v *AreaValidator) ValidateLevelRange(min, max int) error {
	if min < 1 || min > 20 {
		return fmt.Errorf("level min must be between 1 and 20, got %d", min)
	}
	if max < 1 || max > 20 {
		return fmt.Errorf("level max must be between 1 and 20, got %d", max)
	}
	if min >= max {
		return fmt.Errorf("level min (%d) must be less than max (%d)", min, max)
	}
	return nil
}

// ValidateCRAppropriate checks if CR is appropriate for level range.
func (v *AreaValidator) ValidateCRAppropriate(cr float64, levelRange domain.LevelRange) error {
	// Simple validation: CR should be within level range ±2
	minCR := float64(levelRange.Min) - 2
	maxCR := float64(levelRange.Max) + 2
	
	if cr < minCR || cr > maxCR {
		return fmt.Errorf("CR %.1f not appropriate for level range %d-%d", cr, levelRange.Min, levelRange.Max)
	}
	return nil
}

// ValidateSequentialNumbers checks if area numbers are sequential.
func (v *AreaValidator) ValidateSequentialNumbers(numbers []int) error {
	if len(numbers) == 0 {
		return nil
	}
	for i := 1; i < len(numbers); i++ {
		if numbers[i] != numbers[i-1]+1 {
			return fmt.Errorf("gap in area numbers: %d followed by %d", numbers[i-1], numbers[i])
		}
	}
	return nil
}

// ValidateMinimumContent checks if area has minimum required content.
func (v *AreaValidator) ValidateMinimumContent(area *domain.Area) error {
	if len(area.Features) == 0 {
		return fmt.Errorf("area must have at least 1 feature")
	}
	if len(area.Encounters) == 0 {
		return fmt.Errorf("area must have at least 1 encounter")
	}
	return nil
}
