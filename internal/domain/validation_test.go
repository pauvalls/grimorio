package domain

import "testing"

func TestValidationReport_Validate(t *testing.T) {
	tests := []struct {
		name    string
		report  ValidationReport
		wantErr bool
	}{
		{
			name: "valid approved report",
			report: ValidationReport{
				ArtifactID:    "prop-001",
				ArtifactType:  "act",
				OverallStatus: "approved",
				Checks:        []CheckResult{},
			},
			wantErr: false,
		},
		{
			name: "valid rejected report",
			report: ValidationReport{
				ArtifactID:    "prop-001",
				ArtifactType:  "act",
				OverallStatus: "rejected",
				Checks: []CheckResult{
					{Rule: "npc_alive_check", Passed: false, Severity: "critical", Message: "NPC is dead"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing artifact ID",
			report: ValidationReport{
				ArtifactType:  "act",
				OverallStatus: "approved",
			},
			wantErr: true,
		},
		{
			name: "missing overall status",
			report: ValidationReport{
				ArtifactID:   "prop-001",
				ArtifactType: "act",
			},
			wantErr: true,
		},
		{
			name: "invalid overall status",
			report: ValidationReport{
				ArtifactID:    "prop-001",
				ArtifactType:  "act",
				OverallStatus: "maybe",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.report.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidationReport_AddCheck(t *testing.T) {
	report := &ValidationReport{
		ArtifactID:    "prop-001",
		OverallStatus: "approved",
	}

	report.AddCheck("npc_alive_check", true, "info", "All NPCs are alive", "act_1")
	if len(report.Checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(report.Checks))
	}
	if report.Checks[0].Rule != "npc_alive_check" {
		t.Fatalf("expected rule 'npc_alive_check', got %s", report.Checks[0].Rule)
	}
}

func TestValidationReport_ComputeOverallStatus(t *testing.T) {
	tests := []struct {
		name           string
		checks         []CheckResult
		expectedStatus string
	}{
		{
			name:           "all passing",
			checks:         []CheckResult{{Passed: true, Severity: "info"}},
			expectedStatus: "approved",
		},
		{
			name:           "one warning",
			checks:         []CheckResult{{Passed: false, Severity: "warning"}},
			expectedStatus: "warning",
		},
		{
			name:           "one error",
			checks:         []CheckResult{{Passed: false, Severity: "error"}},
			expectedStatus: "rejected",
		},
		{
			name:           "one critical",
			checks:         []CheckResult{{Passed: false, Severity: "critical"}},
			expectedStatus: "rejected",
		},
		{
			name:           "mixed with critical",
			checks:         []CheckResult{{Passed: true, Severity: "info"}, {Passed: false, Severity: "critical"}},
			expectedStatus: "rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &ValidationReport{
				ArtifactID:    "prop-001",
				OverallStatus: "approved",
				Checks:        tt.checks,
			}
			report.ComputeOverallStatus()
			if report.OverallStatus != tt.expectedStatus {
				t.Fatalf("expected status %q, got %q", tt.expectedStatus, report.OverallStatus)
			}
		})
	}
}

func TestValidationReport_AddSuggestion(t *testing.T) {
	report := &ValidationReport{
		ArtifactID:    "prop-001",
		OverallStatus: "rejected",
	}

	report.AddSuggestion("NPC is dead", "Replace with a messenger", "Canon says NPC died in session 2")
	if len(report.Suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(report.Suggestions))
	}
	if report.Suggestions[0].Problem != "NPC is dead" {
		t.Fatalf("expected problem 'NPC is dead', got %s", report.Suggestions[0].Problem)
	}
}
