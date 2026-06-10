package domain

// CoinPurse represents a collection of coins.
type CoinPurse struct {
	CP int `json:"cp"`
	SP int `json:"sp"`
	GP int `json:"gp"`
	PP int `json:"pp"`
}

// TotalGP returns the total value in gold pieces.
func (c CoinPurse) TotalGP() int {
	return c.GP + c.PP*10 + c.SP/10 + c.CP/100
}

// ArtObject represents a piece of art or valuable object.
type ArtObject struct {
	Description string `json:"description"`
	ValueGP     int    `json:"value_gp"`
}

// Gem represents a precious stone.
type Gem struct {
	Description string `json:"description"`
	ValueGP     int    `json:"value_gp"`
}

// MagicItemRoll represents a rolled magic item result.
type MagicItemRoll struct {
	Name   string          `json:"name"`
	Rarity MagicItemRarity `json:"rarity"`
	Table  string          `json:"table,omitempty"`
}

// TreasureHoard represents the complete result of a treasure hoard roll.
type TreasureHoard struct {
	Tier       int             `json:"tier"`
	Coins      []CoinPurse     `json:"coins"`
	ArtObjects []ArtObject     `json:"art_objects"`
	Gems       []Gem           `json:"gems"`
	MagicItems []MagicItemRoll `json:"magic_items"`
}
