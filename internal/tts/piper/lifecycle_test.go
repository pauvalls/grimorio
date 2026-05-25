package piper

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"
)

// mockProcess implements Process for testing.
type mockProcess struct {
	mu        sync.Mutex
	signals   []os.Signal
	waitState *os.ProcessState
	waitErr   error
}

func (m *mockProcess) Pid() int { return 12345 }

func (m *mockProcess) Signal(sig os.Signal) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signals = append(m.signals, sig)
	return nil
}

func (m *mockProcess) Wait() (*os.ProcessState, error) {
	return m.waitState, m.waitErr
}

func (m *mockProcess) Signals() []os.Signal {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]os.Signal(nil), m.signals...)
}

func TestLifecycleManagerIsInstalled(t *testing.T) {
	t.Run("installed", func(t *testing.T) {
		lm := NewLifecycleManager(DefaultConfig())
		lm.lookPath = func(string) (string, error) {
			return "/usr/bin/piper", nil
		}
		if !lm.IsInstalled() {
			t.Error("expected IsInstalled() = true")
		}
	})

	t.Run("not installed", func(t *testing.T) {
		lm := NewLifecycleManager(DefaultConfig())
		lm.lookPath = func(string) (string, error) {
			return "", errors.New("not found")
		}
		if lm.IsInstalled() {
			t.Error("expected IsInstalled() = false")
		}
	})
}

func TestLifecycleManagerStart(t *testing.T) {
	t.Run("not installed", func(t *testing.T) {
		lm := NewLifecycleManager(DefaultConfig())
		lm.lookPath = func(string) (string, error) {
			return "", errors.New("not found")
		}
		ctx := context.Background()
		err := lm.Start(ctx)
		if err == nil {
			t.Fatal("expected error when piper not installed")
		}
	})

	t.Run("success", func(t *testing.T) {
		lm := NewLifecycleManager(DefaultConfig())
		lm.lookPath = func(string) (string, error) {
			return "/usr/bin/piper", nil
		}
		lm.startCmd = func(name string, arg ...string) *exec.Cmd {
			// Return a command that does nothing (sleep briefly)
			return exec.Command("sleep", "10")
		}
		lm.newProcess = func(cmd *exec.Cmd) (Process, error) {
			return &mockProcess{}, nil
		}

		ctx := context.Background()
		if err := lm.Start(ctx); err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		if !lm.IsRunning() {
			t.Error("expected IsRunning() = true after Start")
		}

		// Clean up
		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = lm.Stop(stopCtx)
	})

	t.Run("already running", func(t *testing.T) {
		lm := NewLifecycleManager(DefaultConfig())
		lm.lookPath = func(string) (string, error) {
			return "/usr/bin/piper", nil
		}
		lm.startCmd = func(name string, arg ...string) *exec.Cmd {
			return exec.Command("sleep", "10")
		}
		lm.newProcess = func(cmd *exec.Cmd) (Process, error) {
			return &mockProcess{}, nil
		}

		ctx := context.Background()
		_ = lm.Start(ctx)
		err := lm.Start(ctx)
		if err != nil {
			t.Fatalf("Start() when already running should return nil, got %v", err)
		}

		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = lm.Stop(stopCtx)
	})
}

func TestLifecycleManagerStop(t *testing.T) {
	t.Run("graceful shutdown", func(t *testing.T) {
		mock := &mockProcess{}
		lm := NewLifecycleManager(DefaultConfig())
		lm.lookPath = func(string) (string, error) {
			return "/usr/bin/piper", nil
		}
		lm.startCmd = func(name string, arg ...string) *exec.Cmd {
			return exec.Command("sleep", "10")
		}
		lm.newProcess = func(cmd *exec.Cmd) (Process, error) {
			return mock, nil
		}

		ctx := context.Background()
		_ = lm.Start(ctx)

		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := lm.Stop(stopCtx); err != nil {
			t.Fatalf("Stop() error = %v", err)
		}

		if lm.IsRunning() {
			t.Error("expected IsRunning() = false after Stop")
		}

		signals := mock.Signals()
		if len(signals) == 0 {
			t.Fatal("expected at least one signal sent")
		}
		if signals[0] != syscall.SIGTERM {
			t.Errorf("expected first signal SIGTERM, got %v", signals[0])
		}
	})

	t.Run("not running", func(t *testing.T) {
		lm := NewLifecycleManager(DefaultConfig())
		ctx := context.Background()
		if err := lm.Stop(ctx); err != nil {
			t.Fatalf("Stop() when not running should return nil, got %v", err)
		}
	})
}

func TestLifecycleManagerRestart(t *testing.T) {
	t.Run("success after failure", func(t *testing.T) {
		lm := NewLifecycleManager(DefaultConfig())
		lm.config.MaxRestarts = 2
		lm.lookPath = func(string) (string, error) {
			return "/usr/bin/piper", nil
		}

		failCount := 0
		lm.startCmd = func(name string, arg ...string) *exec.Cmd {
			return exec.Command("sleep", "10")
		}
		lm.newProcess = func(cmd *exec.Cmd) (Process, error) {
			failCount++
			if failCount == 1 {
				return nil, errors.New("start failed")
			}
			return &mockProcess{}, nil
		}

		ctx := context.Background()
		// Force backoff to be fast for tests
		lm.config.MaxRestarts = 3

		if err := lm.Restart(ctx); err != nil {
			t.Fatalf("Restart() error = %v", err)
		}

		if !lm.IsRunning() {
			t.Error("expected IsRunning() = true after successful restart")
		}

		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = lm.Stop(stopCtx)
	})

	t.Run("exhausted retries", func(t *testing.T) {
		lm := NewLifecycleManager(DefaultConfig())
		lm.config.MaxRestarts = 2
		lm.lookPath = func(string) (string, error) {
			return "/usr/bin/piper", nil
		}
		lm.startCmd = func(name string, arg ...string) *exec.Cmd {
			return exec.Command("sleep", "10")
		}
		lm.newProcess = func(cmd *exec.Cmd) (Process, error) {
			return nil, errors.New("always fails")
		}

		ctx := context.Background()
		err := lm.Restart(ctx)
		if err == nil {
			t.Fatal("expected error after exhausted retries")
		}
	})
}
