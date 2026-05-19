package campaign

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/urfave/cli/v2"
)

// binaryExts lists file extensions that should be skipped during diff.
var binaryExts = map[string]bool{
	".svg":    true,
	".pdf":    true,
	".png":    true,
	".jpg":    true,
	".jpeg":   true,
	".gif":    true,
	".webp":   true,
	".bmp":    true,
	".ico":    true,
	".zip":    true,
	".gz":     true,
}

// isBinaryOrSkipped returns true if the file extension is in the skip list.
func isBinaryOrSkipped(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".gz" && strings.HasSuffix(strings.ToLower(name), ".tar.gz") {
		return true
	}
	return binaryExts[ext]
}

// NewDiffCommand returns the "campaign diff" subcommand.
func NewDiffCommand() *cli.Command {
	return &cli.Command{
		Name:      "diff",
		Usage:     "Show markdown diff between two campaigns",
		ArgsUsage: "<campaign-a> <campaign-b>",
		Action: func(cCtx *cli.Context) error {
			return runDiff(cCtx)
		},
	}
}

func runDiff(cCtx *cli.Context) error {
	if cCtx.NArg() < 2 {
		return fmt.Errorf("usage: grimorio campaign diff <campaign-a> <campaign-b>")
	}

	campaignA := cCtx.Args().Get(0)
	campaignB := cCtx.Args().Get(1)

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	campaignsDir := filepath.Join(home, "campaigns")

	dirA := filepath.Join(campaignsDir, campaignA)
	dirB := filepath.Join(campaignsDir, campaignB)

	// Validate both directories exist
	infoA, err := os.Stat(dirA)
	if err != nil || !infoA.IsDir() {
		return fmt.Errorf("campaign not found: %s", campaignA)
	}
	infoB, err := os.Stat(dirB)
	if err != nil || !infoB.IsDir() {
		return fmt.Errorf("campaign not found: %s", campaignB)
	}

	// Build file maps
	filesA := buildFileMap(dirA)
	filesB := buildFileMap(dirB)

	dmp := diffmatchpatch.New()
	hasDifferences := false

	// Files in both — diff them
	for relPath, pathA := range filesA {
		pathB, inB := filesB[relPath]
		if !inB {
			fmt.Printf("Only in %s: %s\n", campaignA, relPath)
			hasDifferences = true
			continue
		}

		if isBinaryOrSkipped(relPath) {
			fmt.Printf("Skipping binary file: %s\n", relPath)
			continue
		}

		contentA, errA := readFile(pathA)
		contentB, errB := readFile(pathB)
		if errA != nil || errB != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", relPath, errA)
			continue
		}

		if contentA == contentB {
			continue
		}

		hasDifferences = true
		diffs := dmp.DiffMain(contentA, contentB, true)
		diffs = dmp.DiffCleanupSemantic(diffs)
		prettyDiff := dmp.DiffPrettyText(diffs)

		fmt.Printf("--- %s/%s\n", campaignA, relPath)
		fmt.Printf("+++ %s/%s\n", campaignB, relPath)
		fmt.Println(prettyDiff)
	}

	// Files only in B
	for relPath := range filesB {
		if _, inA := filesA[relPath]; !inA {
			fmt.Printf("Only in %s: %s\n", campaignB, relPath)
			hasDifferences = true
		}
	}

	if !hasDifferences {
		fmt.Println("No differences found")
	}

	return nil
}

// buildFileMap returns a map of relative path → absolute path for all files in dir.
func buildFileMap(dir string) map[string]string {
	files := make(map[string]string)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		files[rel] = path
		return nil
	})
	return files
}

// readFile reads a file and returns its contents as a string.
func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
