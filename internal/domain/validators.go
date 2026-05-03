package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a domain validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) error {
	return ValidationError{Field: field, Message: message}
}

// IsValidKebabCase checks if a string is valid kebab-case
func IsValidKebabCase(s string) bool {
	if s == "" {
		return false
	}
	// Must start and end with alphanumeric
	if s[0] == '-' || s[len(s)-1] == '-' {
		return false
	}
	// Only lowercase letters, numbers, and hyphens
	matched, _ := regexp.MatchString(`^[a-z0-9]+(-[a-z0-9]+)*$`, s)
	return matched
}

// SanitizeFilename sanitizes a string for use as a filename
func SanitizeFilename(s string) string {
	// Replace spaces and special chars with underscores
	result := []rune(s)
	for i, r := range result {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' && r != '_' {
			result[i] = '_'
		}
	}
	return strings.ToLower(string(result))
}

// Common DnD 5e races
var ValidRaces = []string{
	"humano", "elfo", "enano", "mediano", "semielfo", "semiorco",
	"gnomo", "tiefling", "draconido", "trasgo", "firbolg", "gith",
}

// Common DnD 5e classes
var ValidClasses = []string{
	"guerrero", "mago", "clerigo", "picaro", "paladin", "barbaro",
	"ranger", "brujo", "druida", "monje", "bardo", "hechicero",
	"artifice", "sangre",
}

// Common DnD 5e backgrounds
var ValidBackgrounds = []string{
	"soldado", "acolito", "criminal", "sabio", "noble", "artesano",
	"marinero", "ermitano", "charlatan", "heroe del pueblo", "animador",
	"gladiador", "mercenario", "pirata",
}

// Alignments
var ValidAlignments = []string{
	"LG", "NG", "CG", "LN", "N", "CN", "LE", "NE", "CE",
}

// Contains checks if a string slice contains a value (case-insensitive)
func Contains(slice []string, item string) bool {
	item = strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == item {
			return true
		}
	}
	return false
}
