package services

import (
	"context"
	"os"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

// mockPlayerMapRepo implements PlayerMapRepository for testing.
type mockPlayerMapRepo struct {
	maps map[string]*domain.PlayerMap
}

func newMockPlayerMapRepo() *mockPlayerMapRepo {
	return &mockPlayerMapRepo{maps: make(map[string]*domain.PlayerMap)}
}

func (r *mockPlayerMapRepo) Create(ctx context.Context, campaignID string, pm *domain.PlayerMap) error {
	r.maps[pm.ID] = pm
	return nil
}

func (r *mockPlayerMapRepo) Read(ctx context.Context, campaignID string, mapID string) (*domain.PlayerMap, error) {
	pm, ok := r.maps[mapID]
	if !ok {
		return nil, os.ErrNotExist
	}
	return pm, nil
}

func (r *mockPlayerMapRepo) Update(ctx context.Context, campaignID string, pm *domain.PlayerMap) error {
	r.maps[pm.ID] = pm
	return nil
}

func (r *mockPlayerMapRepo) Delete(ctx context.Context, campaignID string, mapID string) error {
	delete(r.maps, mapID)
	return nil
}

func (r *mockPlayerMapRepo) GetByArea(ctx context.Context, campaignID string, areaID string) ([]*domain.PlayerMap, error) {
	var result []*domain.PlayerMap
	for _, pm := range r.maps {
		if pm.AreaID == areaID {
			result = append(result, pm)
		}
	}
	return result, nil
}

func TestPlayerMapService_GeneratePlayerVariant(t *testing.T) {
	repo := newMockPlayerMapRepo()
	svc := NewPlayerMapService(repo)

	// Create a source DM map with secret features
	dmMap := &domain.PlayerMap{
		ID:          "dm_map_1",
		SourceMapID: "dm_map_1",
		CampaignID:  "campaign-1",
		AreaID:      "area_1",
		Title:       "DM Map",
		Description: "DM map with secrets",
		IncludedFeatures: []domain.MapFeature{
			{ID: "f1", Type: "room", Label: "Entry", IsSecret: false},
			{ID: "f2", Type: "trap", Label: "Hidden Pit", IsSecret: true},
			{ID: "f3", Type: "treasure", Label: "Secret Chest", IsSecret: true},
			{ID: "f4", Type: "door", Label: "Exit", IsSecret: false},
		},
		ExportFormats: []string{"svg", "pdf", "png"},
		GridVisible:   true,
		Scale:         "1 square = 5 feet",
	}

	// Save DM map
	if err := repo.Create(context.Background(), "campaign-1", dmMap); err != nil {
		t.Fatalf("failed to seed DM map: %v", err)
	}

	// Generate player variant
	playerMap, err := svc.GeneratePlayerVariant(context.Background(), "campaign-1", "dm_map_1", "area_2")
	if err != nil {
		t.Fatalf("GeneratePlayerVariant() error: %v", err)
	}

	// Check that secret features are excluded
	for _, f := range playerMap.IncludedFeatures {
		if f.IsSecret {
			t.Errorf("player map includes secret feature: %s", f.ID)
		}
	}

	// Check excluded features list
	if len(playerMap.ExcludedFeatures) != 2 {
		t.Errorf("expected 2 excluded features, got %d", len(playerMap.ExcludedFeatures))
	}

	// Check area ID was overridden
	if playerMap.AreaID != "area_2" {
		t.Errorf("AreaID = %s, want area_2", playerMap.AreaID)
	}

	// Check that the player map was persisted
	persisted, err := repo.Read(context.Background(), "campaign-1", playerMap.ID)
	if err != nil {
		t.Fatalf("persisted player map not found: %v", err)
	}
	if persisted == nil {
		t.Fatal("persisted player map is nil")
	}
}

func TestPlayerMapService_GeneratePlayerVariant_SourceNotFound(t *testing.T) {
	repo := newMockPlayerMapRepo()
	svc := NewPlayerMapService(repo)

	_, err := svc.GeneratePlayerVariant(context.Background(), "campaign-1", "nonexistent", "area_1")
	if err == nil {
		t.Error("expected error for nonexistent source map, got nil")
	}
}

func TestPlayerMapService_RedactSecretFeatures(t *testing.T) {
	repo := newMockPlayerMapRepo()
	svc := NewPlayerMapService(repo)

	dmMap := &domain.PlayerMap{
		ID:          "dm_map_2",
		SourceMapID: "dm_map_2",
		CampaignID:  "campaign-1",
		AreaID:      "area_1",
		Title:       "DM Map",
		Description: "Map with secrets",
		IncludedFeatures: []domain.MapFeature{
			{ID: "pub", Type: "room", Label: "Public Room", IsSecret: false},
			{ID: "secret_1", Type: "treasure", Label: "Hidden Cache", IsSecret: true},
		},
		ExportFormats: []string{"svg"},
		GridVisible:   true,
	}

	redacted, err := svc.RedactSecretFeatures(context.Background(), dmMap)
	if err != nil {
		t.Fatalf("RedactSecretFeatures() error: %v", err)
	}

	for _, f := range redacted.IncludedFeatures {
		if f.IsSecret {
			t.Errorf("redacted map still has secret feature: %s", f.ID)
		}
	}
	if len(redacted.ExcludedFeatures) != 1 || redacted.ExcludedFeatures[0] != "secret_1" {
		t.Errorf("ExcludedFeatures = %v, want [secret_1]", redacted.ExcludedFeatures)
	}
}

func TestPlayerMapService_RedactSecretFeatures_NilInput(t *testing.T) {
	repo := newMockPlayerMapRepo()
	svc := NewPlayerMapService(repo)

	_, err := svc.RedactSecretFeatures(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil dmMap, got nil")
	}
}
