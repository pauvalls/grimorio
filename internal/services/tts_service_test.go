package services

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/dm"
	"github.com/pauvalls/grimorio/internal/tts/piper"
)

// ---- Mocks ----

type mockPipeline struct {
	mu       sync.Mutex
	narrated []string
	err      error
	running  bool
}

func (m *mockPipeline) Narrate(ctx context.Context, text string) error {
	m.mu.Lock()
	m.narrated = append(m.narrated, text)
	m.running = true
	m.mu.Unlock()
	return m.err
}

func (m *mockPipeline) Stop() error {
	m.mu.Lock()
	m.running = false
	m.mu.Unlock()
	return nil
}
func (m *mockPipeline) Skip() error     { return nil }
func (m *mockPipeline) Pause() error    { return nil }
func (m *mockPipeline) Resume() error   { return nil }
func (m *mockPipeline) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

type mockLifecycle struct {
	installed bool
	running   bool
}

func (m *mockLifecycle) IsInstalled() bool { return m.installed }
func (m *mockLifecycle) IsRunning() bool   { return m.running }
func (m *mockLifecycle) Start(ctx context.Context) error {
	if !m.installed {
		return errors.New("not installed")
	}
	m.running = true
	return nil
}
func (m *mockLifecycle) Stop(ctx context.Context) error {
	m.running = false
	return nil
}
func (m *mockLifecycle) Restart(ctx context.Context) error { return nil }

// ---- Tests ----

func TestTTSServiceDeliverResponse(t *testing.T) {
	pipe := &mockPipeline{}
	life := &mockLifecycle{installed: true, running: true}
	svc := NewTTSService(pipe, life, nil, true)

	t.Run("delivers in tts mode", func(t *testing.T) {
		if err := svc.SetMode(dm.ModeTTS); err != nil {
			t.Fatalf("SetMode() error = %v", err)
		}
		if err := svc.DeliverResponse("Hello world"); err != nil {
			t.Fatalf("DeliverResponse() error = %v", err)
		}

		pipe.mu.Lock()
		if len(pipe.narrated) != 1 || pipe.narrated[0] != "Hello world" {
			t.Errorf("expected narrated ['Hello world'], got %v", pipe.narrated)
		}
		pipe.mu.Unlock()
	})

	t.Run("ignores in written mode", func(t *testing.T) {
		pipe.mu.Lock()
		pipe.narrated = nil
		pipe.mu.Unlock()

		if err := svc.SetMode(dm.ModeWritten); err != nil {
			t.Fatalf("SetMode() error = %v", err)
		}
		if err := svc.DeliverResponse("Hello world"); err != nil {
			t.Fatalf("DeliverResponse() error = %v", err)
		}

		pipe.mu.Lock()
		if len(pipe.narrated) != 0 {
			t.Errorf("expected no narration in written mode, got %v", pipe.narrated)
		}
		pipe.mu.Unlock()
	})

	t.Run("ignores when pipeline is nil", func(t *testing.T) {
		svc2 := NewTTSService(nil, life, nil, true)
		if err := svc2.SetMode(dm.ModeTTS); err != nil {
			t.Fatalf("SetMode() error = %v", err)
		}
		if err := svc2.DeliverResponse("Hello world"); err != nil {
			t.Fatalf("DeliverResponse() error = %v", err)
		}
	})
}

func TestTTSServiceSetMode(t *testing.T) {
	svc := NewTTSService(&mockPipeline{}, &mockLifecycle{}, nil, false)

	if svc.GetMode() != dm.ModeWritten {
		t.Errorf("expected default mode written, got %q", svc.GetMode())
	}

	if err := svc.SetMode(dm.ModeTTS); err == nil {
		t.Error("expected SetMode(TTS) to fail when not enabled/available")
	}

	if err := svc.SetMode(dm.ModeTTS); err == nil {
		t.Error("expected SetMode(TTS) to fail when not enabled/available")
	}

	if err := svc.SetMode(dm.ModeWritten); err != nil {
		t.Fatalf("SetMode(Written) error = %v", err)
	}
	if svc.GetMode() != dm.ModeWritten {
		t.Errorf("expected mode written, got %q", svc.GetMode())
	}
}

func TestTTSServiceIsAvailable(t *testing.T) {
	t.Run("available when installed and running", func(t *testing.T) {
		svc := NewTTSService(&mockPipeline{}, &mockLifecycle{installed: true, running: true}, nil, true)
		if !svc.IsAvailable() {
			t.Error("expected IsAvailable() = true")
		}
	})

	t.Run("not available when not installed", func(t *testing.T) {
		svc := NewTTSService(&mockPipeline{}, &mockLifecycle{installed: false, running: false}, nil, true)
		if svc.IsAvailable() {
			t.Error("expected IsAvailable() = false")
		}
	})

	t.Run("not available when nil lifecycle", func(t *testing.T) {
		svc := NewTTSService(&mockPipeline{}, nil, nil, true)
		if svc.IsAvailable() {
			t.Error("expected IsAvailable() = false")
		}
	})
}

func TestTTSServiceStart(t *testing.T) {
	t.Run("starts when installed", func(t *testing.T) {
		life := &mockLifecycle{installed: true}
		svc := NewTTSService(&mockPipeline{}, life, nil, true)
		if err := svc.Start(context.Background()); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		if !life.running {
			t.Error("expected lifecycle running after Start")
		}
	})

	t.Run("fails when not installed", func(t *testing.T) {
		life := &mockLifecycle{installed: false}
		svc := NewTTSService(&mockPipeline{}, life, nil, true)
		if err := svc.Start(context.Background()); err == nil {
			t.Fatal("expected error when not installed")
		}
	})

	t.Run("noop when nil lifecycle", func(t *testing.T) {
		svc := NewTTSService(&mockPipeline{}, nil, nil, true)
		if err := svc.Start(context.Background()); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
	})
}

func TestTTSServiceShutdown(t *testing.T) {
	life := &mockLifecycle{installed: true, running: true}
	svc := NewTTSService(&mockPipeline{}, life, nil, true)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := svc.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	if life.running {
		t.Error("expected lifecycle not running after Shutdown")
	}
}

func TestTTSServiceControl(t *testing.T) {
	pipe := &mockPipeline{}
	svc := NewTTSService(pipe, &mockLifecycle{installed: true, running: true}, nil, true)

	for _, action := range []string{"skip", "stop", "pause", "resume"} {
		if err := svc.Control(action); err != nil {
			t.Errorf("Control(%q) error = %v", action, err)
		}
	}

	if err := svc.Control("invalid"); err == nil {
		t.Error("expected error for unknown action")
	}

	// nil pipeline
	svc2 := NewTTSService(nil, nil, nil, true)
	if err := svc2.Control("stop"); err == nil {
		t.Error("expected error when pipeline is nil")
	}
}

func TestTTSServiceAssignNPCVoice(t *testing.T) {
	store := NewFileCampaignVoiceStore(t.TempDir())
	svc := NewTTSService(&mockPipeline{}, &mockLifecycle{installed: true, running: true}, store, true)

	voiceID, created, err := svc.AssignNPCVoice("campaign-1", "Goblin Chief", "gruff and low")
	if err != nil {
		t.Fatalf("AssignNPCVoice() error = %v", err)
	}
	if !created {
		t.Error("expected created = true for new voice")
	}
	if voiceID != "npc-goblin-chief" {
		t.Errorf("expected voiceID npc-goblin-chief, got %q", voiceID)
	}

	// Update existing
	_, created2, err := svc.AssignNPCVoice("campaign-1", "Goblin Chief", "gruff and low updated")
	if err != nil {
		t.Fatalf("AssignNPCVoice() error = %v", err)
	}
	if created2 {
		t.Error("expected created = false for existing voice")
	}

	// nil store
	svc2 := NewTTSService(&mockPipeline{}, nil, nil, true)
	if _, _, err := svc2.AssignNPCVoice("c", "n", "v"); err == nil {
		t.Error("expected error when store is nil")
	}
}

func TestTTSServiceListVoices(t *testing.T) {
	store := NewFileCampaignVoiceStore(t.TempDir())
	svc := NewTTSService(&mockPipeline{}, nil, store, true)

	if err := store.SetVoicePrompt("campaign-1", "NPC1", "voice1"); err != nil {
		t.Fatalf("SetVoicePrompt() error = %v", err)
	}

	voices := svc.ListVoices("campaign-1")
	if len(voices) != 1 {
		t.Errorf("expected 1 voice, got %d", len(voices))
	}

	// nil store
	svc2 := NewTTSService(&mockPipeline{}, nil, nil, true)
	if v := svc2.ListVoices("campaign-1"); len(v) != 0 {
		t.Errorf("expected 0 voices with nil store, got %d", len(v))
	}
}

func TestTTSServiceGetStatus(t *testing.T) {
	pipe := &mockPipeline{running: true}
	svc := NewTTSService(pipe, &mockLifecycle{installed: true, running: true}, nil, true)
	if err := svc.SetMode(dm.ModeTTS); err != nil {
		t.Fatalf("SetMode() error = %v", err)
	}

	status := svc.GetStatus()
	if !status.Enabled {
		t.Error("expected Enabled = true")
	}
	if status.Mode != "tts" {
		t.Errorf("expected Mode = tts, got %q", status.Mode)
	}
	if !status.Available {
		t.Error("expected Available = true")
	}
	if !status.Playing {
		t.Error("expected Playing = true")
	}
}

func TestTTSServiceIntegrationWithRealNarrator(t *testing.T) {
	// End-to-end: TTSService → Narrator → Filter → Chunk → MockEngine → MockPlayer
	filter := &piper.DefaultTextFilter{}
	engine := &mockTTSEngine{}
	player := &mockTTSPlayer{}

	narrator := piper.NewNarrator(filter, engine, player, 20)
	life := &mockLifecycle{installed: true, running: true}
	svc := NewTTSService(narrator, life, nil, true)
	if err := svc.SetMode(dm.ModeTTS); err != nil {
		t.Fatalf("SetMode() error = %v", err)
	}

	text := `El dragón ataca.
| Stat | Valor |
| HP   | 250   |
El héroe huye.`

	if err := svc.DeliverResponse(text); err != nil {
		t.Fatalf("DeliverResponse() error = %v", err)
	}

	// Player should have received filtered chunks, no table lines
	player.mu.Lock()
	for _, enq := range player.enqueued {
		if strings.Contains(enq, "|") {
			t.Errorf("player should not receive table lines, got: %q", enq)
		}
	}
	if len(player.enqueued) != 2 {
		t.Errorf("expected 2 enqueued chunks, got %d: %v", len(player.enqueued), player.enqueued)
	}
	player.mu.Unlock()
}

// ---- Integration mocks (satisfy piper interfaces) ----

type mockTTSEngine struct {
	mu    sync.Mutex
	calls []string
}

func (m *mockTTSEngine) Synthesize(ctx context.Context, text string) (io.ReadCloser, error) {
	m.mu.Lock()
	m.calls = append(m.calls, text)
	m.mu.Unlock()
	return io.NopCloser(strings.NewReader(text)), nil
}
func (m *mockTTSEngine) HealthCheck(ctx context.Context) error { return nil }
func (m *mockTTSEngine) Close() error                           { return nil }

type mockTTSPlayer struct {
	mu       sync.Mutex
	enqueued []string
}

func (m *mockTTSPlayer) Enqueue(wav io.Reader) error {
	data, _ := io.ReadAll(wav)
	m.mu.Lock()
	m.enqueued = append(m.enqueued, string(data))
	m.mu.Unlock()
	return nil
}
func (m *mockTTSPlayer) Skip() error     { return nil }
func (m *mockTTSPlayer) Stop() error     { return nil }
func (m *mockTTSPlayer) Pause() error    { return nil }
func (m *mockTTSPlayer) Resume() error   { return nil }
func (m *mockTTSPlayer) IsPlaying() bool { return false }
func (m *mockTTSPlayer) Close() error    { return nil }
