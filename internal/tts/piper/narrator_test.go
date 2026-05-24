package piper

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ---- Mocks ----

type mockFilter struct {
	result string
}

func (m *mockFilter) Filter(text string) string {
	if m.result != "" {
		return m.result
	}
	return text
}

type mockEngine struct {
	mu        sync.Mutex
	calls     []string
	delay     time.Duration
	errOnText map[string]error
	blockOn   map[string]chan struct{}
}

func (m *mockEngine) Synthesize(ctx context.Context, text string) (io.ReadCloser, error) {
	m.mu.Lock()
	m.calls = append(m.calls, text)
	ch := m.blockOn[text]
	m.mu.Unlock()

	if ch != nil {
		select {
		case <-ch:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	m.mu.Lock()
	err := m.errOnText[text]
	m.mu.Unlock()
	if err != nil {
		return nil, err
	}

	return io.NopCloser(strings.NewReader(text)), nil
}

func (m *mockEngine) HealthCheck(ctx context.Context) error { return nil }
func (m *mockEngine) Close() error                          { return nil }

type mockPlayer struct {
	mu           sync.Mutex
	enqueued     []string
	delay        time.Duration
	enqueueErr   error
	blockOnFirst chan struct{}
	firstOnce    int32
}

func (m *mockPlayer) Enqueue(wav io.Reader) error {
	if m.enqueueErr != nil {
		return m.enqueueErr
	}
	data, _ := io.ReadAll(wav)
	txt := string(data)

	m.mu.Lock()
	m.enqueued = append(m.enqueued, txt)
	m.mu.Unlock()

	if m.blockOnFirst != nil && atomic.CompareAndSwapInt32(&m.firstOnce, 0, 1) {
		<-m.blockOnFirst
	}

	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	return nil
}

func (m *mockPlayer) Skip() error       { return nil }
func (m *mockPlayer) Stop() error       { return nil }
func (m *mockPlayer) Pause() error      { return nil }
func (m *mockPlayer) Resume() error     { return nil }
func (m *mockPlayer) IsPlaying() bool   { return false }
func (m *mockPlayer) Close() error      { return nil }

func (m *mockPlayer) enqueuedCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.enqueued)
}

// ---- Tests ----

func TestNarratorPipelineComplete(t *testing.T) {
	filter := &mockFilter{}
	engine := &mockEngine{}
	player := &mockPlayer{}

	// Use small maxLen to force multiple chunks
	n := NewNarrator(filter, engine, player, 20)
	ctx := context.Background()

	text := "First sentence. Second sentence."
	if err := n.Narrate(ctx, text); err != nil {
		t.Fatalf("Narrate() error = %v", err)
	}

	engine.mu.Lock()
	if len(engine.calls) != 2 {
		t.Errorf("expected 2 synthesize calls, got %d: %v", len(engine.calls), engine.calls)
	}
	engine.mu.Unlock()

	if player.enqueuedCount() != 2 {
		t.Errorf("expected 2 enqueued chunks, got %d", player.enqueuedCount())
	}
}

func TestNarratorPreload(t *testing.T) {
	filter := &mockFilter{}
	engine := &mockEngine{}

	// Player blocks on first chunk until test signals
	blockCh := make(chan struct{})
	player := &mockPlayer{blockOnFirst: blockCh}

	n := NewNarrator(filter, engine, player, 20)

	// Two chunks guaranteed with this text at maxLen 20
	text := "First chunk here. Second chunk here."

	done := make(chan error, 1)
	go func() {
		done <- n.Narrate(context.Background(), text)
	}()

	// Give time for first chunk to reach player and block
	time.Sleep(50 * time.Millisecond)

	// The synthesizer should have started on chunk 2 (preload) even though
	// chunk 1 is blocked in playback.
	engine.mu.Lock()
	callCount := len(engine.calls)
	engine.mu.Unlock()

	if callCount < 2 {
		t.Errorf("expected preload: engine should have received 2 chunks, got %d", callCount)
	}

	// Unblock player
	close(blockCh)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Narrate() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for Narrate")
	}

	if player.enqueuedCount() != 2 {
		t.Errorf("expected 2 enqueued chunks, got %d", player.enqueuedCount())
	}
}

func TestNarratorCancellation(t *testing.T) {
	filter := &mockFilter{}

	// Engine blocks forever on first call
	blockCh := make(chan struct{})
	engine := &mockEngine{
		blockOn: map[string]chan struct{}{
			"First chunk here.": blockCh,
		},
	}
	player := &mockPlayer{}

	n := NewNarrator(filter, engine, player, 20)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- n.Narrate(ctx, "First chunk here. Second chunk here.")
	}()

	// Let it start
	time.Sleep(30 * time.Millisecond)

	// Cancel
	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for cancellation")
	}

	// Should not be running
	if n.IsRunning() {
		t.Error("expected IsRunning() = false after cancellation")
	}
}

func TestNarratorSkipFailedChunk(t *testing.T) {
	filter := &mockFilter{}
	engine := &mockEngine{
		errOnText: map[string]error{
			"bad chunk.": errors.New("synthesis failed"),
		},
	}
	player := &mockPlayer{}

	n := NewNarrator(filter, engine, player, 15)
	ctx := context.Background()

	// First chunk will fail, second should succeed
	text := "bad chunk. good chunk."
	if err := n.Narrate(ctx, text); err != nil {
		t.Fatalf("Narrate() error = %v", err)
	}

	// Engine should have attempted both
	engine.mu.Lock()
	if len(engine.calls) != 2 {
		t.Errorf("expected 2 synthesize attempts, got %d", len(engine.calls))
	}
	engine.mu.Unlock()

	// Player should only have received the good chunk
	if player.enqueuedCount() != 1 {
		t.Errorf("expected 1 enqueued chunk (skipped failed one), got %d", player.enqueuedCount())
	}

	player.mu.Lock()
	if len(player.enqueued) > 0 && player.enqueued[0] != "good chunk." {
		t.Errorf("expected enqueued 'good chunk.', got %q", player.enqueued[0])
	}
	player.mu.Unlock()
}

func TestNarratorStop(t *testing.T) {
	filter := &mockFilter{}
	engine := &mockEngine{delay: 100 * time.Millisecond}
	player := &mockPlayer{delay: 50 * time.Millisecond}

	n := NewNarrator(filter, engine, player, 10)

	done := make(chan error, 1)
	go func() {
		done <- n.Narrate(context.Background(), "One. Two. Three.")
	}()

	time.Sleep(30 * time.Millisecond)

	if err := n.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for Stop")
	}

	if n.IsRunning() {
		t.Error("expected IsRunning() = false after Stop")
	}
}

func TestNarratorEmptyText(t *testing.T) {
	filter := &mockFilter{result: ""}
	engine := &mockEngine{}
	player := &mockPlayer{}

	n := NewNarrator(filter, engine, player, 150)

	if err := n.Narrate(context.Background(), "   "); err != nil {
		t.Fatalf("Narrate() error = %v", err)
	}

	if player.enqueuedCount() != 0 {
		t.Errorf("expected 0 chunks for empty text, got %d", player.enqueuedCount())
	}
}

func TestNarratorFilterRemovesTables(t *testing.T) {
	filter := &DefaultTextFilter{}
	engine := &mockEngine{}
	player := &mockPlayer{}

	n := NewNarrator(filter, engine, player, 150)

	text := `El dragón ataca.
| Stat | Valor |
|------|-------|
| HP   | 250   |
El héroe huye.`

	if err := n.Narrate(context.Background(), text); err != nil {
		t.Fatalf("Narrate() error = %v", err)
	}

	engine.mu.Lock()
	for _, call := range engine.calls {
		if strings.Contains(call, "|") {
			t.Errorf("engine should not receive table lines, got: %q", call)
		}
	}
	engine.mu.Unlock()
}
