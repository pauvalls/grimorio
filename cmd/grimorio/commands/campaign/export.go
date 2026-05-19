package campaign

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pauvalls/grimorio/internal/config"
	"github.com/urfave/cli/v2"
)

// NewExportCommand returns the "campaign export" subcommand.
func NewExportCommand() *cli.Command {
	return &cli.Command{
		Name:      "export",
		Usage:     "Export a campaign directory as tar.gz archive",
		ArgsUsage: "<campaign>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file path (default: <campaign>.tar.gz)",
			},
		},
		Action: func(cCtx *cli.Context) error {
			return runExport(cCtx)
		},
	}
}

func runExport(cCtx *cli.Context) error {
	if cCtx.NArg() < 1 {
		return fmt.Errorf("usage: grimorio campaign export <campaign> [--output out.tar.gz]")
	}

	campaignName := cCtx.Args().First()
	outputPath := cCtx.String("output")

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	configPath := filepath.Join(home, ".config", "grimorio", "config.json")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	campaignDir := filepath.Join(cfg.OutputDir, campaignName)
	info, err := os.Stat(campaignDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("campaign not found: %s", campaignName)
	}

	if outputPath == "" {
		outputPath = campaignName + ".tar.gz"
	}

	if err := CreateTarGz(campaignDir, outputPath); err != nil {
		return fmt.Errorf("error creating archive: %w", err)
	}

	fmt.Printf("Campaign '%s' exported to %s\n", campaignName, outputPath)
	return nil
}

// CreateTarGz creates a tar.gz archive of the given directory. Exported for testing.
// The archive includes the base directory name as the top-level entry so that
// extraction naturally creates a subdirectory named after the campaign.
func CreateTarGz(sourceDir, targetFile string) error {
	outFile, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("cannot create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	gzWriter := gzip.NewWriter(outFile)
	defer func() { _ = gzWriter.Close() }()

	tarWriter := tar.NewWriter(gzWriter)
	defer func() { _ = tarWriter.Close() }()

	baseName := filepath.Base(sourceDir)

	// Write the top-level directory entry
	if err := tarWriter.WriteHeader(&tar.Header{
		Name:     baseName + "/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}); err != nil {
		return fmt.Errorf("error writing directory header: %w", err)
	}

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		// Prefix with base directory name
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("error creating tar header for %s: %w", relPath, err)
		}
		header.Name = baseName + "/" + relPath

		if info.IsDir() {
			header.Name += "/"
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("error writing tar header for %s: %w", relPath, err)
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("error opening file %s: %w", relPath, err)
			}
			defer func() { _ = file.Close() }()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return fmt.Errorf("error writing file %s to archive: %w", relPath, err)
			}
		}

		return nil
	})
}
