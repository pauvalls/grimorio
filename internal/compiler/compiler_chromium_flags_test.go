package compiler

import (
	"context"
	"testing"
)

// TestCompile_ChromiumFlags asserts the flags passed to the Chromium headless
// binary. REQ-1.1: must use --no-pdf-header-footer (replaces the legacy
// --print-to-pdf-no-header) and --paper=A4 so the print surface matches the
// CSS-implied A4. The old flag must NOT be present.
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
	if !found["--no-pdf-header-footer"] {
		t.Errorf("missing --no-pdf-header-footer flag, args: %v", args)
	}
	if !found["--paper=A4"] {
		t.Errorf("missing --paper=A4 flag, args: %v", args)
	}
	if found["--print-to-pdf-no-header"] {
		t.Errorf("legacy --print-to-pdf-no-header still present, args: %v", args)
	}
}
