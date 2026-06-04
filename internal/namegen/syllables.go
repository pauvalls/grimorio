package namegen

// SyllablePool holds phonetic fragments for a given category and style.
type SyllablePool struct {
	Onsets       []string // consonant starts, e.g. ["th", "br", "el"]
	Vowels       []string // vowel cores, e.g. ["a", "ae", "o"]
	Codas        []string // consonant ends, e.g. ["n", "r", "k"]
	Prefixes     []string // for compound / descriptive names (cities, taverns, factions)
	Suffixes     []string // for compound / descriptive names
	MinSyllables int      // minimum syllables for syllable-stacking categories
	MaxSyllables int      // maximum syllables for syllable-stacking categories
}

// Pools is the global registry of syllable data per (Category, Style).
var Pools = map[Category]map[Style]*SyllablePool{
	CategoryCharacter: characterPools(),
	CategoryNPC:       characterPools(), // same phonetic rules as characters
	CategoryCity:      cityPools(),
	CategoryTavern:    tavernPools(),
	CategoryMonster:   monsterPools(),
	CategoryFaction:   factionPools(),
	CategoryItem:      itemPools(),
}

func characterPools() map[Style]*SyllablePool {
	return map[Style]*SyllablePool{
		StyleGenericFantasy: {
			Onsets: []string{"th", "r", "l", "sh", "br", "dr", "tr", "kr", "gl", "gr", "bl", "cl", "fl", "sl", "str", "spr", "scr", "spl", "thr", "chr"},
			Vowels: []string{"a", "e", "i", "o", "u", "ae", "ai", "ee", "ei", "oo", "ou", "au", "ia", "io", "ua"},
			Codas:  []string{"n", "r", "th", "s", "k", "t", "d", "l", "m", "ng", "x", "g", "ch", "sh", "rk"},
			MinSyllables: 2,
			MaxSyllables: 4,
		},
		StyleElven: {
			Onsets: []string{"l", "r", "th", "s", "m", "n", "v", "f", "sh", "el", "al", "il", "ol", "vel", "sil", "mel", "nal", "thal", "ar", "er"},
			Vowels: []string{"a", "e", "i", "o", "u", "ae", "ai", "ee", "ia", "io", "au", "ei", "oo", "ue", "eo"},
			Codas:  []string{"l", "n", "r", "s", "th", "m", "ng", "v", "f", "nn", "ll", "el", "ar", "ir", "or"},
			MinSyllables: 2,
			MaxSyllables: 4,
		},
		StyleDwarven: {
			Onsets: []string{"br", "dr", "gr", "kr", "tr", "thr", "str", "dw", "dur", "gor", "krak", "mor", "nor", "bor", "thor", "kh", "grum", "thra", "grom", "drok"},
			Vowels: []string{"a", "o", "u", "i", "e", "au", "oo", "uu", "ar", "or", "ur", "ir", "er", "ai", "ou"},
			Codas:  []string{"k", "g", "r", "n", "m", "t", "d", "rk", "rg", "rm", "rn", "ng", "z", "th", "ch"},
			MinSyllables: 2,
			MaxSyllables: 3,
		},
		StyleOrcish: {
			Onsets: []string{"gr", "kr", "tr", "br", "dr", "z", "gh", "kh", "th", "sh", "r", "g", "k", "t", "zg", "sk", "grz", "thr", "brk", "ghr"},
			Vowels: []string{"a", "o", "u", "i", "e", "aa", "oo", "uu", "au", "ou", "ar", "or", "ur", "ir", "er"},
			Codas:  []string{"k", "g", "r", "z", "t", "d", "n", "m", "sh", "th", "ch", "rg", "rk", "rz", "gh"},
			MinSyllables: 2,
			MaxSyllables: 3,
		},
		StyleHumanMedieval: {
			Onsets: []string{"r", "l", "s", "t", "d", "n", "m", "b", "p", "g", "c", "f", "h", "w", "st", "sp", "sc", "sh", "th", "ch"},
			Vowels: []string{"a", "e", "i", "o", "u", "y", "ai", "ee", "oo", "ou", "au", "ie", "ea", "oa", "ue"},
			Codas:  []string{"n", "r", "s", "t", "d", "l", "m", "k", "g", "th", "sh", "ch", "nd", "rd", "ld"},
			MinSyllables: 2,
			MaxSyllables: 4,
		},
	}
}

func cityPools() map[Style]*SyllablePool {
	return map[Style]*SyllablePool{
		StyleGenericFantasy: {
			Prefixes: []string{"Shadow", "Iron", "Storm", "Bright", "Dark", "Silver", "Golden", "Red", "Blue", "Green", "Frost", "Fire", "Wind", "Stone", "Wood", "Moon", "Sun", "Star", "Mist", "Wild"},
			Suffixes: []string{"dale", "ford", "keep", "hold", "haven", "port", "vale", "wood", "moor", "field", "gate", "wall", "march", "ridge", "crest", "fall", "mere", "water", "land", "way"},
		},
		StyleElven: {
			Prefixes: []string{"Moon", "Star", "Silver", "Mist", "Dawn", "Even", "Light", "Shadow", "Dream", "Leaf", "Wood", "Sun", "Sky", "Crystal", "Whisper", "Twilight", "Ember", "Frost", "Gale", "Vale"},
			Suffixes: []string{"dore", "lin", "lor", "thas", "weir", "na", "riel", "dor", "wen", "las", "thal", "mar", "wind", "glade", "vale", "rest", "haven", "light", "dawn", "shore"},
		},
		StyleDwarven: {
			Prefixes: []string{"Iron", "Stone", "Gold", "Deep", "Forge", "Hammer", "Ax", "Shield", "Fire", "Dark", "Grim", "Bold", "Steel", "Rock", "Mithril", "Ember", "Coal", "Anvil", "Ore", "Under"},
			Suffixes: []string{"hold", "deep", "forge", "hall", "mount", "peak", "stone", "mine", "vault", "dell", "barrow", "cairn", "spire", "crag", "pass", "gap", "run", "guard", "watch", "end"},
		},
		StyleOrcish: {
			Prefixes: []string{"Blood", "Iron", "Skull", "Bone", "Dark", "Grim", "Death", "War", "Gore", "Rust", "Ash", "Fire", "Rot", "Shadow", "Venom", "Scream", "Ravage", "Terror", "Doom", "Chaos"},
			Suffixes: []string{"skull", "blood", "fang", "claw", "maw", "pit", "den", "lair", "gorge", "waste", "scar", "mark", "blight", "rot", "ash", "doom", "death", "ruin", "crush", "rend"},
		},
		StyleHumanMedieval: {
			Prefixes: []string{"New", "Old", "High", "Low", "North", "South", "East", "West", "Kings", "Queens", "Castle", "Port", "Bridge", "River", "Hill", "Green", "Fair", "Market", "Grand", "Royal"},
			Suffixes: []string{"ton", "bury", "wick", "ham", "ford", "port", "field", "wood", "hill", "bridge", "castle", "gate", "market", "cross", "well", "spring", "mill", "abbey", "manor", "court"},
		},
	}
}

func tavernPools() map[Style]*SyllablePool {
	return map[Style]*SyllablePool{
		StyleGenericFantasy: {
			Prefixes: []string{"Golden", "Silver", "Red", "Green", "Black", "White", "Rusty", "Drunken", "Laughing", "Sleeping", "Broken", "Hidden", "Wandering", "Twisted", "Shattered", "Mystic", "Singing", "Roaring", "Prancing", "Howling"},
			Suffixes: []string{"Dragon", "Griffin", "Tankard", "Barrel", "Mug", "Horse", "Lion", "Eagle", "Stag", "Raven", "Fox", "Bear", "Wolf", "Owl", "Rose", "Thistle", "Crown", "Anchor", "Sword", "Shield"},
		},
		StyleElven: {
			Prefixes: []string{"Moonlit", "Starlit", "Silver", "Golden", "Whispering", "Dreaming", "Eternal", "Enchanted", "Ethereal", "Glimmering", "Radiant", "Serene", "Tranquil", "Verdant", "Luminous", "Mystic", "Singing", "Dancing", "Shimmering", "Crystal"},
			Suffixes: []string{"Unicorn", "Phoenix", "Nightingale", "Sparrow", "Willow", "Oak", "Meadow", "Brook", "Glade", "Grove", "Fawn", "Dove", "Swan", "Pearl", "Sapphire", "Emerald", "Ruby", "Amethyst", "Star", "Moon"},
		},
		StyleDwarven: {
			Prefixes: []string{"Sturdy", "Iron", "Golden", "Deep", "Hammered", "Forged", "Stout", "Bitter", "Fiery", "Merry", "Grim", "Bold", "Hardened", "Keen", "Royal", "Rusted", "Sooty", "Smoking", "Thundering", "Drunken"},
			Suffixes: []string{"Anvil", "Forge", "Mug", "Tankard", "Keg", "Barrel", "Horn", "Axe", "Hammer", "Shield", "Boar", "Goat", "Ram", "Bear", "Wolf", "Eagle", "Owl", "Raven", "Mountain", "Hearth"},
		},
		StyleOrcish: {
			Prefixes: []string{"Bloody", "Rotten", "Broken", "Skull", "Gore", "Filthy", "Savage", "Feral", "Crimson", "Black", "Burning", "Crushed", "Ravaged", "Screaming", "Wailing", "Howling", "Spiked", "Rusty", "Chained", "Cursed"},
			Suffixes: []string{"Tusk", "Claw", "Fang", "Maw", "Boar", "Wolf", "Bear", "Vulture", "Crow", "Rat", "Snake", "Scorpion", "Spider", "Worm", "Leech", "Maggot", "Goblin", "Troll", "Ogre", "Demon"},
		},
		StyleHumanMedieval: {
			Prefixes: []string{"Kings", "Queens", "Noble", "Common", "Old", "New", "Royal", "Travelers", "Merchants", "Farmers", "Hunters", "Knights", "Ladies", "Good", "Jolly", "Merry", "Drunken", "Sleeping", "Wandering", "Prancing"},
			Suffixes: []string{"Inn", "Tavern", "Alehouse", "Pub", "Hostel", "Lodge", "Rest", "Hearth", "Hall", "Manor", "Castle", "Bridge", "Cross", "Well", "Mill", "Oak", "Rose", "Lion", "Bear", "Eagle"},
		},
	}
}

func monsterPools() map[Style]*SyllablePool {
	return map[Style]*SyllablePool{
		StyleGenericFantasy: {
			Onsets: []string{"gr", "kr", "tr", "br", "dr", "vr", "xr", "zr", "skr", "gry", "myr", "nyr", "pyr", "tyr", "syr", "wyr", "hyr", "lyr", "fyr", "ryr"},
			Vowels: []string{"a", "o", "u", "i", "e", "aa", "oo", "uu", "au", "ou", "ar", "or", "ur", "ir", "er"},
			Codas:  []string{"k", "g", "r", "z", "t", "d", "n", "m", "sh", "th", "ch", "rg", "rk", "rz", "gh", "x", "xx", "kk", "gg", "zz"},
			MinSyllables: 1,
			MaxSyllables: 3,
		},
		StyleElven: {
			Onsets: []string{"sh", "th", "v", "f", "s", "m", "n", "l", "r", "shal", "thal", "vel", "sil", "mel", "nal", "el", "al", "il", "ol", "ar"},
			Vowels: []string{"a", "e", "i", "o", "u", "ae", "ai", "ee", "ia", "io", "au", "ei", "oo", "ue", "eo"},
			Codas:  []string{"l", "n", "r", "s", "th", "m", "ng", "v", "f", "nn", "ll", "el", "ar", "ir", "or"},
			MinSyllables: 1,
			MaxSyllables: 3,
		},
		StyleDwarven: {
			Onsets: []string{"br", "dr", "gr", "kr", "tr", "thr", "str", "kh", "grum", "thra", "grom", "drok", "grak", "trak", "brok", "krok", "mok", "nok", "rok", "tok"},
			Vowels: []string{"a", "o", "u", "i", "e", "au", "oo", "uu", "ar", "or", "ur", "ir", "er", "ai", "ou"},
			Codas:  []string{"k", "g", "r", "n", "m", "t", "d", "rk", "rg", "rm", "rn", "ng", "z", "th", "ch"},
			MinSyllables: 1,
			MaxSyllables: 3,
		},
		StyleOrcish: {
			Onsets: []string{"gr", "kr", "tr", "br", "dr", "z", "gh", "kh", "th", "sh", "zg", "sk", "grz", "thr", "brk", "ghr", "grk", "krk", "trk", "drk"},
			Vowels: []string{"a", "o", "u", "i", "e", "aa", "oo", "uu", "au", "ou", "ar", "or", "ur", "ir", "er"},
			Codas:  []string{"k", "g", "r", "z", "t", "d", "n", "m", "sh", "th", "ch", "rg", "rk", "rz", "gh", "kk", "gg", "zz", "xx", "rr"},
			MinSyllables: 1,
			MaxSyllables: 3,
		},
		StyleHumanMedieval: {
			Onsets: []string{"r", "l", "s", "t", "d", "n", "m", "b", "p", "g", "c", "f", "h", "w", "st", "sp", "sc", "sh", "th", "ch"},
			Vowels: []string{"a", "e", "i", "o", "u", "y", "ai", "ee", "oo", "ou", "au", "ie", "ea", "oa", "ue"},
			Codas:  []string{"n", "r", "s", "t", "d", "l", "m", "k", "g", "th", "sh", "ch", "nd", "rd", "ld"},
			MinSyllables: 1,
			MaxSyllables: 3,
		},
	}
}

func factionPools() map[Style]*SyllablePool {
	return map[Style]*SyllablePool{
		StyleGenericFantasy: {
			Prefixes: []string{"Order", "Guild", "League", "Circle", "Cult", "Brotherhood", "Sisterhood", "Alliance", "Covenant", "Pact", "Society", "Fellowship", "Syndicate", "Council", "Assembly", "Congregation", "Tribe", "Clan", "House", "Faction"},
			Suffixes: []string{"Flame", "Shadow", "Light", "Storm", "Void", "Ancients", "Moon", "Sun", "Stars", "Earth", "Sea", "Sky", "Raven", "Wolf", "Dragon", "Rose", "Thorn", "Veil", "Crown", "Scepter"},
		},
		StyleElven: {
			Prefixes: []string{"Circle", "Order", "Covenant", "Enclave", "Sanctum", "Assembly", "Conclave", "Gathering", "Vigil", "Watch", "Sentinel", "Guardian", "Warden", "Keeper", "Herald", "Speaker", "Voice", "Song", "Weave", "Path"},
			Suffixes: []string{"Moon", "Stars", "Sun", "Forest", "Glade", "Brook", "Wind", "Dawn", "Twilight", "Dream", "Song", "Weave", "Light", "Shadow", "Veil", "Shore", "Vale", "Grove", "Leaf", "Petal"},
		},
		StyleDwarven: {
			Prefixes: []string{"Clan", "Hold", "Forge", "Guild", "Order", "Brotherhood", "Council", "Assembly", "Company", "Band", "Crew", "Squad", "Unit", "Legion", "Army", "Host", "Kin", "Kinship", "Blood", "Oath"},
			Suffixes: []string{"Anvil", "Hammer", "Shield", "Axe", "Deep", "Mountain", "Stone", "Forge", "Fire", "Ember", "Iron", "Steel", "Mithril", "Gold", "Ore", "Mine", "Hearth", "Hall", "Guard", "Watch"},
		},
		StyleOrcish: {
			Prefixes: []string{"Horde", "Clan", "Tribe", "Pack", "Gang", "Crew", "Warband", "Raiders", "Marauders", "Reavers", "Butchers", "Slayers", "Crushers", "Breakers", "Ravagers", "Scourge", "Plague", "Blight", "Doom", "Ruin"},
			Suffixes: []string{"Blood", "Skull", "Bone", "Fang", "Claw", "Maw", "Pit", "Wastes", "Ash", "Fire", "Gore", "Rot", "Venom", "Scream", "Terror", "Doom", "Death", "Ruin", "Crush", "Rend"},
		},
		StyleHumanMedieval: {
			Prefixes: []string{"Order", "Guild", "League", "Brotherhood", "Company", "Society", "Fellowship", "Alliance", "Council", "Court", "Chapel", "Abbey", "Monastery", "Convent", "Chapter", "House", "Family", "Kin", "Clan", "Band"},
			Suffixes: []string{"Crown", "Scepter", "Throne", "Cross", "Shield", "Sword", "Grail", "Chalice", "Rose", "Lily", "Oak", "Eagle", "Lion", "Lamb", "Dove", "Raven", "Stag", "Hart", "Falcon", "Swan"},
		},
	}
}

func itemPools() map[Style]*SyllablePool {
	return map[Style]*SyllablePool{
		StyleGenericFantasy: {
			Prefixes: []string{"Magnificent", "Divine", "Ancient", "Cursed", "Blessed", "Infernal", "Celestial", "Shadow", "Frost", "Flame", "Storm", "Thunder", "Arcane", "Mystic", "Eldritch", "Radiant", "Venomous", "Ghostly", "Crystalline", "Enchanted"},
			Suffixes: []string{"Blade", "Armor", "Staff", "Ring", "Amulet", "Cloak", "Helm", "Shield", "Bow", "Dagger", "Axe", "Hammer", "Spear", "Wand", "Tome", "Orb", "Crown", "Belt", "Boots", "Gauntlets"},
		},
		StyleElven: {
			Prefixes: []string{"Moon", "Star", "Silver", "Crystal", "Ethereal", "Enchanted", "Whispering", "Glimmering", "Luminous", "Radiant", "Serene", "Verdant", "Eternal", "Dreaming", "Singing", "Dancing", "Shimmering", "Mystic", "Arcane", "Eldritch"},
			Suffixes: []string{"Blade", "Bow", "Staff", "Cloak", "Amulet", "Ring", "Crown", "Tome", "Wand", "Orb", "Harp", "Lyre", "Flute", "Mirror", "Chalice", "Pendant", "Circlet", "Bracer", "Scepter", "Shard"},
		},
		StyleDwarven: {
			Prefixes: []string{"Iron", "Steel", "Mithril", "Golden", "Deep", "Forge", "Hammered", "Hardened", "Sturdy", "Grim", "Bold", "Keen", "Ancient", "Cursed", "Blessed", "Runic", "Dwarven", "Stout", "Mighty", "Thunder"},
			Suffixes: []string{"Axe", "Hammer", "Shield", "Armor", "Helm", "Ring", "Amulet", "Belt", "Boots", "Gauntlets", "Cloak", "Blade", "Spear", "Mace", "Flail", "Pick", "Drill", "Anvil", "Tongs", "Forge"},
		},
		StyleOrcish: {
			Prefixes: []string{"Bloody", "Cursed", "Savage", "Feral", "Crimson", "Black", "Burning", "Crushed", "Ravaged", "Screaming", "Spiked", "Rusty", "Chained", "Rotten", "Filthy", "Gore", "Venomous", "Toxic", "Plagued", "Doomed"},
			Suffixes: []string{"Cleaver", "Maul", "Club", "Spear", "Axe", "Dagger", "Shield", "Armor", "Helm", "Gauntlets", "Spikes", "Chains", "Whip", "Flail", "Mace", "Skull", "Tusk", "Claw", "Fang", "Mask"},
		},
		StyleHumanMedieval: {
			Prefixes: []string{"Noble", "Royal", "Knightly", "Common", "Old", "Ancient", "Blessed", "Cursed", "Holy", "Unholy", "Mystic", "Arcane", "Divine", "Infernal", "Celestial", "Shadow", "Frost", "Flame", "Storm", "Thunder"},
			Suffixes: []string{"Sword", "Blade", "Armor", "Shield", "Helm", "Cloak", "Ring", "Amulet", "Staff", "Bow", "Dagger", "Axe", "Hammer", "Spear", "Mace", "Flail", "Lance", "Cross", "Crown", "Scepter"},
		},
	}
}
