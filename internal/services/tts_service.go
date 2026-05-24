package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pauvalls/grimorio/internal/dm"
	"github.com/pauvalls/grimorio/internal/tts/piper"
)

// TTSService coordinates TTS narration for DM responses.
type TTSService struct {
	pipeline  piper.NarratorPipeline
	lifecycle piper.LifecycleManager
	mode      dm.DMMode
	mu        sync.RWMutex
}

// NewTTSService creates a new TTSService with the given pipeline and lifecycle.
func NewTTSService(pipeline piper.NarratorPipeline, lifecycle piper.LifecycleManager) *TTSService {
	return &TTSService{
		pipeline:  pipeline,
		lifecycle: lifecycle,
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

// SetMode changes the output mode between written and TTS.
func (s *TTSService) SetMode(mode dm.DMMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mode = mode
}

// GetMode returns the current DM mode.
func (s *TTSService) GetMode() dm.DMMode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.mode
}

// IsAvailable returns true if Piper is installed and running.
func (s *TTSService) IsAvailable() bool {
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
