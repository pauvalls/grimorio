package validators

import (
	"strings"
	"testing"
)

func TestValidateDevelopments_EnglishDefaults(t *testing.T) {
	input := `# Act 1

### Development

**If the PCs enter the cave:**
- **Immediate consequence:** The goblins raise an alarm.
- **Future consequence:** Reinforcements arrive in area 3.
- **Recovery:** The PCs can barricade the entrance to buy time.

**If the PCs sneak around:**
- **Immediate consequence:** They avoid the guards.
- **Future consequence:** They miss a clue about the boss.
- **Recovery:** A captured scout can be interrogated.

**If the PCs attack:**
- **Immediate consequence:** Combat begins.
- **Future consequence:** The boss is warned.
- **Recovery:** Surrendering goblins offer to negotiate.
`

	result := ValidateDevelopments(input)
	if !result.Valid {
		t.Errorf("expected valid English developments, got errors: %v", result.Errors)
	}
}

func TestValidateDevelopments_MissingSection(t *testing.T) {
	input := `# Act 1

No development section here.
`
	result := ValidateDevelopments(input)
	if result.Valid {
		t.Error("expected failure when Development section is missing")
	}
	found := false
	for _, err := range result.Errors {
		if err.Field == "developments" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing developments error, got: %v", result.Errors)
	}
}

func TestValidateDevelopments_TooFewBranches(t *testing.T) {
	input := `# Act 1

### Development

**If the PCs enter:**
- **Immediate consequence:** A
- **Future consequence:** B
- **Recovery:** C
`
	result := ValidateDevelopments(input)
	if result.Valid {
		t.Error("expected failure with fewer than 3 branches")
	}
	found := false
	for _, err := range result.Errors {
		if err.Field == "branch_count" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected branch_count error, got: %v", result.Errors)
	}
}

func TestValidateDevelopments_MixedLanguage(t *testing.T) {
	input := `# Act 1

### Development

**If the PCs enter:**
- **Immediate consequence:** A
- **Future consequence:** B
- **Recovery:** C

**Si los PJs atacan:**
- **Consecuencia inmediata:** D
- **Recuperación:** E
`
	result := ValidateDevelopments(input)
	if result.Valid {
		t.Error("expected failure for mixed English/Spanish markers")
	}
	found := false
	for _, err := range result.Errors {
		if err.Field == "mixed_language" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected mixed_language error, got: %v", result.Errors)
	}
}

func TestValidateDevelopments_MissingRecovery(t *testing.T) {
	input := `# Act 1

### Development

**If the PCs enter:**
- **Immediate consequence:** A
- **Future consequence:** B

**If the PCs sneak:**
- **Immediate consequence:** C
- **Future consequence:** D

**If the PCs talk:**
- **Immediate consequence:** E
- **Future consequence:** F
`
	result := ValidateDevelopments(input)
	if result.Valid {
		t.Error("expected failure when no recovery paths are present")
	}
	found := false
	for _, err := range result.Errors {
		if err.Field == "recovery" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected recovery error, got: %v", result.Errors)
	}
}

func TestValidateBoxedText_EnglishDefaults(t *testing.T) {
	filler := strings.Repeat("You see a long corridor stretching before you. ", 10)
	input := `### Area 1

>> **Read-Aloud Text:** *` + filler + `*
`
	result := ValidateBoxedText(input)
	if !result.Valid {
		t.Errorf("expected valid English boxed text, got errors: %v", result.Errors)
	}
}

func TestValidateBoxedText_Missing(t *testing.T) {
	input := `### Area 1

No boxed text here.
`
	result := ValidateBoxedText(input)
	if result.Valid {
		t.Error("expected failure when boxed text is missing")
	}
}

func TestValidateBoxedText_MixedLanguage(t *testing.T) {
	filler := strings.Repeat("You see a long corridor. ", 10)
	input := `### Area 1

>> **Read-Aloud Text:** *` + filler + `*

>> **Texto para Leer:** *Texto en español.*
`
	result := ValidateBoxedText(input)
	if result.Valid {
		t.Error("expected failure for mixed English/Spanish boxed text labels")
	}
	found := false
	for _, err := range result.Errors {
		if err.Field == "mixed_language" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected mixed_language error, got: %v", result.Errors)
	}
}
