package validators

import (
	"os"
	"strings"
	"testing"
)

func TestIntegratorAgentReferencesConsistencyTools(t *testing.T) {
	data, err := os.ReadFile(os.Getenv("HOME") + "/.config/opencode/plugins/grimorio/agents/grimorio-integrator.md")
	if err != nil {
		t.Fatalf("failed to read integrator agent: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "grimorio_check_consistency") {
		t.Error("integrator agent missing reference to grimorio_check_consistency")
	}
	if !strings.Contains(content, "grimorio_process_consistency_gate") {
		t.Error("integrator agent missing reference to grimorio_process_consistency_gate")
	}
	if !strings.Contains(content, "grimorio_validate_canon") {
		t.Error("integrator agent missing reference to grimorio_validate_canon")
	}
}
