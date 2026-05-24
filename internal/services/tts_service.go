package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pauvalls/grimorio/internal/dm"
	"github.com/pauvalls/grimorio/internal/tts/piper"
)

// TTSService coordinates TTS narration for DM responses.
type TTSService struct {
	pipeline  piper.NarratorPipeline
	lifecycle piper.LifecycleManager
	store     CampaignVoiceStore
	enabled   bool
	mode      dm.DMMode
	mu        sync.RWMutex
}

// NewTTSService creates a new TTSService with the given dependencies.
func NewTTSService(pipeline piper.NarratorPipeline, lifecycle piper.LifecycleManager, store CampaignVoiceStore, enabled bool) *TTSService {
	return &TTSService{
		pipeline:  pipeline,
		lifecycle: lifecycle,
		store:     store,
		enabled:   enabled,
		mode:      dm.ModeWritten,
	}
}

// DeliverResponse sends DM text to the TTS pipeline if in TTS mode.
func (s *TTSService) DeliverResponse(text string) error {
	s.mu.RLock()
	mode := s.mode
	pipeline := s.pipeline
	s.mu.RUnlock()

	if mode != dm.ModeTTS || pipeline == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	return pipeline.Narrate(ctx, text)
}

// SetMode switches between written and TTS modes.
func (s *TTSService) SetMode(mode dm.DMMode) error {
	if mode != dm.ModeWritten && mode != dm.ModeTTS {
		return fmt.Errorf("invalid mode: %s", mode)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.mode == mode {
		return nil
	}
	if mode == dm.ModeTTS && (!s.enabled || !s.isAvailableUnlocked()) {
		return fmt.Errorf("TTS is not enabled or not available")
	}
	s.mode = mode
	return nil
}

// GetMode returns the current DM mode.
func (s *TTSService) GetMode() dm.DMMode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mode
}

// IsAvailable returns true if Piper is installed and running.
func (s *TTSService) IsAvailable() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isAvailableUnlocked()
}

func (s *TTSService) isAvailableUnlocked() bool {
	if s.lifecycle == nil {
		return false
	}
	return s.lifecycle.IsInstalled() && s.lifecycle.IsRunning()
}

// Start initializes the TTS lifecycle (starts Piper if available).
func (s *TTSService) Start(ctx context.Context) error {
	if s.lifecycle == nil {
		return nil
	}
	if !s.lifecycle.IsInstalled() {
		return fmt.Errorf("piper not installed")
	}
	return s.lifecycle.Start(ctx)
}

// Shutdown gracefully stops the TTS service.
func (s *TTSService) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	pipeline := s.pipeline
	s.mu.Unlock()

	if pipeline != nil {
		_ = pipeline.Stop()
	}

	if s.lifecycle != nil {
		return s.lifecycle.Stop(ctx)
	}
	return nil
}

// Control sends a playback control command.
func (s *TTSService) Control(action string) error {
	s.mu.RLock()
	pipeline := s.pipeline
	s.mu.RUnlock()

	if pipeline == nil {
		return fmt.Errorf("TTS pipeline not available")
	}
	switch strings.ToLower(action) {
	case "skip":
		return pipeline.Skip()
	case "stop":
		return pipeline.Stop()
	case "pause":
		return pipeline.Pause()
	case "resume":
		return pipeline.Resume()
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

// AssignNPCVoice assigns a voice prompt to an NPC.
func (s *TTSService) AssignNPCVoice(campaignID, npcName, voicePrompt string) (string, bool, error) {
	if s.store == nil {
		return "", false, fmt.Errorf("voice store not available")
	}
	voiceID := "npc-" + dm.Slugify(npcName)
	_, err := s.store.GetVoiceID(campaignID, npcName)
	exists := err == nil
	if err := s.store.SetVoicePrompt(campaignID, npcName, voicePrompt); err != nil {
		return "", false, fmt.Errorf("failed to set voice: %w", err)
	}
	return voiceID, !exists, nil
}

// ListVoices lists assigned NPC voices for a campaign.
func (s *TTSService) ListVoices(campaignID string) map[string]string {
	if s.store == nil {
		return map[string]string{}
	}
	return s.store.ListVoices(campaignID)
}

// GetStatus returns the current TTS status.
func (s *TTSService) GetStatus() dm.TTSStatus {
	s.mu.RLock()
	mode := s.mode
	pipeline := s.pipeline
	s.mu.RUnlock()

	status := dm.TTSStatus{
		Enabled:   s.enabled,
		Mode:      string(mode),
		Available: s.IsAvailable(),
	}
	if pipeline != nil {
		status.Playing = pipeline.IsRunning()
	}
	return status
}
