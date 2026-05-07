package domain

// ConsequenceRule represents a rule that triggers effects based on narrative state
type ConsequenceRule struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Trigger      Trigger     `json:"trigger"`
	Conditions   []Condition `json:"conditions"`
	Effects      []Effect    `json:"effects"`
	Scope        string      `json:"scope"`
	Priority     int         `json:"priority"`
	IsRepeatable bool        `json:"is_repeatable"`
	DMOverride   bool        `json:"dm_override"`
}

// Validate checks if the consequence rule is valid
func (c *ConsequenceRule) Validate() error {
	if c.ID == "" {
		return NewValidationError("id", "rule ID is required")
	}
	if c.Name == "" {
		return NewValidationError("name", "rule name is required")
	}
	if c.Priority <= 0 {
		return NewValidationError("priority", "priority must be greater than 0")
	}
	if len(c.Effects) == 0 {
		return NewValidationError("effects", "at least one effect is required")
	}
	return nil
}

// Trigger represents what activates a consequence rule
type Trigger struct {
	Type     string `json:"type"`
	EntityID string `json:"entity_id"`
	Value    string `json:"value"`
}

// Condition represents a condition that must be met for a rule to fire
type Condition struct {
	Type     string `json:"type"`
	Target   string `json:"target"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

// Effect represents an outcome of a consequence rule
type Effect struct {
	Type        string `json:"type"`
	Target      string `json:"target"`
	Value       any    `json:"value"`
	Delay       string `json:"delay"`
	Description string `json:"description"`
}

// ConsequenceEvaluation holds the result of evaluating consequence rules
type ConsequenceEvaluation struct {
	CampaignID       string            `json:"campaign_id"`
	SessionNum       int               `json:"session_num"`
	TriggeredRules   []ConsequenceRule `json:"triggered_rules"`
	ImmediateEffects []Effect          `json:"immediate_effects"`
	DelayedEffects   []DelayedEffect   `json:"delayed_effects"`
}

// DelayedEffect is an effect queued for a future session
type DelayedEffect struct {
	Effect         Effect `json:"effect"`
	TriggerSession int    `json:"trigger_session"`
	ApplySession   int    `json:"apply_session"`
}

// WorldReactorFunc is a function that handles a trigger and produces effects
type WorldReactorFunc func(trigger Trigger, state *NarrativeState) ([]Effect, error)
