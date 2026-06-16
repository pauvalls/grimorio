// migrate-areas-to-chapters converts legacy campaign layouts
//
//	<campaign>/areas/chapter_NN_*.md
//
// into the canonical WotC layout
//
//	<campaign>/chapters/chapter-NN/<basename>.md
//
// The tool is safe by default: with no flags, it runs in dry-run mode,
// prints the migration plan, and exits 0 without touching the
// filesystem. To actually mutate state, pass --apply. The migration
// requires a backup unless --no-backup is set; supplying both
// --apply and --no-backup is a hard error.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/urfave/cli/v2"
)

// chapterFileRE matches source area files like "chapter_01_sombras.md"
// or "chapter_1_foo.md". Captures the zero-padded number and the rest.
var chapterFileRE = regexp.MustCompile(`^chapter_(\d+)_([^/]+)\.md$`)

func main() {
	app := buildApp()
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// buildApp constructs the urfave/cli app. Exposed for in-process tests.
func buildApp() *cli.App {
	return &cli.App{
		Name:  "migrate-areas-to-chapters",
		Usage: "Migrate legacy campaign areas/ to chapters/ layout",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "campaign",
				Aliases:  []string{"c"},
				Usage:    "Path to the campaign directory containing areas/",
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "apply",
				Aliases: []string{"a"},
				Usage:   "Actually mutate the filesystem (default: dry-run)",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:  "no-backup",
				Usage: "Skip creating a backup (only allowed in dry-run mode)",
				Value: false,
			},
		},
		Action: runMigration,
	}
}

func runMigration(c *cli.Context) error {
	campaignDir := c.String("campaign")
	apply := c.Bool("apply")
	noBackup := c.Bool("no-backup")
	out := c.App.Writer
	if out == nil {
		out = os.Stdout
	}

	// Hard refusal: --apply + --no-backup is not allowed.
	if apply && noBackup {
		return fmt.Errorf("refusing to run: --apply requires a backup; remove --no-backup (or use the default --dry-run)")
	}

	// Run the migration plan.
	plan, err := planMigration(campaignDir)
	if err != nil {
		return err
	}
	if len(plan) == 0 {
		fmt.Fprintln(out, "No areas to migrate (areas/ is empty or missing).")
		return nil
	}

	// Print the plan in both modes.
	for _, op := range plan {
		fmt.Fprintf(out, "  %s (%d bytes) -> %s\n", op.Src, op.Size, op.Dst)
	}

	if !apply {
		fmt.Fprintf(out, "\nDry-run complete. %d file(s) would be migrated. Re-run with --apply to execute.\n", len(plan))
		return nil
	}

	// Real run: create backup first.
	if !noBackup {
		backupDir, err := createBackup(campaignDir)
		if err != nil {
			return fmt.Errorf("create backup: %w", err)
		}
		fmt.Fprintf(out, "Backup created: %s\n", backupDir)
	}

	// Execute the plan.
	for _, op := range plan {
		if err := copyFile(op.Src, op.Dst); err != nil {
			return fmt.Errorf("copy %s -> %s: %w", op.Src, op.Dst, err)
		}
	}
	fmt.Fprintf(out, "\nMigration complete: %d file(s) copied to chapters/.\n", len(plan))
	return nil
}

// migrationOp represents a single source -> destination copy.
type migrationOp struct {
	Src  string
	Dst  string
	Size int64
}

// planMigration walks <campaign>/areas/ and produces a slice of operations.
// It never mutates the filesystem.
func planMigration(campaignDir string) ([]migrationOp, error) {
	areasDir := filepath.Join(campaignDir, "areas")
	entries, err := os.ReadDir(areasDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read areas/: %w", err)
	}
	// Sort entries by name for deterministic output.
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	var plan []migrationOp
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		m := chapterFileRE.FindStringSubmatch(name)
		if m == nil {
			// Not a chapter_NN_*.md file — skip silently (preserves
			// unknown assets; we only migrate what matches the pattern).
			continue
		}
		chapterNum := m[1]
		basename := name // preserve the full original filename
		chapterDir := filepath.Join(campaignDir, "chapters", "chapter-"+chapterNum)
		dst := filepath.Join(chapterDir, basename)

		src := filepath.Join(areasDir, name)
		info, err := os.Stat(src)
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", src, err)
		}
		plan = append(plan, migrationOp{
			Src:  src,
			Dst:  dst,
			Size: info.Size(),
		})
	}
	return plan, nil
}

// createBackup byte-copies the entire areas/ tree to a timestamped backup
// directory at <campaignDir>/.backup-areas-<unix-ts>/. The on-disk layout
// mirrors the source tree so the user can restore with `cp -r`.
func createBackup(campaignDir string) (string, error) {
	ts := time.Now().Unix()
	backupDir := filepath.Join(campaignDir, fmt.Sprintf(".backup-areas-%d", ts))
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir backup: %w", err)
	}
	src := filepath.Join(campaignDir, "areas")
	if _, err := os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			return backupDir, nil
		}
		return "", err
	}
	dst := filepath.Join(backupDir, "areas")
	if err := copyTree(src, dst); err != nil {
		return "", err
	}
	return backupDir, nil
}

// copyFile byte-copies src -> dst, creating parent dirs as needed.
// Existing dst is overwritten (caller is expected to plan ordering or
// accept last-writer-wins). For idempotency, the migration plan
// generation only emits unique destinations.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

// copyTree recursively byte-copies srcDir to dstDir.
func copyTree(srcDir, dstDir string) error {
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		src := filepath.Join(srcDir, e.Name())
		dst := filepath.Join(dstDir, e.Name())
		if e.IsDir() {
			if err := copyTree(src, dst); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}
