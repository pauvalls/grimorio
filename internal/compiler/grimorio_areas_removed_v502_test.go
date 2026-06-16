package compiler

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// projectRoot returns the absolute path to the grimorio repo root by walking
// up from the compiler package's working directory until it finds go.mod.
// The tests in this file must run from the package directory.
func projectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find go.mod walking up from %s", dir)
		}
		dir = parent
	}
}

// TestGrimorioAreasSkillRemoved asserts that the legacy grimorio-areas
// SKILL folder is gone — its deprecation banner from WU3 is no longer
// needed because the entire skill is being eliminated in WU7.
func TestGrimorioAreasSkillRemoved(t *testing.T) {
	root := projectRoot(t)
	skillDir := filepath.Join(root, "skills", "grimorio-areas")
	_, err := os.Stat(skillDir)
	if err == nil {
		t.Errorf("legacy skill directory still exists: %s", skillDir)
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("unexpected stat error for %s: %v", skillDir, err)
	}
}

// TestGrimorioAreasAgentRemoved asserts that the legacy grimorio-areas
// agent file (a thin pointer to the now-removed skill) is gone.
func TestGrimorioAreasAgentRemoved(t *testing.T) {
	root := projectRoot(t)
	agentFile := filepath.Join(root, "agents", "grimorio-areas.md")
	_, err := os.Stat(agentFile)
	if err == nil {
		t.Errorf("legacy agent file still exists: %s", agentFile)
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("unexpected stat error for %s: %v", agentFile, err)
	}
}

// TestAreasTemplateRemoved asserts that the legacy areas.md.tmpl is gone
// — chapters use chapter.md.tmpl now, and the MCP save_areas tool
// referenced this template (also removed in WU7).
func TestAreasTemplateRemoved(t *testing.T) {
	root := projectRoot(t)
	tmplFile := filepath.Join(root, "internal", "compiler", "templates", "areas.md.tmpl")
	_, err := os.Stat(tmplFile)
	if err == nil {
		t.Errorf("legacy areas.md.tmpl still exists: %s", tmplFile)
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("unexpected stat error for %s: %v", tmplFile, err)
	}
}

// TestAreasTemplateVarRemoved asserts that compiler.go no longer declares
// `var areasTemplate string` — the embed of areas.md.tmpl is gone.
func TestAreasTemplateVarRemoved(t *testing.T) {
	root := projectRoot(t)
	src, err := os.ReadFile(filepath.Join(root, "internal", "compiler", "compiler.go"))
	if err != nil {
		t.Fatalf("read compiler.go: %v", err)
	}
	content := string(src)
	if strings.Contains(content, "var areasTemplate string") {
		t.Errorf("compiler.go still declares `var areasTemplate string` — the legacy areas template is removed but the var wasn't")
	}
	if strings.Contains(content, "//go:embed templates/areas.md.tmpl") {
		t.Errorf("compiler.go still embeds templates/areas.md.tmpl — remove the embed directive")
	}
}

// TestSaveAreasToolRemoved asserts that the MCP server no longer registers
// a `save_areas` tool — it is replaced by `save_chapter` (which always
// used the chapters/ structure internally).
func TestSaveAreasToolRemoved(t *testing.T) {
	root := projectRoot(t)
	text, err := os.ReadFile(filepath.Join(root, "internal", "mcp", "server.go"))
	if err != nil {
		t.Fatalf("read server.go: %v", err)
	}
	if strings.Contains(string(text), `"save_areas"`) {
		t.Errorf("server.go still references the `save_areas` tool — remove its s.AddTool block")
	}
}
