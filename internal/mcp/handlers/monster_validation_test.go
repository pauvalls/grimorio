package handlers

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pauvalls/grimorio/internal/services/monster"
)

// makeTestCampaignBestiary creates a campaign directory with a
// bestiary/bestiary.md file containing a single WotC stat block
// (a Goblin). The base directory is the campaigns root; the campaign
// is created at <baseDir>/<campaignID>/bestiary/bestiary.md.
func makeTestCampaignBestiary(t *testing.T, baseDir, campaignID, body string) {
	t.Helper()
	bestiaryDir := filepath.Join(baseDir, campaignID, "bestiary")
	if err := os.MkdirAll(bestiaryDir, 0o755); err != nil {
		t.Fatalf("mkdir bestiary: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bestiaryDir, "bestiary.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write bestiary: %v", err)
	}
}

const goblinStatBlock = `## Goblin

*Small humanoid, neutral evil*

**Initiative** +2 (+12) (Dexterity)

**Armor Class** 15 (Leather Armor, Shield)

**Hit Points** 7 (2d6)

**Speed** 30 ft.

| STR | DEX | CON | INT | WIS | CHA |
|:---:|:---:|:---:|:---:|:---:|:---:|
| 8 (-1) | 14 (+2) | 10 (+0) | 10 (+0) | 8 (-1) | 8 (-1) |

**Senses** darkvision 60 ft., passive Perception 9

**Languages** Common, Goblin

**Challenge** 1/4 (50 XP)

***Nimble Escape.*** The goblin can take the Disengage or Hide action as a bonus action on each of its turns.
`

// callValidateMonsterHandler invokes the validate_monster handler with
// the given args and returns the parsed JSON body (or nil on error).
func callValidateMonsterHandler(t *testing.T, h serverHandlerFunc, args map[string]any) (map[string]any, *mcp.CallToolResult) {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	result, err := h(context.Background(), req)
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
		t.Fatalf("unmarshal result: %v\nbody=%s", err, text.Text)
	}
	return parsed, result
}

func TestValidateMonster_Markdown_HappyPath(t *testing.T) {
	t.Parallel()
	h := NewMonsterValidationHandlers(t.TempDir())
	parsed, result := callValidateMonsterHandler(t, h.HandleValidateMonster(), map[string]any{
		"markdown": goblinStatBlock,
	})
	if result.IsError {
		t.Fatalf("expected success, got error result: %+v", result)
	}
	if got, _ := parsed["severity"].(string); got != "ok" {
		t.Errorf("severity = %q, want ok", got)
	}
	off, _ := parsed["official_cr"].(float64)
	calc, _ := parsed["calculated_cr"].(float64)
	if off != 0.25 {
		t.Errorf("official_cr = %v, want 0.25", off)
	}
	if calc != 0.25 {
		t.Errorf("calculated_cr = %v, want 0.25", calc)
	}
	if _, ok := parsed["findings"].([]any); !ok {
		t.Errorf("expected findings to be an array, got %T (%v)", parsed["findings"], parsed["findings"])
	}
	if _, ok := parsed["suggestions"].([]any); !ok {
		t.Errorf("expected suggestions to be an array, got %T (%v)", parsed["suggestions"], parsed["suggestions"])
	}
}

func TestValidateMonster_MonsterName_HappyPath(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	makeTestCampaignBestiary(t, base, "test", goblinStatBlock)
	h := NewMonsterValidationHandlers(base)
	parsed, result := callValidateMonsterHandler(t, h.HandleValidateMonster(), map[string]any{
		"monster_name": "Goblin",
		"campaign":     "test",
	})
	if result.IsError {
		t.Fatalf("expected success, got error result: %+v", result)
	}
	if got, _ := parsed["severity"].(string); got != "ok" {
		t.Errorf("severity = %q, want ok", got)
	}
}

func TestValidateMonster_MissingInput(t *testing.T) {
	t.Parallel()
	h := NewMonsterValidationHandlers(t.TempDir())
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}
	result, err := h.HandleValidateMonster()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error for missing input, got success")
	}
	text, _ := result.Content[0].(mcp.TextContent)
	if !strings.Contains(text.Text, "monster_name or markdown required") {
		t.Errorf("error message = %q, want contains 'monster_name or markdown required'", text.Text)
	}
}

func TestValidateMonster_MonsterNotFound(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	makeTestCampaignBestiary(t, base, "test", goblinStatBlock)
	h := NewMonsterValidationHandlers(base)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"monster_name": "Dragon",
		"campaign":     "test",
	}
	result, err := h.HandleValidateMonster()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error for missing monster, got success")
	}
	text, _ := result.Content[0].(mcp.TextContent)
	if !strings.Contains(strings.ToLower(text.Text), "not_found") {
		t.Errorf("error message = %q, want contains 'not_found'", text.Text)
	}
	if !strings.Contains(text.Text, "Dragon") {
		t.Errorf("error message = %q, want contains 'Dragon'", text.Text)
	}
}

func TestValidateMonster_MalformedMarkdown(t *testing.T) {
	t.Parallel()
	h := NewMonsterValidationHandlers(t.TempDir())
	req := mcp.CallToolRequest{}
	// A non-numeric Armor Class line triggers a *ParseError.
	bad := "## Bad Monster\n\n**Armor Class** not-a-number\n"
	req.Params.Arguments = map[string]any{
		"markdown": bad,
	}
	result, err := h.HandleValidateMonster()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error for malformed markdown, got success")
	}
	text, _ := result.Content[0].(mcp.TextContent)
	if !strings.Contains(strings.ToLower(text.Text), "parse") {
		t.Errorf("error message = %q, want contains 'parse'", text.Text)
	}
}

func TestSuggestMonsterCR_MarkdownOutput(t *testing.T) {
	t.Parallel()
	h := NewMonsterValidationHandlers(t.TempDir())
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"target_cr": float64(5),
		"output":    "markdown",
	}
	result, err := h.HandleSuggestMonsterCR()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %+v", result)
	}
	text, _ := result.Content[0].(mcp.TextContent)
	// The rendered stat block should at least contain a CR line.
	if !strings.Contains(text.Text, "Challenge") {
		t.Errorf("markdown output = %q, want contains 'Challenge'", text.Text)
	}
}

func TestSuggestMonsterCR_JSONOutput(t *testing.T) {
	t.Parallel()
	h := NewMonsterValidationHandlers(t.TempDir())
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"target_cr": float64(5),
		"output":    "json",
	}
	result, err := h.HandleSuggestMonsterCR()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %+v", result)
	}
	text, _ := result.Content[0].(mcp.TextContent)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text.Text), &parsed); err != nil {
		t.Fatalf("unmarshal json output: %v\nbody=%s", err, text.Text)
	}
	// The JSON should at least have AC and HP.
	ac, ok := parsed["AC"].(float64)
	if !ok {
		t.Errorf("expected AC field as number, got %T (%v)", parsed["AC"], parsed["AC"])
	}
	if ac != 15 {
		t.Errorf("AC = %v, want 15 for CR 5", ac)
	}
}

func TestSuggestMonsterCR_OutOfRange(t *testing.T) {
	t.Parallel()
	h := NewMonsterValidationHandlers(t.TempDir())
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"target_cr": float64(31),
	}
	result, err := h.HandleSuggestMonsterCR()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error for CR > 30, got success")
	}
	text, _ := result.Content[0].(mcp.TextContent)
	if !strings.Contains(strings.ToLower(text.Text), "cr") {
		t.Errorf("error message = %q, want contains 'CR'", text.Text)
	}
}

func TestSuggestMonsterCR_SubInteger(t *testing.T) {
	t.Parallel()
	h := NewMonsterValidationHandlers(t.TempDir())
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"target_cr": 0.25,
		"output":    "json",
	}
	result, err := h.HandleSuggestMonsterCR()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success for CR 0.25, got error: %+v", result)
	}
	text, _ := result.Content[0].(mcp.TextContent)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text.Text), &parsed); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	cr, ok := parsed["CR"].(float64)
	if !ok {
		t.Errorf("expected CR field as number, got %T (%v)", parsed["CR"], parsed["CR"])
	}
	if cr != 0.25 {
		t.Errorf("CR = %v, want 0.25", cr)
	}
}

func TestAuditMonsterCR_HappyPath(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	makeTestCampaignBestiary(t, base, "test", goblinStatBlock)
	h := NewMonsterValidationHandlers(base)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"campaign": "test",
	}
	result, err := h.HandleAuditMonsterCR()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %+v", result)
	}
	text, _ := result.Content[0].(mcp.TextContent)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text.Text), &parsed); err != nil {
		t.Fatalf("unmarshal: %v\nbody=%s", err, text.Text)
	}
	if got, _ := parsed["campaign"].(string); got != "test" {
		t.Errorf("campaign = %q, want 'test'", got)
	}
	monsters, ok := parsed["monsters"].([]any)
	if !ok || len(monsters) == 0 {
		t.Fatalf("expected non-empty monsters array, got %v", parsed["monsters"])
	}
	summary, ok := parsed["summary"].(map[string]any)
	if !ok {
		t.Fatalf("expected summary object, got %T", parsed["summary"])
	}
	total, _ := summary["total"].(float64)
	if total < 1 {
		t.Errorf("summary.total = %v, want >= 1", total)
	}
}

func TestAuditMonsterCR_CampaignNotFound(t *testing.T) {
	t.Parallel()
	h := NewMonsterValidationHandlers(t.TempDir())
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"campaign": "nonexistent",
	}
	result, err := h.HandleAuditMonsterCR()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error for missing campaign, got success")
	}
	text, _ := result.Content[0].(mcp.TextContent)
	if !strings.Contains(strings.ToLower(text.Text), "not found") {
		t.Errorf("error message = %q, want contains 'not found'", text.Text)
	}
}

func TestAuditMonsterCR_MissingCampaign(t *testing.T) {
	t.Parallel()
	h := NewMonsterValidationHandlers(t.TempDir())
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}
	result, err := h.HandleAuditMonsterCR()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if !result.IsError {
		t.Fatalf("expected error for missing campaign arg, got success")
	}
}

// _ ensures monster import is used (some linters complain otherwise).
var _ = monster.SeverityOK

// --- Triangulation tests: drive the handlers through paths not
// covered by the happy/error cases above.

func TestValidateMonster_MajorDrift(t *testing.T) {
	t.Parallel()
	// HP=999 / AC=99 against CR 1/4 — should report severity=major.
	h := NewMonsterValidationHandlers(t.TempDir())
	bad := `## Outlier

*Medium humanoid, unaligned*

**Armor Class** 99
**Hit Points** 999
**Speed** 30 ft.
**Challenge** 1/4 (50 XP)
`
	parsed, result := callValidateMonsterHandler(t, h.HandleValidateMonster(), map[string]any{
		"markdown": bad,
	})
	if result.IsError {
		t.Fatalf("expected success, got error: %+v", result)
	}
	if got, _ := parsed["severity"].(string); got != "major" {
		t.Errorf("severity = %q, want major", got)
	}
	if delta, _ := parsed["delta"].(float64); delta < 1.5 {
		t.Errorf("delta = %v, want >= 1.5", delta)
	}
	suggestions, _ := parsed["suggestions"].([]any)
	if len(suggestions) == 0 {
		t.Errorf("expected non-empty suggestions for major drift, got empty")
	}
}

func TestSuggestMonsterCR_ConceptFlows(t *testing.T) {
	t.Parallel()
	h := NewMonsterValidationHandlers(t.TempDir())
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"target_cr": 10,
		"concept":   "fire-breathing dragon",
		"output":    "json",
	}
	result, err := h.HandleSuggestMonsterCR()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %+v", result)
	}
	text, _ := result.Content[0].(mcp.TextContent)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text.Text), &parsed); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	cr, _ := parsed["CR"].(float64)
	if cr != 10 {
		t.Errorf("CR = %v, want 10", cr)
	}
	hp, _ := parsed["HP"].(float64)
	// CR 10 band: HP 206-220
	if hp < 200 || hp > 230 {
		t.Errorf("HP = %v, want in [206,220] for CR 10", hp)
	}
}

func TestAuditMonsterCR_SummaryCounts(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	body := `## Minor

*Small humanoid, neutral evil*

**Armor Class** 13
**Hit Points** 80
**Speed** 30 ft.
**Challenge** 1 (200 XP)

## Major

*Large dragon, lawful evil*

**Armor Class** 10
**Hit Points** 50
**Speed** 30 ft.
**Challenge** 20 (25000 XP)
`
	makeTestCampaignBestiary(t, base, "mixed", body)
	h := NewMonsterValidationHandlers(base)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"campaign": "mixed",
	}
	result, err := h.HandleAuditMonsterCR()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %+v", result)
	}
	text, _ := result.Content[0].(mcp.TextContent)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text.Text), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	summary, ok := parsed["summary"].(map[string]any)
	if !ok {
		t.Fatalf("expected summary object, got %T", parsed["summary"])
	}
	total, _ := summary["total"].(float64)
	if total != 2 {
		t.Errorf("summary.total = %v, want 2", total)
	}
	major, _ := summary["major"].(float64)
	if major < 1 {
		t.Errorf("summary.major = %v, want >= 1 (Major monster is way off-band)", major)
	}
}
