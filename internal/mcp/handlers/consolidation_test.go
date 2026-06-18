package handlers

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pauvalls/grimorio/internal/services"
	"github.com/pauvalls/grimorio/internal/services/consolidation"
)

func writeMDFiles(t *testing.T, baseDir, campaignID string, files map[string]string) {
	t.Helper()
	campDir := filepath.Join(baseDir, campaignID)
	if err := os.MkdirAll(campDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for name, body := range files {
		full := filepath.Join(campDir, name)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(body), 0644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
	}
}

func callConsolidationHandler(t *testing.T, handler serverHandlerFunc, args map[string]any) (map[string]any, *mcp.CallToolResult) {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if result.IsError {
		return nil, result
	}
	if len(result.Content) == 0 {
		t.Fatal("no content in result")
	}
	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", result.Content[0])
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text.Text), &parsed); err != nil {
		t.Fatalf("unmarshal: %v\nbody=%s", err, text.Text)
	}
	return parsed, result
}

// serverHandlerFunc is a local alias to avoid importing server into every test.
type serverHandlerFunc = func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error)

func TestConsolidationHandlers_DetectInconsistencies_FindsDrift(t *testing.T) {
	base := t.TempDir()
	writeMDFiles(t, base, "drift", map[string]string{
		"lore.md":          "Treaty of Ashford 1247.\nTreaty of Ashford 1251.\n",
		"npcs/a.md":        "# Velara\n",
		"npcs/b.md":        "# Velarra\n",
		"bestiary/boss.md": "# Boss\n\nCR 5\n",
		"acts/act1.md":     "# Boss\n\nCR 9\n",
	})
	adapter := services.NewConsolidationAdapter(base)
	h := NewConsolidationHandlers(adapter)

	parsed, result := callConsolidationHandler(t, h.HandleDetectInconsistencies(), map[string]any{
		"campaign": "drift",
	})
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}
	checksRaw, ok := parsed["checks_run"].([]any)
	if !ok || len(checksRaw) == 0 {
		t.Fatalf("expected checks_run to be a non-empty array, got: %v", parsed["checks_run"])
	}
	// We expect at least 6 checks (one per analyzer).
	if len(checksRaw) < 6 {
		t.Errorf("expected at least 6 checks, got %d", len(checksRaw))
	}
	// Remaining issues should not be empty (we injected drift).
	if issues, ok := parsed["remaining_issues"].([]any); !ok || len(issues) == 0 {
		t.Errorf("expected remaining_issues to be non-empty, got: %v", parsed["remaining_issues"])
	}
}

func TestConsolidationHandlers_ConsolidateCampaign_AutoFixRemovesDuplicates(t *testing.T) {
	base := t.TempDir()
	camp := "autofix"
	dup := "Same body\n"
	writeMDFiles(t, base, camp, map[string]string{
		"areas/dup1.md": dup,
		"areas/dup2.md": dup,
	})
	adapter := services.NewConsolidationAdapter(base)
	h := NewConsolidationHandlers(adapter)

	parsed, result := callConsolidationHandler(t, h.HandleConsolidateCampaign(), map[string]any{
		"campaign":  camp,
		"auto_fix":  true,
	})
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}
	fixes, ok := parsed["fixes_applied"].([]any)
	if !ok {
		t.Fatalf("expected fixes_applied array, got: %T (%v)", parsed["fixes_applied"], parsed["fixes_applied"])
	}
	if len(fixes) == 0 {
		t.Errorf("expected at least one auto-fix, got none")
	}
}

func TestConsolidationHandlers_RegenerateIndex_WritesIndexMD(t *testing.T) {
	base := t.TempDir()
	writeMDFiles(t, base, "idx", map[string]string{
		"intro.md":    "# Intro\n",
		"chapters/c1.md": "# Chapter 1\n",
	})
	adapter := services.NewConsolidationAdapter(base)
	h := NewConsolidationHandlers(adapter)

	parsed, result := callConsolidationHandler(t, h.HandleRegenerateIndex(), map[string]any{
		"campaign": "idx",
	})
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}
	if got, _ := parsed["regenerated"].(bool); !got {
		t.Errorf("expected regenerated=true, got %v", parsed)
	}
	indexPath := filepath.Join(base, "idx", "INDEX.md")
	body, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("INDEX.md not written: %v", err)
	}
	if !strings.Contains(string(body), "intro.md") {
		t.Errorf("INDEX.md should link to intro.md, got: %s", body)
	}
	if !strings.Contains(string(body), "chapters/c1.md") {
		t.Errorf("INDEX.md should link to chapters/c1.md, got: %s", body)
	}
}

func TestConsolidationHandlers_VerifyFreshness_ReportsStale(t *testing.T) {
	base := t.TempDir()
	camp := "fresh"
	campDir := filepath.Join(base, camp)
	if err := os.MkdirAll(campDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Make the source newer than the generated file by at least 1s
	// (filesystem mtime resolution).
	now := time.Now()
	if err := os.WriteFile(filepath.Join(campDir, "source.md"), []byte("source content"), 0644); err != nil {
		t.Fatal(err)
	}
	stale := now.Add(-time.Hour)
	if err := os.WriteFile(filepath.Join(campDir, "campaign.md"), []byte("stale content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(filepath.Join(campDir, "campaign.md"), stale, stale); err != nil {
		t.Fatal(err)
	}

	adapter := services.NewConsolidationAdapter(base)
	h := NewConsolidationHandlers(adapter)

	parsed, result := callConsolidationHandler(t, h.HandleVerifyCampaignFreshness(), map[string]any{
		"campaign": camp,
	})
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}
	if stale, _ := parsed["campaign_md_stale"].(bool); !stale {
		t.Errorf("expected campaign_md_stale=true, got %v", parsed)
	}
}

func TestConsolidationHandlers_ResolveAmbiguity_RequiresFields(t *testing.T) {
	adapter := services.NewConsolidationAdapter(t.TempDir())
	h := NewConsolidationHandlers(adapter)

	// Missing required fields → error result.
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"campaign": "x"}
	result, err := h.HandleResolveAmbiguity()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if !result.IsError {
		t.Errorf("expected error for missing question_id and decision")
	}
}

func TestConsolidationHandlers_ConsolidateCampaign_MissingCampaign(t *testing.T) {
	adapter := services.NewConsolidationAdapter(t.TempDir())
	h := NewConsolidationHandlers(adapter)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}
	result, err := h.HandleConsolidateCampaign()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if !result.IsError {
		t.Errorf("expected error for missing campaign argument")
	}
}

func TestConsolidationHandlers_DetectInconsistencies_InvalidArgs(t *testing.T) {
	adapter := services.NewConsolidationAdapter(t.TempDir())
	h := NewConsolidationHandlers(adapter)

	// Pass an args payload that is not a map → handler must return an error.
	req := mcp.CallToolRequest{}
	req.Params.Arguments = "not-a-map"
	result, err := h.HandleDetectInconsistencies()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if !result.IsError {
		t.Errorf("expected error for invalid args shape")
	}
}

func TestConsolidationHandlers_ResolveAmbiguity_AppliesEntityDecision(t *testing.T) {
	base := t.TempDir()
	camp := "decision"
	writeMDFiles(t, base, camp, map[string]string{
		"npcs/a.md": "# Velara the Bold\n",
		"npcs/b.md": "# Velarra the Bold\n",
	})
	adapter := services.NewConsolidationAdapter(base)
	h := NewConsolidationHandlers(adapter)

	parsed, result := callConsolidationHandler(t, h.HandleResolveAmbiguity(), map[string]any{
		"campaign":    camp,
		"question_id": "entity-velara-the-bold-velarra-the-bold",
		"decision":    "Velara the Bold",
	})
	if result.IsError {
		t.Fatalf("unexpected error: %v", result.Content)
	}
	if resolved, _ := parsed["resolved"].(bool); !resolved {
		t.Errorf("expected resolved=true, got %v", parsed)
	}
}

// _ ensures consolidation import is used (some linters complain otherwise).
var _ = consolidation.CampaignFile{}
