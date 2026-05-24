package dm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// NPCVoiceEntry stores voice design metadata for a single NPC.
type NPCVoiceEntry struct {
	NPCID       string  `json:"npc_id"`
	NPCName     string  `json:"npc_name"`
	VoicePrompt string  `json:"voice_prompt"`
	Language    string  `json:"language"`
	Speed       float64 `json:"speed"`
}

// VoiceRegistry provides persistent storage and lookup for NPC voices.
type VoiceRegistry interface {
	GetVoice(npcID string) (NPCVoiceEntry, error)
	SetVoice(npcID string, entry NPCVoiceEntry) error
	ListVoices() map[string]NPCVoiceEntry
	DeleteVoice(npcID string) error
}

// FileVoiceRegistry implements VoiceRegistry using a JSON file on disk.
type FileVoiceRegistry struct {
	mu       sync.RWMutex
	cacheDir string
	voices   map[string]NPCVoiceEntry
}

// NewVoiceRegistry creates a new FileVoiceRegistry backed by JSON in cacheDir.
func NewVoiceRegistry(cacheDir string) *FileVoiceRegistry {
	reg := &FileVoiceRegistry{
		cacheDir: cacheDir,
		voices:   make(map[string]NPCVoiceEntry),
	}
	_ = reg.load() // best-effort load on init
	return reg
}

func (r *FileVoiceRegistry) filePath() string {
	return filepath.Join(r.cacheDir, "voices.json")
}

func (r *FileVoiceRegistry) load() error {
	data, err := os.ReadFile(r.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read voice registry: %w", err)
	}

	var loaded map[string]NPCVoiceEntry
	if err := json.Unmarshal(data, &loaded); err != nil {
		return fmt.Errorf("failed to unmarshal voice registry: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.voices = loaded
	return nil
}

func (r *FileVoiceRegistry) save() error {
	r.mu.RLock()
	data, err := json.MarshalIndent(r.voices, "", "  ")
	r.mu.RUnlock()
	if err != nil {
		return fmt.Errorf("failed to marshal voice registry: %w", err)
	}

	if err := os.MkdirAll(r.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	if err := os.WriteFile(r.filePath(), data, 0644); err != nil {
		return fmt.Errorf("failed to write voice registry: %w", err)
	}
	return nil
}

// GetVoice retrieves a voice entry by NPC ID.
func (r *FileVoiceRegistry) GetVoice(npcID string) (NPCVoiceEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, ok := r.voices[npcID]
	if !ok {
		return NPCVoiceEntry{}, fmt.Errorf("voice not found for npc %q", npcID)
	}
	return entry, nil
}

// SetVoice stores or updates a voice entry for an NPC.
func (r *FileVoiceRegistry) SetVoice(npcID string, entry NPCVoiceEntry) error {
	r.mu.Lock()
	r.voices[npcID] = entry
	r.mu.Unlock()

	if err := r.save(); err != nil {
		return fmt.Errorf("failed to save voice registry: %w", err)
	}
	return nil
}

// ListVoices returns a copy of all registered voices.
func (r *FileVoiceRegistry) ListVoices() map[string]NPCVoiceEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]NPCVoiceEntry, len(r.voices))
	for k, v := range r.voices {
		result[k] = v
	}
	return result
}

// DeleteVoice removes a voice entry from the registry.
func (r *FileVoiceRegistry) DeleteVoice(npcID string) error {
	r.mu.Lock()
	delete(r.voices, npcID)
	r.mu.Unlock()

	if err := r.save(); err != nil {
		return fmt.Errorf("failed to save voice registry after delete: %w", err)
	}
	return nil
}
