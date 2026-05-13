package svg

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

var rng = rand.New(rand.NewSource(rand.Int63()))

type MapStyle string

const (
	MapStyleDungeon   MapStyle = "dungeon"
	MapStyleLandscape MapStyle = "landscape"
	MapStyleCity      MapStyle = "city"
)

type BattleMapConfig struct {
	Width       int
	Height      int
	GridSize    int
	NumRooms    int
	MinRoomSize int
	MaxRoomSize int
	Style       MapStyle
	Title       string
	Seed        int64
	Labels      []string
}

func DefaultBattleMapConfig() BattleMapConfig {
	return BattleMapConfig{
		Width:       800,
		Height:      600,
		GridSize:    40,
		NumRooms:    6,
		MinRoomSize: 2,
		MaxRoomSize: 5,
		Style:       MapStyleDungeon,
		Seed:        0,
		Labels:      []string{},
	}
}

type Room struct {
	X, Y, W, H int
	Label      string
	ID         int
}

type Corridor struct {
	X1, Y1, X2, Y2 int
}

func GenerateBattleMap(cfg BattleMapConfig) string {
	if cfg.Seed != 0 {
		rng = rand.New(rand.NewSource(cfg.Seed))
	} else {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}

	cols := cfg.Width / cfg.GridSize
	rows := cfg.Height / cfg.GridSize

	rooms := generateRooms(cfg, cols, rows)
	corridors := generateCorridors(rooms)

	return renderSVG(cfg, rooms, corridors)
}

func generateRooms(cfg BattleMapConfig, cols, rows int) []Room {
	var rooms []Room
	attempts := 0
	maxAttempts := cfg.NumRooms * 100

	for len(rooms) < cfg.NumRooms && attempts < maxAttempts {
		attempts++
		w := cfg.MinRoomSize + rng.Intn(cfg.MaxRoomSize-cfg.MinRoomSize+1)
		h := cfg.MinRoomSize + rng.Intn(cfg.MaxRoomSize-cfg.MinRoomSize+1)
		x := rng.Intn(cols-w-2) + 1
		y := rng.Intn(rows-h-2) + 1

		room := Room{X: x, Y: y, W: w, H: h, ID: len(rooms)}

		// Check overlap - allow adjacent but not overlapping
		overlap := false
		for _, existing := range rooms {
			if x <= existing.X+existing.W && x+w >= existing.X &&
				y <= existing.Y+existing.H && y+h >= existing.Y {
				overlap = true
				break
			}
		}

		if !overlap {
			// Assign label
			if room.ID < len(cfg.Labels) && cfg.Labels[room.ID] != "" {
				room.Label = cfg.Labels[room.ID]
			} else {
				room.Label = fmt.Sprintf("Sala %d", room.ID+1)
			}
			rooms = append(rooms, room)
		}
	}

	return rooms
}

func generateCorridors(rooms []Room) []Corridor {
	if len(rooms) < 2 {
		return nil
	}

	var corridors []Corridor
	connected := make([]bool, len(rooms))
	connected[0] = true

	// Connect rooms using minimum spanning tree approach
	for connectedCount := 1; connectedCount < len(rooms); connectedCount++ {
		bestDist := math.MaxFloat64
		var bestFrom, bestTo int

		for i := 0; i < len(rooms); i++ {
			if !connected[i] {
				continue
			}
			for j := 0; j < len(rooms); j++ {
				if connected[j] {
					continue
				}
				dist := roomDistance(rooms[i], rooms[j])
				if dist < bestDist {
					bestDist = dist
					bestFrom = i
					bestTo = j
				}
			}
		}

		if bestFrom != bestTo {
			corridors = append(corridors, createCorridor(rooms[bestFrom], rooms[bestTo]))
			connected[bestTo] = true
		}
	}

	// Add some extra connections for loops (20% chance per pair)
	for i := 0; i < len(rooms); i++ {
		for j := i + 1; j < len(rooms); j++ {
			if rng.Float64() < 0.2 {
				corridors = append(corridors, createCorridor(rooms[i], rooms[j]))
			}
		}
	}

	return corridors
}

func roomDistance(a, b Room) float64 {
	cx1 := float64(a.X + a.W/2)
	cy1 := float64(a.Y + a.H/2)
	cx2 := float64(b.X + b.W/2)
	cy2 := float64(b.Y + b.H/2)
	return math.Abs(cx1-cx2) + math.Abs(cy1-cy2)
}

func createCorridor(from, to Room) Corridor {
	// Find closest edges
	var startX, startY, endX, endY int

	// Determine horizontal alignment
	if from.X+from.W <= to.X {
		// From is to the left of To
		startX = from.X + from.W
		endX = to.X
	} else if to.X+to.W <= from.X {
		// To is to the left of From
		startX = from.X
		endX = to.X + to.W
	} else {
		// Overlapping horizontally - use center
		startX = from.X + from.W/2
		endX = to.X + to.W/2
	}

	// Determine vertical alignment
	if from.Y+from.H <= to.Y {
		// From is above To
		startY = from.Y + from.H
		endY = to.Y
	} else if to.Y+to.H <= from.Y {
		// To is above From
		startY = from.Y
		endY = to.Y + to.H
	} else {
		// Overlapping vertically - use center
		startY = from.Y + from.H/2
		endY = to.Y + to.H/2
	}

	return Corridor{X1: startX, Y1: startY, X2: endX, Y2: endY}
}

func renderSVG(cfg BattleMapConfig, rooms []Room, corridors []Corridor) string {
	var sb strings.Builder
	gs := cfg.GridSize

	fmt.Fprintf(&sb, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`,
		cfg.Width, cfg.Height, cfg.Width, cfg.Height)

	// Definitions
	sb.WriteString(`<defs>`)

	// Grid pattern
	fmt.Fprintf(&sb, `<pattern id="grid" width="%d" height="%d" patternUnits="userSpaceOnUse">`, gs, gs)
	fmt.Fprintf(&sb, `<path d="M %d 0 L 0 0 0 %d" fill="none" stroke="#c9ad6a" stroke-width="0.5" opacity="0.3"/>`, gs, gs)
	sb.WriteString(`</pattern>`)

	// Shadow filter for rooms
	sb.WriteString(`<filter id="shadow" x="-20%" y="-20%" width="140%" height="140%">`)
	sb.WriteString(`<feDropShadow dx="2" dy="2" stdDeviation="2" flood-color="#000" flood-opacity="0.3"/>`)
	sb.WriteString(`</filter>`)

	sb.WriteString(`</defs>`)

	// Background
	fmt.Fprintf(&sb, `<rect width="%d" height="%d" fill="#2c1e14"/>`, cfg.Width, cfg.Height)
	fmt.Fprintf(&sb, `<rect width="%d" height="%d" fill="url(#grid)"/>`, cfg.Width, cfg.Height)

	// Render corridors first (so they appear under rooms)
	for _, corr := range corridors {
		sb.WriteString(renderCorridor(corr, cfg, gs))
	}

	// Render rooms
	for _, room := range rooms {
		sb.WriteString(renderRoom(room, cfg, gs))
	}

	// Render room labels
	for _, room := range rooms {
		sb.WriteString(renderRoomLabel(room, gs))
	}

	// Title
	if cfg.Title != "" {
		fmt.Fprintf(&sb, `<text x="%d" y="%d" font-family="Arial, Helvetica, sans-serif" font-size="22" fill="#c9ad6a" text-anchor="middle" font-weight="bold">%s</text>`,
			cfg.Width/2, 30, cfg.Title)
	}

	// Room count info
	if len(rooms) > 0 {
		fmt.Fprintf(&sb, `<text x="%d" y="%d" font-family="Arial, sans-serif" font-size="11" fill="#c9ad6a" text-anchor="end" opacity="0.7">%d zonas</text>`,
			cfg.Width-10, cfg.Height-10, len(rooms))
	}

	sb.WriteString(`</svg>`)
	return sb.String()
}

func renderCorridor(corr Corridor, cfg BattleMapConfig, gs int) string {
	var sb strings.Builder

	x1 := corr.X1 * gs
	y1 := corr.Y1 * gs
	x2 := corr.X2 * gs
	y2 := corr.Y2 * gs

	// Corridor width
	cw := gs / 2
	if cw < 8 {
		cw = 8
	}

	switch cfg.Style {
	case MapStyleDungeon:
		// Draw corridor with slight offset for 3D effect
		fmt.Fprintf(&sb, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#1a1a1a" stroke-width="%d" opacity="0.5"/>`,
			x1+2, y1+2, x2+2, y2+2, cw+4)
		fmt.Fprintf(&sb, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#5a3d2b" stroke-width="%d"/>`,
			x1, y1, x2, y2, cw)
		fmt.Fprintf(&sb, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#8b7355" stroke-width="%d" opacity="0.6"/>`,
			x1, y1, x2, y2, cw-4)

		// Doorways at room connections
		fmt.Fprintf(&sb, `<circle cx="%d" cy="%d" r="%d" fill="#3d2e1f" stroke="#c9ad6a" stroke-width="1"/>`,
			x1, y1, cw/2+2)
		fmt.Fprintf(&sb, `<circle cx="%d" cy="%d" r="%d" fill="#3d2e1f" stroke="#c9ad6a" stroke-width="1"/>`,
			x2, y2, cw/2+2)

	case MapStyleLandscape:
		// Natural paths
		fmt.Fprintf(&sb, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#3d5a37" stroke-width="%d" stroke-dasharray="5,3" opacity="0.8"/>`,
			x1, y1, x2, y2, cw)

		// Path markers
		fmt.Fprintf(&sb, `<circle cx="%d" cy="%d" r="4" fill="#2d4a27"/>`, x1, y1)
		fmt.Fprintf(&sb, `<circle cx="%d" cy="%d" r="4" fill="#2d4a27"/>`, x2, y2)

	case MapStyleCity:
		// Streets
		fmt.Fprintf(&sb, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#8b7355" stroke-width="%d"/>`,
			x1, y1, x2, y2, cw+2)
		fmt.Fprintf(&sb, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#d4c5a9" stroke-width="%d"/>`,
			x1, y1, x2, y2, cw-2)

		// Street intersections
		fmt.Fprintf(&sb, `<circle cx="%d" cy="%d" r="5" fill="#c9ad6a"/>`, x1, y1)
		fmt.Fprintf(&sb, `<circle cx="%d" cy="%d" r="5" fill="#c9ad6a"/>`, x2, y2)
	}

	return sb.String()
}

func renderRoom(room Room, cfg BattleMapConfig, gs int) string {
	var sb strings.Builder

	x := room.X * gs
	y := room.Y * gs
	w := room.W * gs
	h := room.H * gs

	// Room shadow
	fmt.Fprintf(&sb, `<rect x="%d" y="%d" width="%d" height="%d" fill="#1a1a1a" opacity="0.3" rx="3"/>`,
		x+3, y+3, w, h)

	switch cfg.Style {
	case MapStyleDungeon:
		// Main room
		fmt.Fprintf(&sb, `<rect x="%d" y="%d" width="%d" height="%d" fill="#3d2e1f" stroke="#c9ad6a" stroke-width="2" rx="3" filter="url(#shadow)"/>`,
			x, y, w, h)

		// Inner floor detail
		fmt.Fprintf(&sb, `<rect x="%d" y="%d" width="%d" height="%d" fill="#4a3a2a" stroke="#8b7355" stroke-width="1" rx="2" opacity="0.5"/>`,
			x+6, y+6, w-12, h-12)

		// Corner accents
		fmt.Fprintf(&sb, `<circle cx="%d" cy="%d" r="3" fill="#8b7355"/>`, x+8, y+8)
		fmt.Fprintf(&sb, `<circle cx="%d" cy="%d" r="3" fill="#8b7355"/>`, x+w-8, y+8)
		fmt.Fprintf(&sb, `<circle cx="%d" cy="%d" r="3" fill="#8b7355"/>`, x+8, y+h-8)
		fmt.Fprintf(&sb, `<circle cx="%d" cy="%d" r="3" fill="#8b7355"/>`, x+w-8, y+h-8)

	case MapStyleLandscape:
		// Natural area
		fmt.Fprintf(&sb, `<rect x="%d" y="%d" width="%d" height="%d" fill="#4a6741" stroke="#2d4a27" stroke-width="2" rx="8" filter="url(#shadow)"/>`,
			x, y, w, h)

		// Trees/bushes
		numTrees := rng.Intn(3) + 1
		for i := 0; i < numTrees; i++ {
			tx := x + rng.Intn(w-20) + 10
			ty := y + rng.Intn(h-20) + 10
			fmt.Fprintf(&sb, `<circle cx="%d" cy="%d" r="8" fill="#2d4a27" opacity="0.7"/>`, tx, ty)
			fmt.Fprintf(&sb, `<circle cx="%d" cy="%d" r="5" fill="#3d5a37"/>`, tx, ty-3)
		}

		// Rocks
		numRocks := rng.Intn(2) + 1
		for i := 0; i < numRocks; i++ {
			rx := x + rng.Intn(w-15) + 8
			ry := y + rng.Intn(h-15) + 8
			fmt.Fprintf(&sb, `<polygon points="%d,%d %d,%d %d,%d" fill="#6b6b6b" opacity="0.6"/>`,
				rx, ry, rx+8, ry-5, rx+12, ry+3)
		}

	case MapStyleCity:
		// Building
		fmt.Fprintf(&sb, `<rect x="%d" y="%d" width="%d" height="%d" fill="#d4c5a9" stroke="#8b7355" stroke-width="2" filter="url(#shadow)"/>`,
			x, y, w, h)

		// Inner courtyard/detail
		fmt.Fprintf(&sb, `<rect x="%d" y="%d" width="%d" height="%d" fill="#c9ad6a" stroke="#a08c5a" stroke-width="1"/>`,
			x+4, y+4, w-8, h-8)

		// Windows/doors
		fmt.Fprintf(&sb, `<rect x="%d" y="%d" width="%d" height="%d" fill="#5a3d2b"/>`,
			x+w/2-4, y+4, 8, 12)
		fmt.Fprintf(&sb, `<rect x="%d" y="%d" width="%d" height="%d" fill="#5a3d2b"/>`,
			x+6, y+h/2-4, 8, 8)
		fmt.Fprintf(&sb, `<rect x="%d" y="%d" width="%d" height="%d" fill="#5a3d2b"/>`,
			x+w-14, y+h/2-4, 8, 8)
	}

	return sb.String()
}

func renderRoomLabel(room Room, gs int) string {
	var sb strings.Builder

	x := room.X * gs
	y := room.Y * gs
	w := room.W * gs
	h := room.H * gs

	if room.Label == "" {
		return ""
	}

	// Background for text readability
	textWidth := len(room.Label) * 8
	if textWidth < 50 {
		textWidth = 50
	}

	fmt.Fprintf(&sb, `<rect x="%d" y="%d" width="%d" height="20" fill="#1a1a1a" opacity="0.8" rx="4"/>`,
		x+w/2-textWidth/2, y+h/2-10, textWidth)

	fmt.Fprintf(&sb, `<text x="%d" y="%d" font-family="Arial, Helvetica, sans-serif" font-size="12" fill="#f5f0e6" text-anchor="middle" dominant-baseline="central" font-weight="bold">%s</text>`,
		x+w/2, y+h/2, room.Label)

	return sb.String()
}

func GenerateDivider(width int, style string) string {
	switch style {
	case "ornate":
		return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="30" viewBox="0 0 %d 30">
<line x1="0" y1="15" x2="%d" y2="15" stroke="#c9ad6a" stroke-width="1"/>
<path d="M %d 15 L %d 5 L %d 25 Z" fill="#c9ad6a" opacity="0.8"/>
<circle cx="%d" cy="15" r="3" fill="#5a3d2b"/>
<circle cx="%d" cy="15" r="3" fill="#5a3d2b"/>
</svg>`, width, width, width, width/2, width/2-10, width/2+10, 40, width-40)
	case "simple":
		return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="10" viewBox="0 0 %d 10">
<line x1="0" y1="5" x2="%d" y2="5" stroke="#c9ad6a" stroke-width="1"/>
</svg>`, width, width, width)
	case "double":
		return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" viewBox="0 0 %d 20">
<line x1="0" y1="7" x2="%d" y2="7" stroke="#c9ad6a" stroke-width="1"/>
<line x1="0" y1="13" x2="%d" y2="13" stroke="#c9ad6a" stroke-width="1"/>
</svg>`, width, width, width, width)
	default:
		return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="10" viewBox="0 0 %d 10">
<line x1="0" y1="5" x2="%d" y2="5" stroke="#c9ad6a" stroke-width="1"/>
</svg>`, width, width, width)
	}
}

func GenerateStatBlockBorder(width, height int) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">
<rect x="2" y="2" width="%d" height="%d" fill="none" stroke="#c9ad6a" stroke-width="3"/>
<rect x="6" y="6" width="%d" height="%d" fill="none" stroke="#c9ad6a" stroke-width="1"/>
<line x1="10" y1="10" x2="%d" y2="10" stroke="#c9ad6a" stroke-width="1"/>
</svg>`, width, height, width, height, width-4, height-4, width-12, height-12, width-10)
}
