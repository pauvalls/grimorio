package services

import (
	"context"
	"testing"
	"github.com/pauvalls/grimorio/internal/domain"
)

// MockMagicItemRepository for testing
type MockMagicItemRepository struct {
	items []*domain.MagicItem
	err   error
}

func (m *MockMagicItemRepository) Create(ctx context.Context, campaignID string, item *domain.MagicItem) error {
	return nil
}

func (m *MockMagicItemRepository) Read(ctx context.Context, campaignID string, itemID string) (*domain.MagicItem, error) {
	return nil, nil
}

func (m *MockMagicItemRepository) Update(ctx context.Context, campaignID string, item *domain.MagicItem) error {
	return nil
}

func (m *MockMagicItemRepository) Delete(ctx context.Context, campaignID string, itemID string) error {
	return nil
}

func (m *MockMagicItemRepository) GetByRarity(ctx context.Context, campaignID string, rarity domain.MagicItemRarity) ([]*domain.MagicItem, error) {
	return m.items, m.err
}

func (m *MockMagicItemRepository) GetByChapter(ctx context.Context, campaignID string, chapterID string) ([]*domain.MagicItem, error) {
	return m.items, m.err
}

func TestItemService_GenerateItem_ValidRarity(t *testing.T) {
	repo := &MockMagicItemRepository{}
	service := NewItemService(repo)

	item, err := service.GenerateItem(context.Background(), "campaign_1", domain.RarityRare, domain.ItemTypeWeapon)
	if err != nil {
		t.Fatalf("GenerateItem() error = %v", err)
	}

	if item.Rarity != domain.RarityRare {
		t.Errorf("Rarity = %s, want rare", item.Rarity)
	}
	if item.Type != domain.ItemTypeWeapon {
		t.Errorf("Type = %s, want weapon", item.Type)
	}
	if err := item.Validate(); err != nil {
		t.Errorf("Generated item invalid: %v", err)
	}
}

func TestItemService_GenerateItem_InvalidRarity(t *testing.T) {
	repo := &MockMagicItemRepository{}
	service := NewItemService(repo)

	_, err := service.GenerateItem(context.Background(), "campaign_1", "invalid", domain.ItemTypeWeapon)
	if err == nil {
		t.Error("Expected error for invalid rarity, got nil")
	}
}

func TestItemService_GenerateCursedItem_HasRemovalMethod(t *testing.T) {
	repo := &MockMagicItemRepository{}
	service := NewItemService(repo)

	item, err := service.GenerateCursedItem(context.Background(), "campaign_1", domain.RarityUncommon)
	if err != nil {
		t.Fatalf("GenerateCursedItem() error = %v", err)
	}

	if item.Curse == nil {
		t.Fatal("Expected curse, got nil")
	}
	if item.Curse.RemovalMethod == "" {
		t.Error("Cursed item must have removal method")
	}
}

func TestItemService_ValidateItemPower_AppropriateForLevel(t *testing.T) {
	repo := &MockMagicItemRepository{}
	service := NewItemService(repo)

	item := &domain.MagicItem{
		ID:       "test_item",
		Name:     "Test Sword",
		Type:     domain.ItemTypeWeapon,
		Rarity:   domain.RarityRare,
		Properties: []domain.MagicItemProperty{{Bonus: 2}},
		Lore:     "A test item",
	}

	valid, err := service.ValidateItemPower(context.Background(), item, 7)
	if err != nil {
		t.Fatalf("ValidateItemPower() error = %v", err)
	}
	if !valid {
		t.Error("Expected item to be appropriate for level 7")
	}
}

func TestItemService_ValidateItemPower_TooPowerful(t *testing.T) {
	repo := &MockMagicItemRepository{}
	service := NewItemService(repo)

	item := &domain.MagicItem{
		ID:       "test_item",
		Name:     "Legendary Sword",
		Type:     domain.ItemTypeWeapon,
		Rarity:   domain.RarityLegendary,
		Properties: []domain.MagicItemProperty{{Bonus: 3}},
		Lore:     "A legendary item",
	}

	valid, err := service.ValidateItemPower(context.Background(), item, 3)
	if err != nil {
		t.Fatalf("ValidateItemPower() error = %v", err)
	}
	if valid {
		t.Error("Expected legendary item to be too powerful for level 3")
	}
}
