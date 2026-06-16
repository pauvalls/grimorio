package compiler

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// captureLog redirects the standard logger to an in-memory buffer for the
// duration of the test. It restores the previous default output in t.Cleanup
// and is intentionally race-safe: the test must NOT call t.Parallel().
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	prevOut := log.Default().Writer()
	prevFlags := log.Default().Flags()
	log.SetOutput(buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(prevOut)
		log.SetFlags(prevFlags)
	})
	return buf
}

// scaffoldCampaign creates a minimal campaign directory with optional
// chapters/ and areas/ subdirs. It returns the temp dir and a teardown func.
func scaffoldCampaign(t *testing.T, withChapters, withAreas bool, chapterCount, areaCount int) string {
	t.Helper()
	dir := t.TempDir()
	if withChapters {
		chDir := filepath.Join(dir, "chapters")
		if err := os.MkdirAll(chDir, 0755); err != nil {
			t.Fatalf("mkdir chapters: %v", err)
		}
		for i := 0; i < chapterCount; i++ {
			name := filepath.Join(chDir, "chapter_"+strconv.Itoa(i)+".md")
			if err := os.WriteFile(name, []byte("# Chapter\n"), 0644); err != nil {
				t.Fatalf("write chapter: %v", err)
			}
		}
	}
	if withAreas {
		aDir := filepath.Join(dir, "areas")
		if err := os.MkdirAll(aDir, 0755); err != nil {
			t.Fatalf("mkdir areas: %v", err)
		}
		for i := 0; i < areaCount; i++ {
			name := filepath.Join(aDir, "area_"+strconv.Itoa(i)+".md")
			if err := os.WriteFile(name, []byte("# Area\n"), 0644); err != nil {
				t.Fatalf("write area: %v", err)
			}
		}
	}
	return dir
}

// TestChaptersAndAreasWarn asserts that a warning is emitted when BOTH
// chapters/ and areas/ exist; the message must name the dropped dir and
// a positive count.
func TestChaptersAndAreasWarn(t *testing.T) {
	dir := scaffoldCampaign(t, true, true, 1, 1)
	buf := captureLog(t)

	c := New(dir, "")
	_, err := c.generateHTML("Test")
	if err != nil {
		t.Fatalf("generateHTML error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "[warn] grimorio: both chapters/ and areas/ exist") {
		t.Errorf("missing warn prefix in log output: %q", got)
	}
	if !strings.Contains(got, "using chapters/") {
		t.Errorf("missing 'using chapters/' in log output: %q", got)
	}
	if !strings.Contains(got, "dropping") {
		t.Errorf("missing 'dropping' in log output: %q", got)
	}
}

// TestChaptersOnlySilent asserts that no warning is emitted when only chapters/
// is present.
func TestChaptersOnlySilent(t *testing.T) {
	dir := scaffoldCampaign(t, true, false, 1, 0)
	buf := captureLog(t)

	c := New(dir, "")
	_, err := c.generateHTML("Test")
	if err != nil {
		t.Fatalf("generateHTML error: %v", err)
	}

	if strings.Contains(buf.String(), "[warn] grimorio: both chapters/ and areas/ exist") {
		t.Errorf("unexpected warn when only chapters/ present: %q", buf.String())
	}
}

// TestAreasOnlySilent asserts that no warning is emitted when only areas/
// is present.
func TestAreasOnlySilent(t *testing.T) {
	dir := scaffoldCampaign(t, false, true, 0, 1)
	buf := captureLog(t)

	c := New(dir, "")
	_, err := c.generateHTML("Test")
	if err != nil {
		t.Fatalf("generateHTML error: %v", err)
	}

	if strings.Contains(buf.String(), "[warn] grimorio: both chapters/ and areas/ exist") {
		t.Errorf("unexpected warn when only areas/ present: %q", buf.String())
	}
}

// TestDroppedCountMatchesAreasFiles asserts that the warning message names the
// correct count of dropped area files (here: 7).
func TestDroppedCountMatchesAreasFiles(t *testing.T) {
	dir := scaffoldCampaign(t, true, true, 3, 7)
	buf := captureLog(t)

	c := New(dir, "")
	_, err := c.generateHTML("Test")
	if err != nil {
		t.Fatalf("generateHTML error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "dropping 7") {
		t.Errorf("expected 'dropping 7' in log output: %q", got)
	}
}
