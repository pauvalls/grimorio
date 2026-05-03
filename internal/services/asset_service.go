package services

import (
	"fmt"
	"path/filepath"

	"github.com/pauvalls/grimorio/internal/image"
	"github.com/pauvalls/grimorio/internal/svg"
)

// AssetService handles asset generation (maps, images, dividers)
type AssetService struct {
	baseDir   string
	imgConfig image.Config
}

// NewAssetService creates a new asset service
func NewAssetService(baseDir string, imgConfig image.Config) *AssetService {
	return &AssetService{
		baseDir:   baseDir,
		imgConfig: imgConfig,
	}
}

func (s *AssetService) campaignDir(campaign string) string {
	return filepath.Join(s.baseDir, campaign)
}

func (s *AssetService) assetsDir(campaign string) string {
	return filepath.Join(s.campaignDir(campaign), "assets")
}

// GenerateMap generates a procedural SVG battle map
func (s *AssetService) GenerateMap(campaign, filename, style, title string, rooms int, labels []string) (string, error) {
	svgCfg := svg.DefaultBattleMapConfig()
	if style != "" {
		svgCfg.Style = svg.MapStyle(style)
	}
	if title != "" {
		svgCfg.Title = title
	}
	svgCfg.NumRooms = rooms
	if len(labels) > 0 {
		svgCfg.Labels = labels
	}

	svgContent := svg.GenerateBattleMap(svgCfg)
	// Note: The caller handles file writing to maintain compatibility
	return svgContent, nil
}

// GenerateDivider generates a decorative SVG divider
func (s *AssetService) GenerateDivider(campaign, filename, style string, width int) (string, error) {
	if style == "" {
		style = "ornate"
	}

	svgContent := svg.GenerateDivider(width, style)
	// Note: The caller handles file writing
	return svgContent, nil
}

// GenerateImage generates an image using AI
func (s *AssetService) GenerateImage(campaign, filename, prompt, imgType string) (string, error) {
	provider, err := image.NewProvider(s.imgConfig)
	if err != nil {
		return "", fmt.Errorf("failed to initialize image provider: %w", err)
	}

	outputPath := filepath.Join(s.assetsDir(campaign), filename+".png")
	if err := image.GenerateAndSave(provider, prompt, outputPath); err != nil {
		return "", fmt.Errorf("%s generation failed: %w", provider.Name(), err)
	}

	return outputPath, nil
}

// GenerateImagesBatch generates multiple images in parallel
func (s *AssetService) GenerateImagesBatch(campaign string, images []BatchImageSpec) ([]BatchImageResult, error) {
	provider, err := image.NewProvider(s.imgConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize image provider: %w", err)
	}

	results := make([]BatchImageResult, len(images))
	
	for i, spec := range images {
		outputPath := filepath.Join(s.assetsDir(campaign), spec.Filename+".png")
		if err := image.GenerateAndSave(provider, spec.Prompt, outputPath); err != nil {
			results[i] = BatchImageResult{
				Filename: spec.Filename,
				Success:  false,
				Error:    err.Error(),
			}
		} else {
			results[i] = BatchImageResult{
				Filename: spec.Filename,
				Success:  true,
				Path:     outputPath,
			}
		}
	}

	return results, nil
}

// BatchImageSpec represents an image to generate
type BatchImageSpec struct {
	Filename string
	Prompt   string
	Type     string
}

// BatchImageResult represents the result of generating an image
type BatchImageResult struct {
	Filename string
	Success  bool
	Path     string
	Error    string
}
