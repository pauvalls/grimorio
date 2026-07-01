package compiler

import (
	"context"
	"testing"
)

// TestCompile_ChromiumFlags asserts the flags passed to the Chromium headless
// binary. Before the WU-2 flag switch, the legacy --print-to-pdf-no-header is
// present and --no-pdf-header-footer + --paper=A4 are absent. This test is the
// regression baseline that WU-2 will flip to the new contract.
func TestCompile_ChromiumFlags(t *testing.T) {
	c := New("/tmp", "test-engine")
	cmd := c.buildChromiumCmd(context.Background(), "/tmp/in.html", "/tmp/out.pdf")
	if cmd == nil {
		t.Fatal("buildChromiumCmd returned nil")
	}
	args := cmd.Args
	found := map[string]bool{}
	for _, a := range args {
		found[a] = true
	}
	// WU-1 baseline: legacy flag present, new flags absent.
	if !found["--print-to-pdf-no-header"] {
		t.Errorf("expected legacy --print-to-pdf-no-header in baseline args, got: %v", args)
	}
	if found["--no-pdf-header-footer"] {
		t.Errorf("--no-pdf-header-footer should not be present in baseline args, got: %v", args)
	}
	if found["--paper=A4"] {
		t.Errorf("--paper=A4 should not be present in baseline args, got: %v", args)
	}
}
