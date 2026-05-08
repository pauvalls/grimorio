package validators

import (
	"encoding/json"
	"regexp"
	"strings"
)

// QuestJSON represents the expected quest JSON structure
type QuestJSON struct {
	Description string   `json:"description"`
	Objectives  []string `json:"objectives"`
	Reward      struct {
		Type        string `json:"type"`
		Description string `json:"description"`
		Value       string `json:"value"`
	} `json:"reward"`
}

// ValidateQuestCompleteness checks that quest JSON has all required fields
func ValidateQuestCompleteness(jsonStr string) ValidationResult {
	result := ValidationResult{Valid: true}
	
	// Check 1: Valid JSON
	var quest QuestJSON
	if err := json.Unmarshal([]byte(jsonStr), &quest); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "json_format",
			Message: "invalid JSON: " + err.Error(),
		})
		return result
	}
	
	// Check 2: Description present and substantial
	if quest.Description == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "description",
			Message: "quest description is required",
		})
	} else if len(strings.TrimSpace(quest.Description)) < 20 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "description_length",
			Message: "quest description must be at least 20 characters, got " + itoa(len(quest.Description)),
		})
	}
	
	// Check 3: Objectives present (2-4 items)
	if len(quest.Objectives) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "objectives",
			Message: "quest must have at least 1 objective",
		})
	} else if len(quest.Objectives) < 2 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "objectives_count",
			Message: "quest should have 2-4 objectives for WotC standard, found " + itoa(len(quest.Objectives)),
		})
	} else if len(quest.Objectives) > 4 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "objectives_count",
			Message: "quest has too many objectives (" + itoa(len(quest.Objectives)) + "), maximum recommended is 4",
		})
	}
	
	// Check 4: Each objective is substantial
	for i, obj := range quest.Objectives {
		if len(strings.TrimSpace(obj)) < 5 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "objective_" + itoa(i+1),
				Message: "objective " + itoa(i+1) + " is too vague (" + obj + ")",
			})
		}
	}
	
	// Check 5: Reward type present
	if quest.Reward.Type == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "reward_type",
			Message: "reward type is required (gold, item, xp, reputation, magic_item)",
		})
	}
	
	// Check 6: Reward description present
	if quest.Reward.Description == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "reward_description",
			Message: "reward description is required",
		})
	}
	
	// Check 7: Reward value present
	if quest.Reward.Value == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "reward_value",
			Message: "reward value is required (e.g., '500 gp', '+2 reputation', '1200 XP')",
		})
	}
	
	// Check 8: XP reward present (WotC standard)
	xpPattern := regexp.MustCompile(`(?i)(xp|experience)`)
	hasXP := xpPattern.MatchString(jsonStr)
	if !hasXP {
		result.Warnings = append(result.Warnings, "quest does not specify XP reward - WotC standards recommend including XP")
	}
	
	result.Valid = len(result.Errors) == 0
	return result
}

// ValidateQuestObjectives checks that objectives are actionable and measurable
func ValidateQuestObjectives(objectives []string) ValidationResult {
	result := ValidationResult{Valid: true}
	
	actionVerbs := regexp.MustCompile(`(?i)(recuper|derrot|investig|habl|encuentr|explor|derrot|rescat|proteg|destru|cre|llev|consig|obten|derrot|ven|supera)`)
	
	for i, obj := range objectives {
		// Check each objective has an action verb
		if !actionVerbs.MatchString(obj) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "objective_" + itoa(i+1) + "_action",
				Message: "objective " + itoa(i+1) + " should start with an action verb (recuperar, derrotar, investigar, etc.)",
			})
		}
		
		// Check objective is specific (not vague)
		vaguePatterns := []string{"algo", "cosas", "varios", "algunos", "maybe", "perhaps"}
		for _, vague := range vaguePatterns {
			if strings.Contains(strings.ToLower(obj), vague) {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "objective_" + itoa(i+1) + "_specificity",
					Message: "objective " + itoa(i+1) + " is too vague - be specific about what/where/how",
				})
			}
		}
	}
	
	result.Valid = len(result.Errors) == 0
	return result
}

// ValidateRewardStructure checks that rewards follow WotC standards
func ValidateRewardStructure(rewardJSON string) ValidationResult {
	result := ValidationResult{Valid: true}
	
	var reward struct {
		Type        string `json:"type"`
		Description string `json:"description"`
		Value       string `json:"value"`
	}
	
	if err := json.Unmarshal([]byte(rewardJSON), &reward); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "reward_json",
			Message: "invalid reward JSON: " + err.Error(),
		})
		return result
	}
	
	// Check reward type is valid
	validTypes := map[string]bool{
		"gold": true, "xp": true, "item": true, "magic_item": true,
		"reputation": true, "information": true, "ally": true, "service": true,
	}
	
	if reward.Type != "" && !validTypes[strings.ToLower(reward.Type)] {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "reward_type",
			Message: "invalid reward type '" + reward.Type + "' - must be one of: gold, xp, item, magic_item, reputation, information, ally, service",
		})
	}
	
	// Check reward value format
	if reward.Value != "" {
		// Should contain a number for gold/xp
		if strings.ToLower(reward.Type) == "gold" || strings.ToLower(reward.Type) == "xp" {
			numberPattern := regexp.MustCompile(`\d+`)
			if !numberPattern.MatchString(reward.Value) {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "reward_value_format",
					Message: "gold/xp rewards should specify a numeric value (e.g., '500 gp', '1200 XP')",
				})
			}
		}
	}
	
	result.Valid = len(result.Errors) == 0
	return result
}
