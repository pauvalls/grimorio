package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func TestSelectToneTemplate(t *testing.T) {
	templates := map[string]string{
		"grim":   "Dark text %s",
		"heroic": "Heroic text %s",
	}

	tests := []struct {
		name string
		tone string
		want string
	}{
		{"grim tone", "grim", "Dark text %s"},
		{"heroic tone", "heroic", "Heroic text %s"},
		{"unknown tone falls back to default", "unknown", "default fallback"},
		{"empty tone falls back to default", "", "default fallback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectToneTemplate(tt.tone, templates, "default fallback")
			if got != tt.want {
				t.Errorf("selectToneTemplate(%q) = %q, want %q", tt.tone, got, tt.want)
			}
		})
	}
}

func TestPickCampaignName(t *testing.T) {
	tests := []struct {
		name         string
		loreContent  string
		introContent string
		campaignID   string
		want         string
	}{
		{
			name:         "from intro heading",
			loreContent:  "",
			introContent: "# My Campaign\n\nDescription",
			campaignID:   "my-campaign",
			want:         "My Campaign",
		},
		{
			name:         "from lore heading when intro empty",
			loreContent:  "# Lore World\n\nHistory...",
			introContent: "",
			campaignID:   "my-campaign",
			want:         "Lore World",
		},
		{
			name:         "fallback to campaign ID when no headings",
			loreContent:  "just some text",
			introContent: "more text",
			campaignID:   "my-campaign",
			want:         "my-campaign",
		},
		{
			name:         "fallback when both empty",
			loreContent:  "",
			introContent: "",
			campaignID:   "my-campaign",
			want:         "my-campaign",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pickCampaignName(tt.loreContent, tt.introContent, tt.campaignID)
			if got != tt.want {
				t.Errorf("pickCampaignName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPickCanonReference(t *testing.T) {
	tests := []struct {
		name     string
		entities []domain.CanonEntity
		wantSkip bool // true if we expect any non-empty result
	}{
		{
			name:     "empty entities returns empty",
			entities: []domain.CanonEntity{},
			wantSkip: false,
		},
		{
			name: "prefers mcguffin and villain",
			entities: []domain.CanonEntity{
				{Name: "Hero", Role: "ally"},
				{Name: "Evil Lord", Role: "villain"},
				{Name: "Amulet", Role: "mcguffin"},
				{Name: "Villager", Role: ""},
			},
			wantSkip: true,
		},
		{
			name: "falls back to any entity with name",
			entities: []domain.CanonEntity{
				{Name: "Town", Role: "location"},
				{Name: "Shopkeeper", Role: "merchant"},
			},
			wantSkip: true,
		},
		{
			name: "entity with empty name is skipped",
			entities: []domain.CanonEntity{
				{Name: "", Role: "villain"},
			},
			wantSkip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pickCanonReference(tt.entities)
			if tt.wantSkip && got == "" {
				t.Errorf("pickCanonReference() = %q, expected non-empty", got)
			}
			if !tt.wantSkip && got != "" {
				t.Errorf("pickCanonReference() = %q, expected empty", got)
			}
		})
	}
}

func TestBuildContextText(t *testing.T) {
	tests := []struct {
		name           string
		settingDesc    string
		loreContent    string
		introContent   string
		campaignName   string
		wantNonEmpty   bool
		wantContains   []string
	}{
		{
			name:         "with setting and lore",
			settingDesc:  "A dark fantasy world",
			loreContent:  "## History\n\nThe ancient kingdom fell centuries ago when the Dark Lord rose to power. The land is still scarred by that war.",
			introContent: "",
			campaignName: "Shadow Realm",
			wantNonEmpty: true,
			wantContains: []string{"A dark fantasy world", "ancient kingdom fell", "Shadow Realm"},
		},
		{
			name:         "empty input uses fallback",
			settingDesc:  "",
			loreContent:  "",
			introContent: "",
			campaignName: "",
			wantNonEmpty: true,
			wantContains: []string{"prólogo"},
		},
		{
			name:         "with setting only",
			settingDesc:  "A mystical forest",
			loreContent:  "",
			introContent: "",
			campaignName: "Elven Realm",
			wantNonEmpty: true,
			wantContains: []string{"A mystical forest", "Elven Realm"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildContextText(tt.settingDesc, tt.loreContent, tt.introContent, tt.campaignName)
			if tt.wantNonEmpty && got == "" {
				t.Errorf("buildContextText() returned empty, expected non-empty")
			}
			for _, contains := range tt.wantContains {
				if !containsString(got, contains) {
					t.Errorf("buildContextText() should contain %q, got: %s", contains, got)
				}
			}
		})
	}
}

func TestBuildConnectionsText(t *testing.T) {
	tests := []struct {
		name         string
		entities     []domain.CanonEntity
		campaignName string
		wantNonEmpty bool
		wantContains []string
	}{
		{
			name:         "empty entities uses fallback",
			entities:     []domain.CanonEntity{},
			campaignName: "Test",
			wantNonEmpty: true,
			wantContains: []string{"personas", "lugares"},
		},
		{
			name: "with villain and allies",
			entities: []domain.CanonEntity{
				{Name: "Dark Lord", Role: "villain"},
				{Name: "Wise Wizard", Role: "ally"},
				{Name: "Crystal of Light", Role: "mcguffin"},
			},
			campaignName: "Test",
			wantNonEmpty: true,
			wantContains: []string{"Dark Lord", "Wise Wizard", "Crystal of Light", "Antagonistas", "Aliados", "Objetos de Interés"},
		},
		{
			name: "entities without roles",
			entities: []domain.CanonEntity{
				{Name: "Forest", Role: "location"},
				{Name: "Merchant", Role: ""},
			},
			campaignName: "Test",
			wantNonEmpty: true,
			wantContains: []string{"Forest", "Merchant", "Otras Entidades"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildConnectionsText(tt.entities, tt.campaignName)
			if tt.wantNonEmpty && got == "" {
				t.Errorf("buildConnectionsText() returned empty, expected non-empty")
			}
			for _, contains := range tt.wantContains {
				if !containsString(got, contains) {
					t.Errorf("buildConnectionsText() should contain %q, got: %s", contains, got)
				}
			}
		})
	}
}

func TestPrologueService_GeneratePrologue_FullCanon(t *testing.T) {
	// Setup temp directory with campaign files
	tmpDir := t.TempDir()
	campaignDir := filepath.Join(tmpDir, "full-canon-test")
	if err := os.MkdirAll(campaignDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write lore.md
	loreContent := "# World of Eldoria\n\n## History\nThe ancient kingdom of Eldoria was once a beacon of hope. Now shadows gather at its borders."
	if err := os.WriteFile(filepath.Join(campaignDir, "lore.md"), []byte(loreContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Write introduction.md
	introContent := "# Eldoria Reborn\n\nA heroic fantasy campaign set in a world recovering from a great war. The players must unite the fractured kingdoms."
	if err := os.WriteFile(filepath.Join(campaignDir, "introduction.md"), []byte(introContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Setup canon repo with entities
	canonRepo := repository.NewMemoryCanonRepository()
	canonDoc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "full-canon-test",
		Entities: []domain.CanonEntity{
			{Name: "Dark Lord Malachar", Role: "villain", Type: domain.EntityTypeNPC},
			{Name: "Crystal of Eternity", Role: "mcguffin", Type: domain.EntityTypeItem},
			{Name: "Lady Seraphina", Role: "ally", Type: domain.EntityTypeNPC},
			{Name: "Kingdom of Eldoria", Role: "location", Type: domain.EntityTypeLocation},
		},
	}
	if err := canonRepo.Save("full-canon-test", canonDoc); err != nil {
		t.Fatal(err)
	}

	service := NewPrologueService(tmpDir, canonRepo)
	prologue, warnings, err := service.GeneratePrologue(context.Background(), "full-canon-test", "heroic", nil)
	if err != nil {
		t.Fatalf("GeneratePrologue() error: %v", err)
	}

	if !prologue.Validate() {
		t.Errorf("GeneratePrologue() produced invalid prologue: %+v", prologue)
	}

	if prologue.Tone != "heroic" {
		t.Errorf("GeneratePrologue() tone = %q, want %q", prologue.Tone, "heroic")
	}

	if prologue.CampaignID != "full-canon-test" {
		t.Errorf("GeneratePrologue() CampaignID = %q, want %q", prologue.CampaignID, "full-canon-test")
	}

	// Part 1 should reference an entity (villain or mcguffin)
	hasEntityRef := containsString(prologue.Parts[0].Content, "Malachar") ||
		containsString(prologue.Parts[0].Content, "Eternity") ||
		containsString(prologue.Parts[0].Content, "Seraphina")
	if !hasEntityRef {
		t.Errorf("Part 1 (Hook) should reference a campaign entity, got: %s", prologue.Parts[0].Content)
	}

	_ = warnings // warnings are optional
}

func TestPrologueService_GeneratePrologue_MinimalCanon(t *testing.T) {
	tmpDir := t.TempDir()
	campaignDir := filepath.Join(tmpDir, "minimal-canon")
	if err := os.MkdirAll(campaignDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write minimal lore (just a heading)
	if err := os.WriteFile(filepath.Join(campaignDir, "lore.md"), []byte("# Minimal World"), 0644); err != nil {
		t.Fatal(err)
	}

	// No introduction file

	// Setup canon repo without entities
	canonRepo := repository.NewMemoryCanonRepository()
	canonDoc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "minimal-canon",
		Entities:      []domain.CanonEntity{},
	}
	if err := canonRepo.Save("minimal-canon", canonDoc); err != nil {
		t.Fatal(err)
	}

	service := NewPrologueService(tmpDir, canonRepo)
	prologue, warnings, err := service.GeneratePrologue(context.Background(), "minimal-canon", "", nil)
	if err != nil {
		t.Fatalf("GeneratePrologue() error: %v", err)
	}

	if !prologue.Validate() {
		t.Errorf("GeneratePrologue() produced invalid prologue")
	}

	// With empty tone and no entities, should have warning about no villain/mcguffin
	hasVillainWarning := false
	for _, w := range warnings {
		if containsString(w, "villain") || containsString(w, "mcguffin") || containsString(w, "generic") {
			hasVillainWarning = true
		}
	}
	if !hasVillainWarning {
		t.Errorf("Expected warning about missing villain/mcguffin, got warnings: %v", warnings)
	}

	// Default tone should be "heroic"
	if prologue.Tone != "heroic" {
		t.Errorf("GeneratePrologue() tone = %q, want %q", prologue.Tone, "heroic")
	}
}

func TestPrologueService_GeneratePrologue_NonexistentCampaign(t *testing.T) {
	canonRepo := repository.NewMemoryCanonRepository()
	service := NewPrologueService("/tmp/nonexistent", canonRepo)

	_, _, err := service.GeneratePrologue(context.Background(), "nonexistent-campaign", "heroic", nil)
	if err == nil {
		t.Errorf("GeneratePrologue() expected error for nonexistent campaign")
	}
	if err != nil && !containsString(err.Error(), "campaign not found") {
		t.Errorf("GeneratePrologue() error = %q, expected 'campaign not found'", err.Error())
	}
}

func TestPrologueService_GeneratePrologue_GrimTone(t *testing.T) {
	tmpDir := t.TempDir()
	campaignDir := filepath.Join(tmpDir, "grim-test")
	if err := os.MkdirAll(campaignDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(campaignDir, "lore.md"), []byte("# Grim World\n\nDarkness falls."), 0644); err != nil {
		t.Fatal(err)
	}

	canonRepo := repository.NewMemoryCanonRepository()
	canonDoc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "grim-test",
		Entities: []domain.CanonEntity{
			{Name: "Shadow Beast", Role: "villain", Type: domain.EntityTypeNPC},
		},
	}
	if err := canonRepo.Save("grim-test", canonDoc); err != nil {
		t.Fatal(err)
	}

	service := NewPrologueService(tmpDir, canonRepo)
	prologue, _, err := service.GeneratePrologue(context.Background(), "grim-test", "grim", nil)
	if err != nil {
		t.Fatalf("GeneratePrologue() error: %v", err)
	}

	if !prologue.Validate() {
		t.Errorf("GeneratePrologue() produced invalid prologue")
	}

	if prologue.Tone != "grim" {
		t.Errorf("GeneratePrologue() tone = %q, want %q", prologue.Tone, "grim")
	}

	// Part 1 should contain grim tone phrasing
	if !containsString(prologue.Parts[0].Content, "sombras") {
		t.Errorf("Part 1 (Hook) should contain grim tone language, got: %s", prologue.Parts[0].Content)
	}

	// Part 4 should contain grim tone phrasing
	if !containsString(prologue.Parts[3].Content, "oscuridad") && !containsString(prologue.Parts[3].Content, "sombras") {
		t.Errorf("Part 4 (RoadAhead) should contain grim tone language, got: %s", prologue.Parts[3].Content)
	}
}

func TestPrologueService_GeneratePrologue_HorrorTone(t *testing.T) {
	tmpDir := t.TempDir()
	campaignDir := filepath.Join(tmpDir, "horror-test")
	if err := os.MkdirAll(campaignDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(campaignDir, "lore.md"), []byte("# Horror Realm"), 0644); err != nil {
		t.Fatal(err)
	}

	canonRepo := repository.NewMemoryCanonRepository()
	canonDoc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "horror-test",
		Entities:      []domain.CanonEntity{
			{Name: "The Ancient One", Role: "villain", Type: domain.EntityTypeNPC},
		},
	}
	if err := canonRepo.Save("horror-test", canonDoc); err != nil {
		t.Fatal(err)
	}

	service := NewPrologueService(tmpDir, canonRepo)
	prologue, _, err := service.GeneratePrologue(context.Background(), "horror-test", "horror", nil)
	if err != nil {
		t.Fatalf("GeneratePrologue() error: %v", err)
	}

	if prologue.Tone != "horror" {
		t.Errorf("GeneratePrologue() tone = %q, want %q", prologue.Tone, "horror")
	}

	// Part 1 should contain horror tone phrasing
	if !containsString(prologue.Parts[0].Content, "terrible") && !containsString(prologue.Parts[0].Content, "pesadillas") {
		t.Errorf("Part 1 (Hook) should contain horror tone language, got: %s", prologue.Parts[0].Content)
	}

	// Part 4 should contain horror tone phrasing
	if !containsString(prologue.Parts[3].Content, "oscuridad") {
		t.Errorf("Part 4 (RoadAhead) should contain horror tone language, got: %s", prologue.Parts[3].Content)
	}
}

func TestPrologueService_GeneratePrologue_WithCharacterHooks(t *testing.T) {
	tmpDir := t.TempDir()
	campaignDir := filepath.Join(tmpDir, "hooks-test")
	if err := os.MkdirAll(campaignDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(campaignDir, "lore.md"), []byte("# Hooks World"), 0644); err != nil {
		t.Fatal(err)
	}

	canonRepo := repository.NewMemoryCanonRepository()
	canonDoc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "hooks-test",
		Entities: []domain.CanonEntity{
			{Name: "Big Bad", Role: "villain", Type: domain.EntityTypeNPC},
		},
	}
	if err := canonRepo.Save("hooks-test", canonDoc); err != nil {
		t.Fatal(err)
	}

	service := NewPrologueService(tmpDir, canonRepo)
	hooks := []string{"Aragorn seeks his destiny", "Legolas hears the call"}
	prologue, _, err := service.GeneratePrologue(context.Background(), "hooks-test", "heroic", hooks)
	if err != nil {
		t.Fatalf("GeneratePrologue() error: %v", err)
	}

	if !prologue.Validate() {
		t.Errorf("GeneratePrologue() produced invalid prologue")
	}

	if len(prologue.Parts) != 4 {
		t.Errorf("GeneratePrologue() produced %d parts, want 4", len(prologue.Parts))
	}
}

func TestPrologueService_GeneratePrologue_PartTitles(t *testing.T) {
	tmpDir := t.TempDir()
	campaignDir := filepath.Join(tmpDir, "titles-test")
	if err := os.MkdirAll(campaignDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(campaignDir, "lore.md"), []byte("# Title World"), 0644); err != nil {
		t.Fatal(err)
	}

	canonRepo := repository.NewMemoryCanonRepository()
	canonDoc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "titles-test",
		Entities: []domain.CanonEntity{
			{Name: "Villain", Role: "villain", Type: domain.EntityTypeNPC},
		},
	}
	if err := canonRepo.Save("titles-test", canonDoc); err != nil {
		t.Fatal(err)
	}

	service := NewPrologueService(tmpDir, canonRepo)
	prologue, _, err := service.GeneratePrologue(context.Background(), "titles-test", "heroic", nil)
	if err != nil {
		t.Fatalf("GeneratePrologue() error: %v", err)
	}

	expectedTitles := []string{
		"Gancho Narrativo",
		"Trasfondo",
		"Conexiones",
		"El Camino por Delante",
	}
	for i, expected := range expectedTitles {
		if prologue.Parts[i].Title != expected {
			t.Errorf("Part %d title = %q, want %q", i+1, prologue.Parts[i].Title, expected)
		}
	}
}

// Helper function for contains string check
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && containsStringSlow(s, substr)
}

func containsStringSlow(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
