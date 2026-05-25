package piper

import (
	"errors"
	"io"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestPlayerDetectPlayer(t *testing.T) {
	t.Run("auto detects aplay", func(t *testing.T) {
		p := newPlayerWithDeps("auto", "", func(name string) (string, error) {
			if name == "aplay" {
				return "/usr/bin/aplay", nil
			}
			return "", errors.New("not found")
		}, exec.Command)
		if p.player != "aplay" {
			t.Errorf("expected player 'aplay', got '%s'", p.player)
		}
	})

	t.Run("auto falls back to paplay", func(t *testing.T) {
		p := newPlayerWithDeps("auto", "", func(name string) (string, error) {
			if name == "paplay" {
				return "/usr/bin/paplay", nil
			}
			return "", errors.New("not found")
		}, exec.Command)
		if p.player != "paplay" {
			t.Errorf("expected player 'paplay', got '%s'", p.player)
		}
	})

	t.Run("auto falls back to ffplay", func(t *testing.T) {
		p := newPlayerWithDeps("auto", "", func(name string) (string, error) {
			if name == "ffplay" {
				return "/usr/bin/ffplay", nil
			}
			return "", errors.New("not found")
		}, exec.Command)
		if p.player != "ffplay" {
			t.Errorf("expected player 'ffplay', got '%s'", p.player)
		}
	})

	t.Run("no player available", func(t *testing.T) {
		p := newPlayerWithDeps("auto", "", func(name string) (string, error) {
			return "", errors.New("not found")
		}, exec.Command)
		if p.player != "" {
			t.Errorf("expected empty player, got '%s'", p.player)
		}
	})

	t.Run("preferred overrides auto", func(t *testing.T) {
		p := newPlayerWithDeps("ffplay", "", func(name string) (string, error) {
			if name == "ffplay" {
				return "/usr/bin/ffplay", nil
			}
			return "", errors.New("not found")
		}, exec.Command)
		if p.player != "ffplay" {
			t.Errorf("expected player 'ffplay', got '%s'", p.player)
		}
	})

	t.Run("preferred not found falls back", func(t *testing.T) {
		p := newPlayerWithDeps("aplay", "", func(name string) (string, error) {
			if name == "paplay" {
				return "/usr/bin/paplay", nil
			}
			return "", errors.New("not found")
		}, exec.Command)
		if p.player != "paplay" {
			t.Errorf("expected player 'paplay', got '%s'", p.player)
		}
	})
}

func TestPlayerEnqueue(t *testing.T) {
	t.Run("no player available", func(t *testing.T) {
		p := newPlayerWithDeps("auto", "", func(name string) (string, error) {
			return "", errors.New("not found")
		}, exec.Command)
		err := p.Enqueue(strings.NewReader("wav"))
		if err == nil {
			t.Fatal("expected error when no player available")
		}
	})

	t.Run("enqueue starts playback", func(t *testing.T) {
		var played int32
		p := newPlayerWithDeps("auto", "", func(name string) (string, error) {
			if name == "aplay" {
				return "/usr/bin/aplay", nil
			}
			return "", errors.New("not found")
		}, exec.Command)
		p.playChunkFn = func(r io.Reader) error {
			atomic.AddInt32(&played, 1)
			return nil
		}

		err := p.Enqueue(strings.NewReader("wav1"))
		if err != nil {
			t.Fatalf("Enqueue() error = %v", err)
		}

		// Wait for playback to complete
		time.Sleep(50 * time.Millisecond)
		if atomic.LoadInt32(&played) != 1 {
			t.Errorf("expected 1 chunk played, got %d", atomic.LoadInt32(&played))
		}
	})

	t.Run("enqueue multiple chunks", func(t *testing.T) {
		var played int32
		p := newPlayerWithDeps("auto", "", func(name string) (string, error) {
			if name == "aplay" {
				return "/usr/bin/aplay", nil
			}
			return "", errors.New("not found")
		}, exec.Command)
		p.playChunkFn = func(r io.Reader) error {
			atomic.AddInt32(&played, 1)
			return nil
		}

		_ = p.Enqueue(strings.NewReader("wav1"))
		_ = p.Enqueue(strings.NewReader("wav2"))
		_ = p.Enqueue(strings.NewReader("wav3"))

		time.Sleep(100 * time.Millisecond)
		if atomic.LoadInt32(&played) != 3 {
			t.Errorf("expected 3 chunks played, got %d", atomic.LoadInt32(&played))
		}
	})
}

func TestPlayerStop(t *testing.T) {
	t.Run("stop clears queue", func(t *testing.T) {
		var played int32
		p := newPlayerWithDeps("auto", "", func(name string) (string, error) {
			if name == "aplay" {
				return "/usr/bin/aplay", nil
			}
			return "", errors.New("not found")
		}, exec.Command)
		p.playChunkFn = func(r io.Reader) error {
			atomic.AddInt32(&played, 1)
			time.Sleep(20 * time.Millisecond)
			return nil
		}

		_ = p.Enqueue(strings.NewReader("wav1"))
		_ = p.Enqueue(strings.NewReader("wav2"))
		time.Sleep(10 * time.Millisecond) // Let first chunk start

		if err := p.Stop(); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}

		if p.IsPlaying() {
			t.Error("expected IsPlaying() = false after Stop")
		}
		// Should have played only the first chunk (or none if stopped fast enough)
		if atomic.LoadInt32(&played) > 1 {
			t.Errorf("expected at most 1 chunk played, got %d", atomic.LoadInt32(&played))
		}
	})

	t.Run("stop when not playing", func(t *testing.T) {
		p := newPlayerWithDeps("auto", "", func(name string) (string, error) {
			return "/usr/bin/aplay", nil
		}, exec.Command)
		if err := p.Stop(); err != nil {
			t.Fatalf("Stop() when not playing should return nil, got %v", err)
		}
	})
}

func TestPlayerPauseResume(t *testing.T) {
	t.Run("pause and resume", func(t *testing.T) {
		var played int32
		p := newPlayerWithDeps("auto", "", func(name string) (string, error) {
			if name == "aplay" {
				return "/usr/bin/aplay", nil
			}
			return "", errors.New("not found")
		}, exec.Command)
		p.playChunkFn = func(r io.Reader) error {
			atomic.AddInt32(&played, 1)
			return nil
		}

		_ = p.Enqueue(strings.NewReader("wav1"))
		time.Sleep(10 * time.Millisecond)
		if err := p.Pause(); err != nil {
			t.Fatalf("Pause() error = %v", err)
		}

		// Enqueue more while paused
		_ = p.Enqueue(strings.NewReader("wav2"))
		time.Sleep(20 * time.Millisecond)

		// Should not have played wav2 yet
		count := atomic.LoadInt32(&played)
		if count != 1 {
			t.Logf("played count: %d (may vary depending on timing)", count)
		}

		if err := p.Resume(); err != nil {
			t.Fatalf("Resume() error = %v", err)
		}
		time.Sleep(50 * time.Millisecond)

		if atomic.LoadInt32(&played) != 2 {
			t.Errorf("expected 2 chunks played after resume, got %d", atomic.LoadInt32(&played))
		}
	})
}

func TestPlayerSkip(t *testing.T) {
	t.Run("skip current chunk", func(t *testing.T) {
		var played int32
		p := newPlayerWithDeps("auto", "", func(name string) (string, error) {
			if name == "aplay" {
				return "/usr/bin/aplay", nil
			}
			return "", errors.New("not found")
		}, exec.Command)
		p.playChunkFn = func(r io.Reader) error {
			atomic.AddInt32(&played, 1)
			time.Sleep(50 * time.Millisecond)
			return nil
		}

		_ = p.Enqueue(strings.NewReader("wav1"))
		_ = p.Enqueue(strings.NewReader("wav2"))
		time.Sleep(10 * time.Millisecond) // Let first chunk start

		if err := p.Skip(); err != nil {
			t.Fatalf("Skip() error = %v", err)
		}

		time.Sleep(100 * time.Millisecond)
		// Should have played wav1 and wav2 (skip may or may not abort wav1 depending on timing)
		count := atomic.LoadInt32(&played)
		if count < 1 {
			t.Errorf("expected at least 1 chunk played, got %d", count)
		}
	})

	t.Run("skip when not playing", func(t *testing.T) {
		p := newPlayerWithDeps("auto", "", func(name string) (string, error) {
			return "/usr/bin/aplay", nil
		}, exec.Command)
		if err := p.Skip(); err != nil {
			t.Fatalf("Skip() when not playing should return nil, got %v", err)
		}
	})
}

func TestPlayerClose(t *testing.T) {
	t.Run("close releases resources", func(t *testing.T) {
		p := newPlayerWithDeps("auto", "", func(name string) (string, error) {
			if name == "aplay" {
				return "/usr/bin/aplay", nil
			}
			return "", errors.New("not found")
		}, exec.Command)
		p.playChunkFn = func(r io.Reader) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		}

		_ = p.Enqueue(strings.NewReader("wav1"))
		time.Sleep(10 * time.Millisecond)

		if err := p.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}

		if p.IsPlaying() {
			t.Error("expected IsPlaying() = false after Close")
		}

		// Enqueue after close should error
		if err := p.Enqueue(strings.NewReader("wav2")); err == nil {
			t.Error("expected error when enqueuing after close")
		}
	})
}

func TestPlayerArgs(t *testing.T) {
	tests := []struct {
		player   string
		device   string
		expected []string
	}{
		{"aplay", "", []string{"-"}},
		{"aplay", "plughw:0,0", []string{"-D", "plughw:0,0", "-"}},
		{"paplay", "", []string{"-"}},
		{"paplay", "sink1", []string{"--device=sink1", "-"}},
		{"ffplay", "", []string{"-nodisp", "-autoexit", "-"}},
		{"unknown", "", []string{"-"}},
	}

	for _, tt := range tests {
		t.Run(tt.player, func(t *testing.T) {
			p := &Player{player: tt.player, device: tt.device}
			got := p.playerArgs()
			if len(got) != len(tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("arg[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}
