package domain

// ValidationReport is the result of validating a content proposal
type ValidationReport struct {
	ArtifactID    string        `json:"artifact_id"`
	ArtifactType  string        `json:"artifact_type"`
	OverallStatus string        `json:"overall_status"`
	Checks        []CheckResult `json:"checks"`
	Suggestions   []Suggestion  `json:"suggestions"`
}

// CheckResult represents the outcome of a single validation check
type CheckResult struct {
	Rule     string `json:"rule"`
	Passed   bool   `json:"passed"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Location string `json:"location,omitempty"`
}

// Suggestion provides a fix recommendation for a failed check
type Suggestion struct {
	Problem   string `json:"problem"`
	Fix       string `json:"fix"`
	Rationale string `json:"rationale"`
}

// ValidationLevel indicates the severity threshold for validation
type ValidationLevel string

const (
	ValidationLevelInfo     ValidationLevel = "info"
	ValidationLevelWarning  ValidationLevel = "warning"
	ValidationLevelError    ValidationLevel = "error"
	ValidationLevelCritical ValidationLevel = "critical"
)

// ConsistencyReport provides a health overview of the campaign
type ConsistencyReport struct {
	CampaignID     string        `json:"campaign_id"`
	OverallHealth  string        `json:"overall_health"`
	TotalChecks    int           `json:"total_checks"`
	Passed         int           `json:"passed"`
	Warnings       int           `json:"warnings"`
	Errors         int           `json:"errors"`
	Criticals      int           `json:"criticals"`
	Issues         []CheckResult `json:"issues"`
}

// Validate checks if the validation report is structurally valid
func (v *ValidationReport) Validate() error {
	if v.ArtifactID == "" {
		return NewValidationError("artifact_id", "artifact ID is required")
	}
	if v.OverallStatus == "" {
		return NewValidationError("overall_status", "overall status is required")
	}
	switch v.OverallStatus {
	case "approved", "warning", "rejected":
		// valid
	default:
		return NewValidationError("overall_status", "invalid status: "+v.OverallStatus)
	}
	return nil
}

// AddCheck adds a check result to the report
func (v *ValidationReport) AddCheck(rule string, passed bool, severity, message, location string) {
	v.Checks = append(v.Checks, CheckResult{
		Rule:     rule,
		Passed:   passed,
		Severity: severity,
		Message:  message,
		Location: location,
	})
}

// ComputeOverallStatus recalculates the overall status based on checks
func (v *ValidationReport) ComputeOverallStatus() {
	hasCritical := false
	hasError := false
	hasWarning := false

	for _, check := range v.Checks {
		if !check.Passed {
			switch check.Severity {
			case "critical":
				hasCritical = true
			case "error":
				hasError = true
			case "warning":
				hasWarning = true
			}
		}
	}

	switch {
	case hasCritical || hasError:
		v.OverallStatus = "rejected"
	case hasWarning:
		v.OverallStatus = "warning"
	default:
		v.OverallStatus = "approved"
	}
}

// AddSuggestion adds a suggestion to the report
func (v *ValidationReport) AddSuggestion(problem, fix, rationale string) {
	v.Suggestions = append(v.Suggestions, Suggestion{
		Problem:   problem,
		Fix:       fix,
		Rationale: rationale,
	})
}
