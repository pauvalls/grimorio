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

	// Check for MCP tools in grimorio_mcp array (format: "tool_name" not "grimorio_tool_name")
	if !strings.Contains(content, `"check_consistency"`) && !strings.Contains(content, `check_consistency(`) {
		t.Error("integrator agent missing reference to check_consistency")
	}
	if !strings.Contains(content, `"process_consistency_gate"`) && !strings.Contains(content, `process_consistency_gate(`) {
		t.Error("integrator agent missing reference to process_consistency_gate")
	}
	if !strings.Contains(content, `"validate_canon"`) && !strings.Contains(content, `validate_canon(`) {
		t.Error("integrator agent missing reference to validate_canon")
	}
}
