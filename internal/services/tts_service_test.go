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
}

func (m *mockPipeline) Narrate(ctx context.Context, text string) error {
	m.mu.Lock()
	m.narrated = append(m.narrated, text)
	m.mu.Unlock()
	return m.err
}

func (m *mockPipeline) Stop() error { return nil }

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
	svc := NewTTSService(pipe, life)

	t.Run("delivers in tts mode", func(t *testing.T) {
		svc.SetMode(dm.ModeTTS)
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

		svc.SetMode(dm.ModeWritten)
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
		svc2 := NewTTSService(nil, life)
		svc2.SetMode(dm.ModeTTS)
		if err := svc2.DeliverResponse("Hello world"); err != nil {
			t.Fatalf("DeliverResponse() error = %v", err)
		}
	})
}

func TestTTSServiceSetMode(t *testing.T) {
	svc := NewTTSService(&mockPipeline{}, &mockLifecycle{})

	if svc.GetMode() != dm.ModeWritten {
		t.Errorf("expected default mode written, got %q", svc.GetMode())
	}

	svc.SetMode(dm.ModeTTS)
	if svc.GetMode() != dm.ModeTTS {
		t.Errorf("expected mode tts, got %q", svc.GetMode())
	}
}

func TestTTSServiceIsAvailable(t *testing.T) {
	t.Run("available when installed and running", func(t *testing.T) {
		svc := NewTTSService(&mockPipeline{}, &mockLifecycle{installed: true, running: true})
		if !svc.IsAvailable() {
			t.Error("expected IsAvailable() = true")
		}
	})

	t.Run("not available when not installed", func(t *testing.T) {
		svc := NewTTSService(&mockPipeline{}, &mockLifecycle{installed: false, running: false})
		if svc.IsAvailable() {
			t.Error("expected IsAvailable() = false")
		}
	})

	t.Run("not available when nil lifecycle", func(t *testing.T) {
		svc := NewTTSService(&mockPipeline{}, nil)
		if svc.IsAvailable() {
			t.Error("expected IsAvailable() = false")
		}
	})
}

func TestTTSServiceStart(t *testing.T) {
	t.Run("starts when installed", func(t *testing.T) {
		life := &mockLifecycle{installed: true}
		svc := NewTTSService(&mockPipeline{}, life)
		if err := svc.Start(context.Background()); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		if !life.running {
			t.Error("expected lifecycle running after Start")
		}
	})

	t.Run("fails when not installed", func(t *testing.T) {
		life := &mockLifecycle{installed: false}
		svc := NewTTSService(&mockPipeline{}, life)
		if err := svc.Start(context.Background()); err == nil {
			t.Fatal("expected error when not installed")
		}
	})

	t.Run("noop when nil lifecycle", func(t *testing.T) {
		svc := NewTTSService(&mockPipeline{}, nil)
		if err := svc.Start(context.Background()); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
	})
}

func TestTTSServiceShutdown(t *testing.T) {
	life := &mockLifecycle{installed: true, running: true}
	svc := NewTTSService(&mockPipeline{}, life)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := svc.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	if life.running {
		t.Error("expected lifecycle not running after Shutdown")
	}
}

func TestTTSServiceIntegrationWithRealNarrator(t *testing.T) {
	// End-to-end: TTSService → Narrator → Filter → Chunk → MockEngine → MockPlayer
	filter := &piper.DefaultTextFilter{}
	engine := &mockTTSEngine{}
	player := &mockTTSPlayer{}

	narrator := piper.NewNarrator(filter, engine, player, 20)
	life := &mockLifecycle{installed: true, running: true}
	svc := NewTTSService(narrator, life)
	svc.SetMode(dm.ModeTTS)

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
