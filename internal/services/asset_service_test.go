package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/image"
)

// mockProvider is a test double for image.Provider
type mockProvider struct {
	name         string
	generateFunc func(prompt string) ([]byte, error)
	callCount    atomic.Int32
	configured   bool
}

func (m *mockProvider) Generate(prompt string) ([]byte, error) {
	m.callCount.Add(1)
	if m.generateFunc != nil {
		return m.generateFunc(prompt)
	}
	return []byte("fake-image-data-" + prompt), nil
}

func (m *mockProvider) IsConfigured() bool { return m.configured }
func (m *mockProvider) Name() string       { return m.name }

func setupTestAssetService(t *testing.T) (*AssetService, string) {
	t.Helper()
	tmpDir := t.TempDir()
	provider := &mockProvider{name: "mock", configured: true}
	service := NewAssetServiceWithProvider(tmpDir, provider)
	return service, tmpDir
}

func TestGenerateImage_Success(t *testing.T) {
	service, tmpDir := setupTestAssetService(t)

	path, err := service.GenerateImage("test-campaign", "cover-art", "epic fantasy cover", "cover")
	if err != nil {
		t.Fatalf("GenerateImage() error: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "test-campaign", "assets", "cover-art.png")
	if path != expectedPath {
		t.Errorf("GenerateImage() path = %s, want %s", path, expectedPath)
	}

	// Verify file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("GenerateImage() did not create file at %s", path)
	}
}

func TestGenerateImage_ProviderError(t *testing.T) {
	tmpDir := t.TempDir()
	provider := &mockProvider{
		name: "mock",
		generateFunc: func(prompt string) ([]byte, error) {
			return nil, fmt.Errorf("provider error: simulated failure")
		},
	}
	service := NewAssetServiceWithProvider(tmpDir, provider)

	_, err := service.GenerateImage("test-campaign", "fail-test", "prompt", "cover")
	if err == nil {
		t.Fatal("GenerateImage() expected error, got nil")
	}

	if !strings.Contains(err.Error(), "provider error") {
		t.Errorf("GenerateImage() error = %v, want containing 'provider error'", err)
	}
}

func TestGenerateImage_Sequential(t *testing.T) {
	service, tmpDir := setupTestAssetService(t)

	// Track call times to verify sequential execution
	var callTimes []time.Time
	provider := &mockProvider{
		name: "mock",
		generateFunc: func(prompt string) ([]byte, error) {
			callTimes = append(callTimes, time.Now())
			return []byte("fake-image-data-" + prompt), nil
		},
	}
	service = NewAssetServiceWithProvider(tmpDir, provider)

	// Generate multiple images
	start := time.Now()
	for i := 0; i < 3; i++ {
		_, err := service.GenerateImage("test-campaign", fmt.Sprintf("img-%d", i), "prompt", "cover")
		if err != nil {
			t.Fatalf("GenerateImage() error: %v", err)
		}
	}
	elapsed := time.Since(start)

	// Should take at least 6 seconds (3 images * 2s minimum between them)
	// Each image has 3s delay
	if elapsed < 6*time.Second {
		t.Errorf("Sequential generation took %v, expected at least 6s", elapsed)
	}

	// Verify images were created
	for i := 0; i < 3; i++ {
		path := filepath.Join(tmpDir, "test-campaign", "assets", fmt.Sprintf("img-%d.png", i))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("GenerateImage() did not create file: %s", path)
		}
	}
}

func TestGenerateImage_WithProviderFallback(t *testing.T) {
	tmpDir := t.TempDir()

	// Primary provider always fails
	primary := &mockProvider{
		name: "primary",
		generateFunc: func(prompt string) ([]byte, error) {
			return nil, fmt.Errorf("primary always fails")
		},
	}

	// Fallback provider succeeds
	fallback := &mockProvider{
		name: "fallback",
		generateFunc: func(prompt string) ([]byte, error) {
			return []byte("fallback-image-" + prompt), nil
		},
	}

	providers := []image.Provider{primary, fallback}

	outputPath := filepath.Join(tmpDir, "test-campaign", "assets", "test-image.png")
	providerName, err := image.GenerateAndSaveWithFallback(providers, "test prompt", outputPath)

	if err != nil {
		t.Fatalf("GenerateAndSaveWithFallback() error: %v", err)
	}

	if providerName != "fallback" {
		t.Errorf("Expected fallback provider, got %s", providerName)
	}

	// Verify file was created by fallback
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Fallback did not create file at %s", outputPath)
	}

	// Primary should have been called
	if primary.callCount.Load() != 1 {
		t.Errorf("Primary provider should have been called once, got %d", primary.callCount.Load())
	}

	// Fallback should have been called
	if fallback.callCount.Load() != 1 {
		t.Errorf("Fallback provider should have been called once, got %d", fallback.callCount.Load())
	}
}

func TestGenerateImage_AllProvidersFail(t *testing.T) {
	tmpDir := t.TempDir()

	// All providers fail
	provider1 := &mockProvider{
		name: "provider1",
		generateFunc: func(prompt string) ([]byte, error) {
			return nil, fmt.Errorf("provider1 error")
		},
	}
	provider2 := &mockProvider{
		name: "provider2",
		generateFunc: func(prompt string) ([]byte, error) {
			return nil, fmt.Errorf("provider2 error")
		},
	}

	providers := []image.Provider{provider1, provider2}
	outputPath := filepath.Join(tmpDir, "test-campaign", "assets", "test-image.png")
	_, err := image.GenerateAndSaveWithFallback(providers, "test prompt", outputPath)

	if err == nil {
		t.Fatal("Expected error when all providers fail")
	}

	if !strings.Contains(err.Error(), "all providers failed") {
		t.Errorf("Expected 'all providers failed' error, got: %v", err)
	}

	// Both providers should have been tried
	if provider1.callCount.Load() != 1 {
		t.Errorf("Provider1 should have been called once, got %d", provider1.callCount.Load())
	}
	if provider2.callCount.Load() != 1 {
		t.Errorf("Provider2 should have been called once, got %d", provider2.callCount.Load())
	}
}

func TestGenerateMap_Success(t *testing.T) {
	service, tmpDir := setupTestAssetService(t)

	path, err := service.GenerateMap("test", "dungeon.svg", "dungeon", "The Dark Cave", 5, []string{"Entrance", "Boss"})
	if err != nil {
		t.Fatalf("GenerateMap() error: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "test", "assets", "dungeon.svg")
	if path != expectedPath {
		t.Errorf("GenerateMap() path = %s, want %s", path, expectedPath)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("GenerateMap() file not found at %s: %v", path, err)
	}

	if !strings.Contains(string(data), "<svg") {
		t.Error("GenerateMap() file should contain SVG markup")
	}

	if !strings.Contains(string(data), "The Dark Cave") {
		t.Error("GenerateMap() file should contain the title")
	}
}

func TestGenerateDivider_Success(t *testing.T) {
	service, tmpDir := setupTestAssetService(t)

	path, err := service.GenerateDivider("test", "divider.svg", "ornate", 600)
	if err != nil {
		t.Fatalf("GenerateDivider() error: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "test", "assets", "divider.svg")
	if path != expectedPath {
		t.Errorf("GenerateDivider() path = %s, want %s", path, expectedPath)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("GenerateDivider() file not found at %s: %v", path, err)
	}

	if !strings.Contains(string(data), "<svg") {
		t.Error("GenerateDivider() file should contain SVG markup")
	}
}

func TestGetProviders_WithInjectedProvider(t *testing.T) {
	provider := &mockProvider{name: "injected"}
	service := NewAssetServiceWithProvider("/tmp", provider)

	providers := service.getProviders()
	if len(providers) != 1 {
		t.Fatalf("getProviders() returned %d providers, want 1", len(providers))
	}

	if providers[0].Name() != "injected" {
		t.Errorf("getProviders()[0] = %s, want 'injected'", providers[0].Name())
	}
}

func TestGetProviders_FallsBackToChain(t *testing.T) {
	service := NewAssetService("/tmp", image.DefaultConfig())

	providers := service.getProviders()
	if len(providers) == 0 {
		t.Fatal("getProviders() returned empty providers")
	}

	// Should have pollinations in the chain
	hasPollinations := false
	for _, p := range providers {
		if p.Name() == "pollinations" {
			hasPollinations = true
			break
		}
	}
	if !hasPollinations {
		t.Errorf("getProviders() should include 'pollinations' in fallback chain")
	}
}

func TestGenerateImage_DoubleExtension(t *testing.T) {
	service, tmpDir := setupTestAssetService(t)

	// filename already includes .png extension
	path, err := service.GenerateImage("test-campaign", "morgus-portrait.png", "A wizard portrait", "portrait")
	if err != nil {
		t.Fatalf("GenerateImage() error: %v", err)
	}

	// Should NOT create morgus-portrait.png.png
	badPath := filepath.Join(tmpDir, "test-campaign", "assets", "morgus-portrait.png.png")
	if _, err := os.Stat(badPath); !os.IsNotExist(err) {
		t.Errorf("GenerateImage() created double extension file: %s", badPath)
	}

	// Should create morgus-portrait.png
	goodPath := filepath.Join(tmpDir, "test-campaign", "assets", "morgus-portrait.png")
	if path != goodPath {
		t.Errorf("GenerateImage() path = %s, want %s", path, goodPath)
	}

	if _, err := os.Stat(goodPath); os.IsNotExist(err) {
		t.Errorf("GenerateImage() did not create file at %s", goodPath)
	}
}

func TestInsertImageReference_AppendToEnd(t *testing.T) {
	service, tmpDir := setupTestAssetService(t)

	// Create a markdown file
	campaignDir := filepath.Join(tmpDir, "test-campaign")
	os.MkdirAll(campaignDir, 0755)
	mdPath := filepath.Join(campaignDir, "npcs.md")
	originalContent := "# NPCs\n\n## Gandalf\nA powerful wizard."
	os.WriteFile(mdPath, []byte(originalContent), 0644)

	// Insert image reference without section (append to end)
	err := service.InsertImageReference("test-campaign", "npcs.md", "", "Gandalf Portrait", "gandalf.png")
	if err != nil {
		t.Fatalf("InsertImageReference() error: %v", err)
	}

	// Verify content
	content, _ := os.ReadFile(mdPath)
	if !strings.Contains(string(content), "![Gandalf Portrait](assets/gandalf.png)") {
		t.Errorf("InsertImageReference() did not append image reference. Got:\n%s", string(content))
	}
}

func TestInsertImageReference_InsertAfterSection(t *testing.T) {
	service, tmpDir := setupTestAssetService(t)

	// Create a markdown file with sections
	campaignDir := filepath.Join(tmpDir, "test-campaign")
	os.MkdirAll(campaignDir, 0755)
	mdPath := filepath.Join(campaignDir, "npcs.md")
	originalContent := "# NPCs\n\n## Gandalf\nA powerful wizard.\n\n## Saruman\nA fallen wizard."
	os.WriteFile(mdPath, []byte(originalContent), 0644)

	// Insert image reference after Gandalf section
	err := service.InsertImageReference("test-campaign", "npcs.md", "Gandalf", "Gandalf Portrait", "gandalf.png")
	if err != nil {
		t.Fatalf("InsertImageReference() error: %v", err)
	}

	// Verify content
	content, _ := os.ReadFile(mdPath)
	result := string(content)
	if !strings.Contains(result, "![Gandalf Portrait](assets/gandalf.png)") {
		t.Errorf("InsertImageReference() did not insert image reference. Got:\n%s", result)
	}
	// Should be after Gandalf section and before Saruman
	gandalfIdx := strings.Index(result, "## Gandalf")
	sarumanIdx := strings.Index(result, "## Saruman")
	imgIdx := strings.Index(result, "![Gandalf Portrait]")
	if imgIdx < gandalfIdx || imgIdx > sarumanIdx {
		t.Errorf("InsertImageReference() inserted at wrong position. Got:\n%s", result)
	}
}

func TestInsertImageReference_MarkdownFileNotFound(t *testing.T) {
	service, _ := setupTestAssetService(t)

	err := service.InsertImageReference("test-campaign", "nonexistent.md", "", "Alt", "img.png")
	if err == nil {
		t.Error("InsertImageReference() expected error for nonexistent file")
	}
}

func TestHeadingLevel(t *testing.T) {
	tests := []struct {
		line string
		want int
	}{
		{"# Heading 1", 1},
		{"## Heading 2", 2},
		{"### Heading 3", 3},
		{"#### Heading 4", 4},
		{"##### Heading 5", 5},
		{"###### Heading 6", 6},
		{"####### Not a heading", 0},
		{"Not a heading", 0},
		{"", 0},
		{"#No space", 0},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := headingLevel(tt.line)
			if got != tt.want {
				t.Errorf("headingLevel(%q) = %d, want %d", tt.line, got, tt.want)
			}
		})
	}
}

func TestInsertAfterSection(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		section   string
		insertion string
		want      string
	}{
		{
			name:      "insert after h2 section",
			content:   "# NPCs\n\n## Gandalf\nA wizard.\n\n## Saruman\nAnother wizard.",
			section:   "Gandalf",
			insertion: "\n![Gandalf](assets/gandalf.png)\n",
			want:      "# NPCs\n\n## Gandalf\nA wizard.\n\n\n![Gandalf](assets/gandalf.png)\n\n## Saruman\nAnother wizard.",
		},
		{
			name:      "insert at end if last section",
			content:   "# NPCs\n\n## Gandalf\nA wizard.",
			section:   "Gandalf",
			insertion: "\n![Gandalf](assets/gandalf.png)\n",
			want:      "# NPCs\n\n## Gandalf\nA wizard.\n\n![Gandalf](assets/gandalf.png)\n",
		},
		{
			name:      "section not found returns original",
			content:   "# NPCs\n\n## Gandalf\nA wizard.",
			section:   "Missing",
			insertion: "\n![Missing](assets/missing.png)\n",
			want:      "# NPCs\n\n## Gandalf\nA wizard.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := insertAfterSection(tt.content, tt.section, tt.insertion)
			if got != tt.want {
				t.Errorf("insertAfterSection() = %q, want %q", got, tt.want)
			}
		})
	}
}
