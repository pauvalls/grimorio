package svg

import (
	"fmt"
	"math/rand"
	"strings"
)

type MapStyle string

const (
	MapStyleDungeon  MapStyle = "dungeon"
	MapStyleLandscape MapStyle = "landscape"
	MapStyleCity     MapStyle = "city"
)

type BattleMapConfig struct {
	Width        int
	Height       int
	GridSize     int
	NumRooms     int
	MinRoomSize  int
	MaxRoomSize  int
	Style        MapStyle
	Title        string
	Seed         int64
}

func DefaultBattleMapConfig() BattleMapConfig {
	return BattleMapConfig{
		Width:       800,
		Height:      600,
		GridSize:    40,
		NumRooms:    6,
		MinRoomSize: 2,
		MaxRoomSize: 4,
		Style:       MapStyleDungeon,
		Seed:        0,
	}
}

type Room struct {
	X, Y, W, H int
	Label      string
}

func GenerateBattleMap(cfg BattleMapConfig) string {
	if cfg.Seed != 0 {
		rand.Seed(cfg.Seed)
	} else {
		rand.Seed(rand.Int63())
	}

	cols := cfg.Width / cfg.GridSize
	rows := cfg.Height / cfg.GridSize

	rooms := placeRooms(cfg, cols, rows)
	connections := connectRooms(rooms, cols, rows)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, cfg.Width, cfg.Height, cfg.Width, cfg.Height))
	sb.WriteString(`<defs>`)
	sb.WriteString(`<pattern id="grid" width="`)
	sb.WriteString(fmt.Sprintf("%d", cfg.GridSize))
	sb.WriteString(`" height="`)
	sb.WriteString(fmt.Sprintf("%d", cfg.GridSize))
	sb.WriteString(`" patternUnits="userSpaceOnUse">`)
	sb.WriteString(`<path d="M `)
	sb.WriteString(fmt.Sprintf("%d 0 L 0 0 0 %d", cfg.GridSize, cfg.GridSize))
	sb.WriteString(`" fill="none" stroke="#c9ad6a" stroke-width="0.5" opacity="0.4"/>`)
	sb.WriteString(`</pattern>`)
	sb.WriteString(`</defs>`)

	sb.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="#2c1e14"/>`, cfg.Width, cfg.Height))
	sb.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="url(#grid)"/>`, cfg.Width, cfg.Height))

	for _, conn := range connections {
		sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#5a3d2b" stroke-width="%d" opacity="0.6"/>`,
			conn.x1, conn.y1, conn.x2, conn.y2, cfg.GridSize-4))
	}

	for _, room := range rooms {
		x := room.X * cfg.GridSize
		y := room.Y * cfg.GridSize
		w := room.W * cfg.GridSize
		h := room.H * cfg.GridSize

		switch cfg.Style {
		case MapStyleDungeon:
			sb.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#3d2e1f" stroke="#c9ad6a" stroke-width="2" rx="2"/>`, x, y, w, h))
			sb.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="#8b7355" stroke-width="1" rx="1"/>`, x+4, y+4, w-8, h-8))
		case MapStyleLandscape:
			sb.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#4a6741" stroke="#2d4a27" stroke-width="2" rx="4"/>`, x, y, w, h))
			g := rand.Intn(3)
			for i := 0; i < g; i++ {
				tx := x + rand.Intn(w-20) + 10
				ty := y + rand.Intn(h-20) + 10
				sb.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="8" fill="#2d4a27"/>`, tx, ty))
				sb.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="5" fill="#3d5a37"/>`, tx, ty-5))
			}
		case MapStyleCity:
			sb.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#d4c5a9" stroke="#8b7355" stroke-width="2"/>`, x, y, w, h))
			sb.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#c9ad6a" stroke="#a08c5a" stroke-width="1"/>`, x+4, y+4, w-8, h-8))
		}

		if room.Label != "" {
			sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Cinzel, serif" font-size="11" fill="#f5f0e6" text-anchor="middle" dominant-baseline="central">%s</text>`,
				x+w/2, y+h/2, room.Label))
		}
	}

	if cfg.Title != "" {
		sb.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Cinzel, serif" font-size="18" fill="#c9ad6a" text-anchor="middle" font-weight="bold">%s</text>`,
			cfg.Width/2, 24, cfg.Title))
	}

	sb.WriteString(`</svg>`)
	return sb.String()
}

type connection struct {
	x1, y1, x2, y2 int
}

func placeRooms(cfg BattleMapConfig, cols, rows int) []Room {
	var rooms []Room
	attempts := 0
	maxAttempts := cfg.NumRooms * 50

	for len(rooms) < cfg.NumRooms && attempts < maxAttempts {
		attempts++
		size := cfg.MinRoomSize + rand.Intn(cfg.MaxRoomSize-cfg.MinRoomSize+1)
		x := rand.Intn(cols - size - 2) + 1
		y := rand.Intn(rows - size - 2) + 1

		room := Room{X: x, Y: y, W: size, H: size}
		overlap := false
		for _, existing := range rooms {
			if x < existing.X+existing.W+1 && x+room.W+1 > existing.X &&
				y < existing.Y+existing.H+1 && y+room.H+1 > existing.Y {
				overlap = true
				break
			}
		}
		if !overlap {
			rooms = append(rooms, room)
		}
	}

	for i := range rooms {
		if i < 3 {
			rooms[i].Label = fmt.Sprintf("Room %d", i+1)
		}
	}

	return rooms
}

func connectRooms(rooms []Room, cols, rows int) []connection {
	var conns []connection
	if len(rooms) < 2 {
		return conns
	}

	for i := 1; i < len(rooms); i++ {
		prev := rooms[i-1]
		curr := rooms[i]

		x1 := (prev.X + prev.W/2) * 40
		y1 := (prev.Y + prev.H/2) * 40
		x2 := (curr.X + curr.W/2) * 40
		y2 := (curr.Y + curr.H/2) * 40

		if rand.Intn(2) == 0 {
			midX := (curr.X + curr.W/2) * 40
			conns = append(conns,
				connection{x1, y1, midX, y1},
				connection{midX, y1, midX, y2},
			)
		} else {
			midY := (curr.Y + curr.H/2) * 40
			conns = append(conns,
				connection{x1, y1, x1, midY},
				connection{x1, midY, x2, midY},
			)
		}
	}

	return conns
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
