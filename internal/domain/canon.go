package domain

import (
	"fmt"
	"time"
)

// SchemaVersionV2 is the current canon schema version
const SchemaVersionV2 = "2.0"

// EntityType represents the type of a narrative entity
type EntityType string

const (
	EntityTypeNPC      EntityType = "npc"
	EntityTypeLocation EntityType = "location"
	EntityTypeItem     EntityType = "item"
	EntityTypeFaction  EntityType = "faction"
	EntityTypeMonster  EntityType = "monster"
)

// EntityState represents the canonical state of an entity
type EntityState string

const (
	EntityStateAlive       EntityState = "alive"
	EntityStateDead        EntityState = "dead"
	EntityStateMissing     EntityState = "missing"
	EntityStateTransformed EntityState = "transformed"
)

// RelationshipType represents the type of relationship between entities
type RelationshipType string

const (
	RelationshipTypeAlly      RelationshipType = "ally"
	RelationshipTypeEnemy     RelationshipType = "enemy"
	RelationshipTypeRival     RelationshipType = "rival"
	RelationshipTypeIndebted  RelationshipType = "indebted"
	RelationshipTypeBloodOath RelationshipType = "blood_oath"
)

// CanonDocument is the authoritative source of truth for a campaign
type CanonDocument struct {
	SchemaVersion      string                   `json:"schema_version"`
	CampaignID         string                   `json:"campaign_id"`
	CreatedAt          time.Time                `json:"created_at"`
	UpdatedAt          time.Time                `json:"updated_at"`
	Facts              []CanonFact              `json:"facts"`
	Entities           []CanonEntity            `json:"entities"`
	Timeline           []CanonTimelineEvent     `json:"timeline"`
	Rules              []CanonRule              `json:"rules"`
	Relationships      []CanonRelationship      `json:"relationships"`
	ChapterProgression []ChapterProgressionRule `json:"chapter_progression,omitempty"`
	PartyState         *PartyState              `json:"party_state,omitempty"`
}

// Validate checks if the canon document is valid
func (c *CanonDocument) Validate() error {
	if c.SchemaVersion == "" {
		return NewValidationError("schema_version", "schema version is required")
	}
	if c.SchemaVersion != SchemaVersionV2 {
		return NewValidationError("schema_version", fmt.Sprintf("unsupported schema version: %s (expected %s)", c.SchemaVersion, SchemaVersionV2))
	}
	if c.CampaignID == "" {
		return NewValidationError("campaign_id", "campaign ID is required")
	}
	if !IsValidKebabCase(c.CampaignID) {
		return NewValidationError("campaign_id", "campaign ID must be kebab-case")
	}
	return nil
}

// CanonFact represents an immutable or mutable fact about the campaign world
type CanonFact struct {
	ID        string    `json:"id"`
	Category  string    `json:"category"`
	Statement string    `json:"statement"`
	Source    string    `json:"source"`
	Immutable bool      `json:"immutable"`
	CreatedAt time.Time `json:"created_at"`
}

// Validate checks if the fact is valid
func (f *CanonFact) Validate() error {
	if f.ID == "" {
		return NewValidationError("id", "fact ID is required")
	}
	if f.Statement == "" {
		return NewValidationError("statement", "fact statement is required")
	}
	if len(f.Statement) < 10 {
		return NewValidationError("statement", "fact statement must be at least 10 characters")
	}
	if f.Category == "" {
		return NewValidationError("category", "fact category is required")
	}
	return nil
}

// CanonEntity represents a narrative entity (NPC, location, item, faction, monster)
type CanonEntity struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Type        EntityType     `json:"type"`
	Role        string         `json:"role"`
	CanonState  EntityState    `json:"canon_state"`
	Properties  map[string]any `json:"properties"`
	Motivation  string         `json:"motivation,omitempty"`
	Secret      string         `json:"secret,omitempty"`
	Connections []string       `json:"connections"`
}

// Validate checks if the entity is valid
func (e *CanonEntity) Validate() error {
	if e.ID == "" {
		return NewValidationError("id", "entity ID is required")
	}
	if e.Name == "" {
		return NewValidationError("name", "entity name is required")
	}
	switch e.Type {
	case EntityTypeNPC, EntityTypeLocation, EntityTypeItem, EntityTypeFaction, EntityTypeMonster:
		// valid
	default:
		return NewValidationError("type", "invalid entity type: "+string(e.Type))
	}
	switch e.CanonState {
	case EntityStateAlive, EntityStateDead, EntityStateMissing, EntityStateTransformed:
		// valid
	default:
		return NewValidationError("canon_state", "invalid entity state: "+string(e.CanonState))
	}
	return nil
}

// CanonRule represents a rule or constraint in the campaign world
type CanonRule struct {
	ID          string `json:"id"`
	Domain      string `json:"domain"`
	Statement   string `json:"statement"`
	Enforcement string `json:"enforcement"`
}

// Validate checks if the rule is valid
func (r *CanonRule) Validate() error {
	if r.ID == "" {
		return NewValidationError("id", "rule ID is required")
	}
	if r.Statement == "" {
		return NewValidationError("statement", "rule statement is required")
	}
	if r.Domain == "" {
		return NewValidationError("domain", "rule domain is required")
	}
	return nil
}

// CanonRelationship represents a relationship between two entities in the canon
type CanonRelationship struct {
	ID       string           `json:"id"`
	From     string           `json:"from"`
	To       string           `json:"to"`
	Type     RelationshipType `json:"type"`
	Strength int8             `json:"strength"`
	Notes    string           `json:"notes,omitempty"`
}

// Validate checks if the relationship is valid
func (r *CanonRelationship) Validate() error {
	if r.ID == "" {
		return NewValidationError("id", "relationship ID is required")
	}
	if r.From == "" {
		return NewValidationError("from", "relationship 'from' is required")
	}
	if r.To == "" {
		return NewValidationError("to", "relationship 'to' is required")
	}
	switch r.Type {
	case RelationshipTypeAlly, RelationshipTypeEnemy, RelationshipTypeRival, RelationshipTypeIndebted, RelationshipTypeBloodOath:
		// valid
	default:
		return NewValidationError("type", "invalid relationship type: "+string(r.Type))
	}
	if r.Strength < -10 || r.Strength > 10 {
		return NewValidationError("strength", "strength must be between -10 and 10")
	}
	return nil
}

// CanonTimelineEvent represents an event in the campaign timeline
type CanonTimelineEvent struct {
	ID          string   `json:"id"`
	Timestamp   string   `json:"timestamp"`
	Description string   `json:"description"`
	Involved    []string `json:"involved"`
	IsRevealed  bool     `json:"is_revealed"`
}

// Validate checks if the timeline event is valid
func (t *CanonTimelineEvent) Validate() error {
	if t.ID == "" {
		return NewValidationError("id", "timeline event ID is required")
	}
	if t.Description == "" {
		return NewValidationError("description", "timeline event description is required")
	}
	return nil
}

// EntityFilter provides filtering options for entity queries
type EntityFilter struct {
	Type       EntityType
	Role       string
	CanonState EntityState
	NameQuery  string
}

// ContentProposal represents a proposed piece of content for validation
type ContentProposal struct {
	ID               string            `json:"id"`
	Type             string            `json:"type"`
	Content          string            `json:"content"`
	EntityReferences []EntityReference `json:"entity_references"`
	BatchID          string            `json:"batch_id,omitempty"`
	FactionContext   string            `json:"faction_context,omitempty"`
}

// EntityReference references an entity within a proposal
type EntityReference struct {
	EntityID      string      `json:"entity_id"`
	RequiredState EntityState `json:"required_state,omitempty"`
	RoleHint      string      `json:"role_hint,omitempty"`
	Location      string      `json:"location,omitempty"`
}

// ConsistencyScope defines the scope of a consistency check
type ConsistencyScope string

const (
	ConsistencyScopeFull     ConsistencyScope = "full"
	ConsistencyScopeLoreOnly ConsistencyScope = "lore_only"
)

// RelationshipGraph represents the full relationship graph for a campaign
type RelationshipGraph struct {
	CampaignID          string              `json:"campaign_id"`
	Nodes               []CanonEntity       `json:"nodes"`
	Edges               []CanonRelationship `json:"edges"`
	ConnectedComponents [][]string          `json:"connected_components,omitempty"`
}

// CampaignBrief represents the initial brief for campaign generation
type CampaignBrief struct {
	Name             string   `json:"name"`
	BriefDescription string   `json:"brief_description"`
	LevelRange       string   `json:"level_range"`
	Tone             string   `json:"tone"`
	SettingType      string   `json:"setting_type"`
	Themes           []string `json:"themes"`
	VillainType      string   `json:"villain_type"`
	McGuffinType     string   `json:"mcguffin_type"`
}
