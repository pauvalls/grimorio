package dm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultTTSConfig(t *testing.T) {
	cfg := DefaultTTSConfig()
	if cfg == nil {
		t.Fatal("DefaultTTSConfig() returned nil")
	}
	if cfg.ServerURL != "ws://localhost:8765/tts" {
		t.Errorf("expected ServerURL 'ws://localhost:8765/tts', got '%s'", cfg.ServerURL)
	}
	if cfg.Enabled != false {
		t.Errorf("expected Enabled false, got %v", cfg.Enabled)
	}
	if cfg.PreloadNext != true {
		t.Errorf("expected PreloadNext true, got %v", cfg.PreloadNext)
	}
	if cfg.ShowSubtitles != true {
		t.Errorf("expected ShowSubtitles true, got %v", cfg.ShowSubtitles)
	}
}

func TestDMModeConstants(t *testing.T) {
	if ModeWritten != "written" {
		t.Errorf("expected ModeWritten 'written', got '%s'", ModeWritten)
	}
	if ModeTTS != "tts" {
		t.Errorf("expected ModeTTS 'tts', got '%s'", ModeTTS)
	}
}

func TestVoiceRegistry_GetSet(t *testing.T) {
	tmpDir := t.TempDir()
	reg := NewVoiceRegistry(tmpDir)

	entry := NPCVoiceEntry{
		NPCID:       "npc-001",
		NPCName:     "Elara",
		VoicePrompt: "warm female voice, soft-spoken",
		Language:    "es",
		Speed:       1.0,
	}

	if err := reg.SetVoice("npc-001", entry); err != nil {
		t.Fatalf("SetVoice failed: %v", err)
	}

	got, err := reg.GetVoice("npc-001")
	if err != nil {
		t.Fatalf("GetVoice failed: %v", err)
	}
	if got.NPCName != "Elara" {
		t.Errorf("expected NPCName 'Elara', got '%s'", got.NPCName)
	}
}

func TestVoiceRegistry_Get_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	reg := NewVoiceRegistry(tmpDir)

	_, err := reg.GetVoice("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent voice")
	}
}

func TestVoiceRegistry_ListVoices(t *testing.T) {
	tmpDir := t.TempDir()
	reg := NewVoiceRegistry(tmpDir)

	_ = reg.SetVoice("npc-001", NPCVoiceEntry{NPCName: "Elara"})
	_ = reg.SetVoice("npc-002", NPCVoiceEntry{NPCName: "Grom"})

	voices := reg.ListVoices()
	if len(voices) != 2 {
		t.Errorf("expected 2 voices, got %d", len(voices))
	}
}

func TestVoiceRegistry_DeleteVoice(t *testing.T) {
	tmpDir := t.TempDir()
	reg := NewVoiceRegistry(tmpDir)

	_ = reg.SetVoice("npc-001", NPCVoiceEntry{NPCName: "Elara"})
	if err := reg.DeleteVoice("npc-001"); err != nil {
		t.Fatalf("DeleteVoice failed: %v", err)
	}

	_, err := reg.GetVoice("npc-001")
	if err == nil {
		t.Error("expected error after deleting voice")
	}
}

func TestVoiceRegistry_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	reg1 := NewVoiceRegistry(tmpDir)

	_ = reg1.SetVoice("npc-001", NPCVoiceEntry{
		NPCName:     "Elara",
		VoicePrompt: "warm female voice",
		Language:    "es",
		Speed:       1.2,
	})

	// Create a new registry pointing to the same dir
	reg2 := NewVoiceRegistry(tmpDir)
	got, err := reg2.GetVoice("npc-001")
	if err != nil {
		t.Fatalf("GetVoice from new registry failed: %v", err)
	}
	if got.NPCName != "Elara" {
		t.Errorf("expected NPCName 'Elara' after reload, got '%s'", got.NPCName)
	}
	if got.Speed != 1.2 {
		t.Errorf("expected Speed 1.2 after reload, got %f", got.Speed)
	}
}

func TestVoiceRegistry_CorruptFile(t *testing.T) {
	tmpDir := t.TempDir()
	corruptPath := filepath.Join(tmpDir, "voices.json")
	_ = os.WriteFile(corruptPath, []byte("not json"), 0644)

	reg := NewVoiceRegistry(tmpDir)
	// Should survive corrupt file and start fresh
	voices := reg.ListVoices()
	if len(voices) != 0 {
		t.Errorf("expected 0 voices after corrupt load, got %d", len(voices))
	}
}
