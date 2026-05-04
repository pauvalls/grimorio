package svg

import (
	"strings"
	"testing"
)

func TestDefaultBattleMapConfig(t *testing.T) {
	cfg := DefaultBattleMapConfig()
	if cfg.Width != 800 {
		t.Errorf("expected Width 800, got %d", cfg.Width)
	}
	if cfg.Height != 600 {
		t.Errorf("expected Height 600, got %d", cfg.Height)
	}
	if cfg.GridSize != 40 {
		t.Errorf("expected GridSize 40, got %d", cfg.GridSize)
	}
	if cfg.NumRooms != 6 {
		t.Errorf("expected NumRooms 6, got %d", cfg.NumRooms)
	}
	if cfg.Style != MapStyleDungeon {
		t.Errorf("expected Style dungeon, got %s", cfg.Style)
	}
}

func TestGenerateBattleMap(t *testing.T) {
	cfg := BattleMapConfig{
		Width:       400,
		Height:      300,
		GridSize:    40,
		NumRooms:    3,
		MinRoomSize: 2,
		MaxRoomSize: 3,
		Style:       MapStyleDungeon,
		Seed:        42,
		Labels:      []string{"Entrance", "Boss Room"},
	}

	svg := GenerateBattleMap(cfg)
	if svg == "" {
		t.Error("GenerateBattleMap() returned empty string")
	}
	if !strings.HasPrefix(svg, "<svg") {
		t.Error("GenerateBattleMap() output doesn't start with <svg")
	}
	if !strings.Contains(svg, "Entrance") {
		t.Error("GenerateBattleMap() output doesn't contain label 'Entrance'")
	}
}

func TestGenerateBattleMap_NoSeed(t *testing.T) {
	cfg := BattleMapConfig{
		Width:       400,
		Height:      300,
		GridSize:    40,
		NumRooms:    2,
		MinRoomSize: 2,
		MaxRoomSize: 3,
		Style:       MapStyleDungeon,
		Seed:        0,
	}

	svg := GenerateBattleMap(cfg)
	if svg == "" {
		t.Error("GenerateBattleMap() returned empty string")
	}
}

func TestGenerateBattleMap_Landscape(t *testing.T) {
	cfg := BattleMapConfig{
		Width:       400,
		Height:      300,
		GridSize:    40,
		NumRooms:    2,
		MinRoomSize: 2,
		MaxRoomSize: 3,
		Style:       MapStyleLandscape,
		Seed:        42,
	}

	svg := GenerateBattleMap(cfg)
	if svg == "" {
		t.Error("GenerateBattleMap() returned empty string")
	}
	if !strings.Contains(svg, "stroke-dasharray") {
		t.Error("Landscape style should have dashed paths")
	}
}

func TestGenerateBattleMap_City(t *testing.T) {
	cfg := BattleMapConfig{
		Width:       400,
		Height:      300,
		GridSize:    40,
		NumRooms:    2,
		MinRoomSize: 2,
		MaxRoomSize: 3,
		Style:       MapStyleCity,
		Seed:        42,
	}

	svg := GenerateBattleMap(cfg)
	if svg == "" {
		t.Error("GenerateBattleMap() returned empty string")
	}
}

func TestGenerateDivider_Ornate(t *testing.T) {
	svg := GenerateDivider(600, "ornate")
	if svg == "" {
		t.Error("GenerateDivider() returned empty string")
	}
	if !strings.Contains(svg, "<svg") {
		t.Error("GenerateDivider() output doesn't start with <svg")
	}
	if !strings.Contains(svg, "path") {
		t.Error("Ornate divider should contain path element")
	}
}

func TestGenerateDivider_Simple(t *testing.T) {
	svg := GenerateDivider(400, "simple")
	if svg == "" {
		t.Error("GenerateDivider() returned empty string")
	}
	if !strings.Contains(svg, "<line") {
		t.Error("Simple divider should contain line element")
	}
}

func TestGenerateDivider_Double(t *testing.T) {
	svg := GenerateDivider(400, "double")
	if svg == "" {
		t.Error("GenerateDivider() returned empty string")
	}
	count := strings.Count(svg, "<line")
	if count != 2 {
		t.Errorf("Double divider should have 2 lines, got %d", count)
	}
}

func TestGenerateDivider_Default(t *testing.T) {
	svg := GenerateDivider(400, "unknown")
	if svg == "" {
		t.Error("GenerateDivider() returned empty string")
	}
	if !strings.Contains(svg, "<line") {
		t.Error("Default divider should contain line element")
	}
}

func TestGenerateStatBlockBorder(t *testing.T) {
	svg := GenerateStatBlockBorder(400, 300)
	if svg == "" {
		t.Error("GenerateStatBlockBorder() returned empty string")
	}
	if !strings.HasPrefix(svg, "<svg") {
		t.Error("GenerateStatBlockBorder() output doesn't start with <svg")
	}
	if !strings.Contains(svg, "rect") {
		t.Error("Stat block border should contain rect elements")
	}
}

func TestRoomDistance(t *testing.T) {
	a := Room{X: 0, Y: 0, W: 2, H: 2}
	b := Room{X: 5, Y: 0, W: 2, H: 2}

	dist := roomDistance(a, b)
	expected := 5.0 // (5+1) - (0+1) = 5
	if dist != expected {
		t.Errorf("expected distance %f, got %f", expected, dist)
	}
}

func TestCreateCorridor(t *testing.T) {
	from := Room{X: 0, Y: 0, W: 2, H: 2}
	to := Room{X: 5, Y: 0, W: 2, H: 2}

	corr := createCorridor(from, to)
	if corr.X1 != from.X+from.W {
		t.Errorf("expected X1 %d, got %d", from.X+from.W, corr.X1)
	}
	if corr.X2 != to.X {
		t.Errorf("expected X2 %d, got %d", to.X, corr.X2)
	}
}

func TestGenerateRooms(t *testing.T) {
	cfg := BattleMapConfig{
		Width:       400,
		Height:      300,
		GridSize:    40,
		NumRooms:    3,
		MinRoomSize: 2,
		MaxRoomSize: 3,
		Seed:        42,
	}

	cols := cfg.Width / cfg.GridSize
	rows := cfg.Height / cfg.GridSize
	rooms := generateRooms(cfg, cols, rows)

	if len(rooms) == 0 {
		t.Error("generateRooms() returned no rooms")
	}
	if len(rooms) > cfg.NumRooms {
		t.Errorf("generateRooms() returned too many rooms: %d > %d", len(rooms), cfg.NumRooms)
	}
}

func TestGenerateCorridors(t *testing.T) {
	rooms := []Room{
		{X: 1, Y: 1, W: 2, H: 2, ID: 0},
		{X: 5, Y: 1, W: 2, H: 2, ID: 1},
		{X: 9, Y: 1, W: 2, H: 2, ID: 2},
	}

	corridors := generateCorridors(rooms)
	if len(corridors) == 0 {
		t.Error("generateCorridors() returned no corridors")
	}
}

func TestGenerateCorridors_SingleRoom(t *testing.T) {
	rooms := []Room{
		{X: 1, Y: 1, W: 2, H: 2, ID: 0},
	}

	corridors := generateCorridors(rooms)
	if corridors != nil {
		t.Error("generateCorridors() with single room should return nil")
	}
}

func TestRenderRoomLabel(t *testing.T) {
	room := Room{X: 1, Y: 1, W: 2, H: 2, Label: "Test Room"}
	label := renderRoomLabel(room, 40)
	if label == "" {
		t.Error("renderRoomLabel() returned empty string for labeled room")
	}
	if !strings.Contains(label, "Test Room") {
		t.Error("renderRoomLabel() doesn't contain room label")
	}
}

func TestRenderRoomLabel_Empty(t *testing.T) {
	room := Room{X: 1, Y: 1, W: 2, H: 2, Label: ""}
	label := renderRoomLabel(room, 40)
	if label != "" {
		t.Error("renderRoomLabel() should return empty string for unlabeled room")
	}
}
