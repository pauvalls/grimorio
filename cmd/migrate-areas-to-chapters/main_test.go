package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// helper: build a tiny stub fixture of an `areas/` directory with 2 chapter files.
func makeFixture(t *testing.T, root string) {
	t.Helper()
	areas := filepath.Join(root, "areas")
	if err := os.MkdirAll(areas, 0o755); err != nil {
		t.Fatalf("mkdir areas: %v", err)
	}
	must := func(path string, data []byte) {
		t.Helper()
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	must(filepath.Join(areas, "chapter_01_sombras.md"), []byte("# Capítulo 1: Sombras\n\nContenido uno.\n"))
	must(filepath.Join(areas, "chapter_02_traiciones.md"), []byte("# Capítulo 2: Traiciones\n\nContenido dos.\n"))
}

// runBinary compiles the current package to a temp binary and executes it
// with the supplied args, returning (stdout, stderr, exitCode, err).
func runBinary(t *testing.T, args ...string) (string, string, int, error) {
	t.Helper()
	binDir := t.TempDir()
	bin := filepath.Join(binDir, "migrate-areas-to-chapters")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go build: %v", err)
	}
	run := exec.Command(bin, args...)
	var stdout, stderr bytes.Buffer
	run.Stdout = &stdout
	run.Stderr = &stderr
	err := run.Run()
	exit := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exit = exitErr.ExitCode()
	}
	return stdout.String(), stderr.String(), exit, err
}

func TestMigrateAreasToChapters_DryRun_Default(t *testing.T) {
	root := t.TempDir()
	makeFixture(t, root)

	// No flags → dry-run, default, prints plan, exits 0
	stdout, _, exit, err := runBinary(t, "--campaign", root)
	if err != nil && exit != 0 {
		t.Fatalf("unexpected error: %v, stderr=%s", err, stdout)
	}
	if exit != 0 {
		t.Errorf("exit = %d, want 0", exit)
	}
	if !strings.Contains(stdout, "chapter_01_sombras.md") {
		t.Errorf("dry-run output must mention source files; got:\n%s", stdout)
	}

	// File system must be UNCHANGED: no chapters/ dir, no backup
	if _, err := os.Stat(filepath.Join(root, "chapters")); !os.IsNotExist(err) {
		t.Error("dry-run must NOT create chapters/ directory")
	}
	entries, _ := os.ReadDir(root)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".backup-areas-") {
			t.Errorf("dry-run must NOT create a backup dir; found %s", e.Name())
		}
	}
}

func TestMigrateAreasToChapters_RealRun_CreatesChapters(t *testing.T) {
	root := t.TempDir()
	makeFixture(t, root)

	stdout, _, exit, err := runBinary(t, "--campaign", root, "--apply")
	if err != nil && exit != 0 {
		t.Fatalf("unexpected error: %v, stdout=%s", err, stdout)
	}
	if exit != 0 {
		t.Errorf("exit = %d, want 0", exit)
	}

	// chapters/chapter-1/chapter_01_sombras.md should exist
	dst := filepath.Join(root, "chapters", "chapter-01", "chapter_01_sombras.md")
	if _, err := os.Stat(dst); err != nil {
		t.Errorf("expected migrated file at %s: %v", dst, err)
	}
	// chapters/chapter-02/...
	dst2 := filepath.Join(root, "chapters", "chapter-02", "chapter_02_traiciones.md")
	if _, err := os.Stat(dst2); err != nil {
		t.Errorf("expected migrated file at %s: %v", dst2, err)
	}
	// Backup dir must exist
	entries, _ := os.ReadDir(root)
	foundBackup := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".backup-areas-") {
			foundBackup = true
		}
	}
	if !foundBackup {
		t.Error("real run must create a .backup-areas-<ts> directory")
	}
}

func TestMigrateAreasToChapters_RefusesWithoutBackup(t *testing.T) {
	root := t.TempDir()
	makeFixture(t, root)

	_, stderr, exit, err := runBinary(t, "--campaign", root, "--apply", "--no-backup")
	if err != nil && exit == 0 {
		t.Fatalf("expected non-zero exit, got 0")
	}
	if exit == 0 {
		t.Error("expected non-zero exit when --no-backup is supplied with --apply")
	}
	if !strings.Contains(stderr, "--backup") && !strings.Contains(stderr, "backup") {
		t.Errorf("stderr should mention --backup; got: %s", stderr)
	}

	// No files should be written
	if _, err := os.Stat(filepath.Join(root, "chapters")); !os.IsNotExist(err) {
		t.Error("refusal path must not create chapters/")
	}
}

func TestMigrateAreasToChapters_EmptyAreas(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "areas"), 0o755); err != nil {
		t.Fatalf("mkdir areas: %v", err)
	}

	stdout, _, exit, err := runBinary(t, "--campaign", root)
	if err != nil && exit != 0 {
		t.Fatalf("unexpected error: %v", err)
	}
	if exit != 0 {
		t.Errorf("empty areas: exit = %d, want 0", exit)
	}
	if !strings.Contains(stdout, "No areas to migrate") {
		t.Errorf("empty areas: expected no-op message; got:\n%s", stdout)
	}
}

func TestMigrateAreasToChapters_MissingCampaign(t *testing.T) {
	_, _, exit, _ := runBinary(t)
	if exit == 0 {
		t.Error("missing --campaign should exit non-zero")
	}
}

func TestMigrateAreasToChapters_Idempotent(t *testing.T) {
	root := t.TempDir()
	makeFixture(t, root)

	// First run: apply
	_, _, exit, err := runBinary(t, "--campaign", root, "--apply")
	if err != nil && exit != 0 {
		t.Fatalf("first run failed: %v", err)
	}
	if exit != 0 {
		t.Fatalf("first run: exit = %d, want 0", exit)
	}

	// Second run: should be a no-op (no error) since destination already exists
	stdout, _, exit, err := runBinary(t, "--campaign", root, "--apply")
	if err != nil && exit != 0 {
		t.Fatalf("second run failed: %v, stdout=%s", err, stdout)
	}
	if exit != 0 {
		t.Errorf("idempotent: exit = %d, want 0", exit)
	}

	// Content must remain intact
	dst := filepath.Join(root, "chapters", "chapter-01", "chapter_01_sombras.md")
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if !strings.Contains(string(got), "Contenido uno") {
		t.Errorf("idempotent: dst content lost; got: %s", got)
	}
}

// TestMigrateAreasToChapters_BackupIsByteEquivalent verifies the backup
// is a byte-preserving copy of the source areas/ tree.
func TestMigrateAreasToChapters_BackupIsByteEquivalent(t *testing.T) {
	root := t.TempDir()
	makeFixture(t, root)

	_, _, exit, err := runBinary(t, "--campaign", root, "--apply")
	if err != nil && exit != 0 {
		t.Fatalf("run failed: %v", err)
	}

	entries, _ := os.ReadDir(root)
	var backupDir string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".backup-areas-") {
			backupDir = filepath.Join(root, e.Name())
			break
		}
	}
	if backupDir == "" {
		t.Fatal("no backup dir found")
	}

	// Compare backup areas/ with source areas/
	originalArea, err := os.ReadFile(filepath.Join(root, "areas", "chapter_01_sombras.md"))
	if err != nil {
		t.Fatalf("read original: %v", err)
	}
	backupArea, err := os.ReadFile(filepath.Join(backupDir, "areas", "chapter_01_sombras.md"))
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if !bytes.Equal(originalArea, backupArea) {
		t.Errorf("backup differs from source:\norig=%s\nback=%s", originalArea, backupArea)
	}
}

// TestMigrateAreasToChapters_PreservesContent is a content-preservation
// test: the migrated chapter file must have the same bytes as the source.
func TestMigrateAreasToChapters_PreservesContent(t *testing.T) {
	root := t.TempDir()
	makeFixture(t, root)

	_, _, exit, err := runBinary(t, "--campaign", root, "--apply")
	if err != nil && exit != 0 {
		t.Fatalf("run failed: %v", err)
	}
	if exit != 0 {
		t.Fatalf("exit = %d, want 0", exit)
	}

	srcBytes, err := os.ReadFile(filepath.Join(root, "areas", "chapter_02_traiciones.md"))
	if err != nil {
		t.Fatal(err)
	}
	dstPath := filepath.Join(root, "chapters", "chapter-02", "chapter_02_traiciones.md")
	dstBytes, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(srcBytes, dstBytes) {
		t.Errorf("content mismatch:\nsrc=%s\ndst=%s", srcBytes, dstBytes)
	}
}

// TestPlanMigration_ExcludesNonChapterFiles verifies that files not
// matching the chapter_NN_*.md pattern are skipped.
func TestPlanMigration_ExcludesNonChapterFiles(t *testing.T) {
	root := t.TempDir()
	areas := filepath.Join(root, "areas")
	if err := os.MkdirAll(areas, 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(areas, "chapter_01_ok.md"), []byte("ok"), 0o644)
	_ = os.WriteFile(filepath.Join(areas, "README.md"), []byte("readme"), 0o644)
	_ = os.WriteFile(filepath.Join(areas, "notes.txt"), []byte("notes"), 0o644)

	plan, err := planMigration(root)
	if err != nil {
		t.Fatalf("planMigration: %v", err)
	}
	if len(plan) != 1 {
		t.Errorf("plan length = %d, want 1 (only chapter_01_ok.md)", len(plan))
		for _, op := range plan {
			t.Logf("  plan op: %s", op.Src)
		}
	}
	if len(plan) == 1 && !strings.HasSuffix(plan[0].Src, "chapter_01_ok.md") {
		t.Errorf("expected chapter_01_ok.md in plan; got %s", plan[0].Src)
	}
}

// TestPlanMigration_MissingAreasDirReturnsEmpty verifies that a campaign
// without an areas/ directory is treated as a no-op (not an error).
func TestPlanMigration_MissingAreasDirReturnsEmpty(t *testing.T) {
	root := t.TempDir()
	plan, err := planMigration(root)
	if err != nil {
		t.Fatalf("planMigration with missing areas: %v", err)
	}
	if len(plan) != 0 {
		t.Errorf("plan length = %d, want 0", len(plan))
	}
}

// TestCreateBackup_TimestampedDir verifies the backup directory uses a
// Unix timestamp and is byte-equivalent to the source.
func TestCreateBackup_TimestampedDir(t *testing.T) {
	root := t.TempDir()
	makeFixture(t, root)

	dir, err := createBackup(root)
	if err != nil {
		t.Fatalf("createBackup: %v", err)
	}
	name := filepath.Base(dir)
	if !strings.HasPrefix(name, ".backup-areas-") {
		t.Errorf("backup name = %q, want prefix .backup-areas-", name)
	}
	if _, err := os.Stat(filepath.Join(dir, "areas", "chapter_01_sombras.md")); err != nil {
		t.Errorf("backup missing source file: %v", err)
	}
}

// TestRunMigration_DryRunInProcess drives runMigration via the in-process
// urfave/cli context so the action function is covered.
func TestRunMigration_DryRunInProcess(t *testing.T) {
	root := t.TempDir()
	makeFixture(t, root)
	app := buildApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr
	err := app.Run([]string{"migrate", "--campaign", root})
	if err != nil {
		t.Fatalf("app.Run: %v, stderr=%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "chapter_01_sombras.md") {
		t.Errorf("expected plan output; got: %s", stdout.String())
	}
	// filesystem must be unchanged
	if _, err := os.Stat(filepath.Join(root, "chapters")); !os.IsNotExist(err) {
		t.Error("dry-run in-process should not create chapters/")
	}
}

// TestRunMigration_ApplyRefusesWithoutBackup covers the refusal path
// in-process.
func TestRunMigration_ApplyRefusesWithoutBackup(t *testing.T) {
	root := t.TempDir()
	makeFixture(t, root)
	app := buildApp()
	var stdout, stderr bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &stderr
	err := app.Run([]string{"migrate", "--campaign", root, "--apply", "--no-backup"})
	if err == nil {
		t.Error("expected error when --no-backup is set with --apply")
	}
	if !strings.Contains(err.Error(), "backup") {
		t.Errorf("error must mention backup; got: %v", err)
	}
}

// TestRunMigration_ApplyInProcess covers the apply path in-process so
// runMigration reaches 100% coverage.
func TestRunMigration_ApplyInProcess(t *testing.T) {
	root := t.TempDir()
	makeFixture(t, root)
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &bytes.Buffer{}
	err := app.Run([]string{"migrate", "--campaign", root, "--apply"})
	if err != nil {
		t.Fatalf("app.Run: %v", err)
	}
	if !strings.Contains(stdout.String(), "Migration complete") {
		t.Errorf("expected 'Migration complete' in output; got: %s", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(root, "chapters", "chapter-01", "chapter_01_sombras.md")); err != nil {
		t.Errorf("expected migrated file: %v", err)
	}
}

// TestRunMigration_EmptyAreasInProcess covers the empty-areas no-op path
// in-process.
func TestRunMigration_EmptyAreasInProcess(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "areas"), 0o755); err != nil {
		t.Fatal(err)
	}
	app := buildApp()
	var stdout bytes.Buffer
	app.Writer = &stdout
	app.ErrWriter = &bytes.Buffer{}
	err := app.Run([]string{"migrate", "--campaign", root})
	if err != nil {
		t.Fatalf("app.Run: %v", err)
	}
	if !strings.Contains(stdout.String(), "No areas to migrate") {
		t.Errorf("expected no-op message; got: %s", stdout.String())
	}
}
