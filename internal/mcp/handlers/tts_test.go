package handlers

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pauvalls/grimorio/internal/dm"
	"github.com/pauvalls/grimorio/internal/services"
)

// mockTTSPipeline is a minimal NarratorPipeline mock for handler tests.
type mockTTSPipeline struct {
	running bool
}

func (m *mockTTSPipeline) Narrate(ctx context.Context, text string) error { return nil }
func (m *mockTTSPipeline) Stop() error                                      { m.running = false; return nil }
func (m *mockTTSPipeline) Skip() error                                      { return nil }
func (m *mockTTSPipeline) Pause() error                                     { return nil }
func (m *mockTTSPipeline) Resume() error                                    { return nil }
func (m *mockTTSPipeline) IsRunning() bool                                  { return m.running }

type mockTTSLifecycle struct {
	installed bool
	running   bool
}

func (m *mockTTSLifecycle) IsInstalled() bool               { return m.installed }
func (m *mockTTSLifecycle) IsRunning() bool                 { return m.running }
func (m *mockTTSLifecycle) Start(ctx context.Context) error { m.running = true; return nil }
func (m *mockTTSLifecycle) Stop(ctx context.Context) error  { m.running = false; return nil }
func (m *mockTTSLifecycle) Restart(ctx context.Context) error { return nil }

func setupTTSService(t *testing.T, enabled, installed bool) *services.TTSService {
	pipe := &mockTTSPipeline{}
	life := &mockTTSLifecycle{installed: installed, running: installed}
	store := services.NewFileCampaignVoiceStore(t.TempDir())
	return services.NewTTSService(pipe, life, store, enabled)
}

func TestHandleSetDMMode(t *testing.T) {
	svc := setupTTSService(t, true, true)
	h := NewTTSHandlers(svc)
	handler := h.HandleSetDMMode()

	// Switch to TTS
	res, err := handler(context.Background(), newToolRequest("set_dm_mode", map[string]any{
		"mode": "tts",
	}))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error result: %v", res.Content)
	}
	if !strings.Contains(res.Content[0].(mcp.TextContent).Text, "tts") {
		t.Errorf("expected response to contain tts, got %v", res.Content)
	}

	// Switch back to written
	res2, err := handler(context.Background(), newToolRequest("set_dm_mode", map[string]any{
		"mode": "written",
	}))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res2.IsError {
		t.Fatalf("unexpected error result: %v", res2.Content)
	}
}

func TestHandleSetDMModeFallbackWhenUnavailable(t *testing.T) {
	// TTS not installed
	svc := setupTTSService(t, true, false)
	h := NewTTSHandlers(svc)
	handler := h.HandleSetDMMode()

	res, err := handler(context.Background(), newToolRequest("set_dm_mode", map[string]any{
		"mode": "tts",
	}))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if !res.IsError {
		t.Error("expected error when TTS is not available")
	}
}

func TestHandleGetTTSStatus(t *testing.T) {
	svc := setupTTSService(t, true, true)
	if err := svc.SetMode(dm.ModeTTS); err != nil {
		t.Fatalf("SetMode error: %v", err)
	}

	h := NewTTSHandlers(svc)
	handler := h.HandleGetTTSStatus()

	res, err := handler(context.Background(), newToolRequest("get_tts_status", map[string]any{}))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error result: %v", res.Content)
	}

	var status dm.TTSStatus
	text := res.Content[0].(mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &status); err != nil {
		t.Fatalf("failed to unmarshal status: %v", err)
	}
	if !status.Enabled {
		t.Error("expected Enabled = true")
	}
	if status.Mode != "tts" {
		t.Errorf("expected Mode = tts, got %q", status.Mode)
	}
	if !status.Available {
		t.Error("expected Available = true")
	}
}

func TestHandleAssignNPCVoice(t *testing.T) {
	svc := setupTTSService(t, true, true)
	h := NewTTSHandlers(svc)
	handler := h.HandleAssignNPCVoice()

	res, err := handler(context.Background(), newToolRequest("assign_npc_voice", map[string]any{
		"campaign_id":  "campaign-1",
		"npc_name":     "Goblin Chief",
		"voice_prompt": "gruff and low",
	}))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error result: %v", res.Content)
	}

	var result map[string]any
	text := res.Content[0].(mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if result["voice_id"] != "npc-goblin-chief" {
		t.Errorf("expected voice_id npc-goblin-chief, got %v", result["voice_id"])
	}
	if result["created"] != true {
		t.Errorf("expected created = true, got %v", result["created"])
	}
}

func TestHandleTTSControl(t *testing.T) {
	svc := setupTTSService(t, true, true)
	h := NewTTSHandlers(svc)
	handler := h.HandleTTSControl()

	for _, action := range []string{"skip", "stop", "pause", "resume"} {
		res, err := handler(context.Background(), newToolRequest("tts_control", map[string]any{
			"action": action,
		}))
		if err != nil {
			t.Fatalf("handler error for %s: %v", action, err)
		}
		if res.IsError {
			t.Errorf("unexpected error for action %s: %v", action, res.Content)
		}
	}

	// Unknown action
	res, err := handler(context.Background(), newToolRequest("tts_control", map[string]any{
		"action": "jump",
	}))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if !res.IsError {
		t.Error("expected error for unknown action")
	}
}

func TestHandleListTTSVoices(t *testing.T) {
	svc := setupTTSService(t, true, true)
	if _, _, err := svc.AssignNPCVoice("campaign-1", "NPC1", "voice1"); err != nil {
		t.Fatalf("AssignNPCVoice error: %v", err)
	}

	h := NewTTSHandlers(svc)
	handler := h.HandleListTTSVoices()

	res, err := handler(context.Background(), newToolRequest("list_tts_voices", map[string]any{
		"campaign_id": "campaign-1",
	}))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error result: %v", res.Content)
	}

	var voices map[string]string
	text := res.Content[0].(mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &voices); err != nil {
		t.Fatalf("failed to unmarshal voices: %v", err)
	}
	if len(voices) != 1 {
		t.Errorf("expected 1 voice, got %d", len(voices))
	}
}

func TestHandleListTTSVoicesMissingCampaign(t *testing.T) {
	svc := setupTTSService(t, true, true)
	h := NewTTSHandlers(svc)
	handler := h.HandleListTTSVoices()

	res, err := handler(context.Background(), newToolRequest("list_tts_voices", map[string]any{}))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if !res.IsError {
		t.Error("expected error when campaign_id is missing")
	}
}

func TestTTSHandlersFallbackWhenDisabled(t *testing.T) {
	// TTS disabled: all handlers should still function but report unavailable
	svc := services.NewTTSService(nil, nil, nil, false)
	h := NewTTSHandlers(svc)

	// set_dm_mode to tts should fail
	res, _ := h.HandleSetDMMode()(context.Background(), newToolRequest("set_dm_mode", map[string]any{
		"mode": "tts",
	}))
	if !res.IsError {
		t.Error("expected error setting TTS mode when disabled")
	}

	// get_tts_status should report unavailable
	res2, _ := h.HandleGetTTSStatus()(context.Background(), newToolRequest("get_tts_status", map[string]any{}))
	if res2.IsError {
		t.Fatal("get_tts_status should not error")
	}
	var status dm.TTSStatus
	text := res2.Content[0].(mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &status); err != nil {
		t.Fatalf("failed to unmarshal status: %v", err)
	}
	if status.Available {
		t.Error("expected Available = false when TTS is disabled")
	}
}

func TestHandleTTSSpeak(t *testing.T) {
	svc := setupTTSService(t, true, true)
	if err := svc.SetMode(dm.ModeTTS); err != nil {
		t.Fatalf("SetMode error: %v", err)
	}

	h := NewTTSHandlers(svc)
	handler := h.HandleTTSSpeak()

	res, err := handler(context.Background(), newToolRequest("tts_speak", map[string]any{
		"text": "El dragón ruge sobre la montaña",
	}))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error result: %v", res.Content)
	}

	// The response should contain the spoken text (displayed on screen)
	text := res.Content[0].(mcp.TextContent).Text
	if text != "El dragón ruge sobre la montaña" {
		t.Errorf("expected text to be returned, got %q", text)
	}
}

func TestHandleTTSSpeakEmptyText(t *testing.T) {
	svc := setupTTSService(t, true, true)
	h := NewTTSHandlers(svc)
	handler := h.HandleTTSSpeak()

	res, err := handler(context.Background(), newToolRequest("tts_speak", map[string]any{
		"text": "",
	}))
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if !res.IsError {
		t.Error("expected error when text is empty")
	}
}
