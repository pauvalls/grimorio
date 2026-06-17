package domain

// GeneralFeatures represents shared environmental properties for a location.
// Rendered as a boxed section before individual areas (WotC pattern).
// Uses a single Content field with ***Name.*** inline sub-features for
// ceilings, doors, light, sound, air, walls, etc.
type GeneralFeatures struct {
	Content string `json:"content"` // Markdown content with ***Name.*** inline sub-features
}
