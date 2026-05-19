package campaign

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pauvalls/grimorio/internal/config"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/urfave/cli/v2"
)

// NewImportCommand returns the "campaign import" subcommand.
func NewImportCommand() *cli.Command {
	return &cli.Command{
		Name:      "import",
		Usage:     "Import a campaign from a tar.gz archive",
		ArgsUsage: "<file.tar.gz>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "force",
				Usage: "Overwrite existing campaign directory",
			},
		},
		Action: func(cCtx *cli.Context) error {
			return runImport(cCtx)
		},
	}
}

func runImport(cCtx *cli.Context) error {
	if cCtx.NArg() < 1 {
		return fmt.Errorf("usage: grimorio campaign import <file.tar.gz> [--force]")
	}

	archivePath := cCtx.Args().First()
	force := cCtx.Bool("force")

	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", archivePath)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	configPath := filepath.Join(home, ".config", "grimorio", "config.json")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	campaignRepo := repository.NewFilesystemCampaignRepository(cfg.OutputDir)

	campaignName, err := GetArchiveTopLevelDir(archivePath)
	if err != nil {
		return fmt.Errorf("invalid archive format: %w", err)
	}
	if campaignName == "" {
		return fmt.Errorf("invalid archive format: empty archive")
	}

	targetDir := filepath.Join(cfg.OutputDir, campaignName)

	if campaignRepo.Exists(campaignName) {
		if !force {
			return fmt.Errorf("campaign exists. Use --force to overwrite")
		}
		if err := os.RemoveAll(targetDir); err != nil {
			return fmt.Errorf("error removing existing campaign: %w", err)
		}
	}

	if err := ExtractTarGz(archivePath, cfg.OutputDir); err != nil {
		return fmt.Errorf("error extracting archive: %w", err)
	}

	fmt.Printf("Campaign '%s' imported successfully\n", campaignName)
	return nil
}

// GetArchiveTopLevelDir reads the first entry of a tar.gz archive to determine
// the top-level directory name. Exported for testing.
func GetArchiveTopLevelDir(archivePath string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("not a valid gzip file: %w", err)
	}
	defer func() { _ = gzReader.Close() }()

	tarReader := tar.NewReader(gzReader)

	header, err := tarReader.Next()
	if err == io.EOF {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("error reading tar: %w", err)
	}

	parts := strings.SplitN(header.Name, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "", nil
	}
	return parts[0], nil
}

// ExtractTarGz extracts a tar.gz archive into the target directory.
// It guards against directory traversal attacks (../) in archive paths.
// Exported for testing.
func ExtractTarGz(archivePath, targetDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("cannot open archive: %w", err)
	}
	defer func() { _ = file.Close() }()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("not a valid gzip file: %w", err)
	}
	defer func() { _ = gzReader.Close() }()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar entry: %w", err)
		}

		if strings.Contains(header.Name, "..") {
			return fmt.Errorf("security error: archive contains path with '..' (%s)", header.Name)
		}

		cleanName := filepath.Clean(header.Name)
		targetPath := filepath.Join(targetDir, cleanName)

		if !strings.HasPrefix(targetPath, filepath.Clean(targetDir)+string(os.PathSeparator)) &&
			targetPath != filepath.Clean(targetDir) {
			return fmt.Errorf("security error: archive path escapes target directory (%s)", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("error creating directory %s: %w", targetPath, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("error creating parent directory for %s: %w", targetPath, err)
			}
			outFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("error creating file %s: %w", targetPath, err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("error writing file %s: %w", targetPath, err)
			}
			outFile.Close()
		}
	}

	return nil
}
