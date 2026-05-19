package services

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// toneHooks maps tone → narrative hook template (Part 1, read-aloud).
// Uses %s as placeholder for the campaign's mcguffin/villain reference.
var toneHooks = map[string]string{
	"grim":    "Las sombras se alargan sobre %s. Un rumor antiguo, apenas susurrado en las tabernas, habla de un poder oscuro que despierta en las profundidades. Quienes han visto sus señales ya no vuelven a sonreír.",
	"heroic":  "Un gran destino aguarda en %s. Leyendas olvidadas resurgen, y los ecos de una llamada a la aventura resuenan en los corazones de quienes se atreven a soñar. El mundo necesita héroes una vez más.",
	"mystery": "Sucesos extraños perturban %s. Mensajes cifrados, desapariciones inexplicables, y un hilo de plata que conecta todos los puntos. Alguien—o algo—esconde la verdad. ¿Quién se atreverá a descubrirla?",
	"horror":  "Algo terrible se agita bajo %s. En las pesadillas de los sabios, una presencia antigua llama desde las tinieblas. Quienes escuchan demasiado tiempo sienten como algo los observa desde el otro lado.",
}

// toneRoadAheads maps tone → road-ahead template (Part 4, read-aloud).
var toneRoadAheads = map[string]string{
	"grim":    "El camino por delante es incierto y peligroso. Pero incluso en la oscuridad más profunda, una chispa de esperanza puede guiar a quienes se niegan a rendirse. La pregunta es: ¿estás listo para pagar el precio?",
	"heroic":  "La aventura promete gloria, peligro y descubrimiento. Los héroes forjarán su destino con cada decisión, y el mundo recordará sus nombres. Que los dioses acompañen sus pasos.",
	"mystery": "Cada respuesta trae una nueva pregunta. Las piezas del rompecabezas están dispersas, esperando a quien las una. El viaje recién comienza, y la verdad es tan esquiva como peligrosa.",
	"horror":  "La oscuridad se cierne, implacable. No todos los que parten regresan, y algunos regresan... cambiados. Que tengas valor para enfrentar lo que aguarda, porque una vez que cruzas el umbral, no hay vuelta atrás.",
}

// defaultHook is used when the tone is not in the toneHooks map.
const defaultHook = "Una historia comienza en %s. Los hilos del destino se entrelazan, y lo que está por venir cambiará el rumbo de quienes se atrevan a mirar más allá del horizonte."

// defaultRoadAhead is used when the tone is not in the toneRoadAheads map.
const defaultRoadAhead = "La aventura comienza ahora. Lo que aguarda más allá del horizonte solo aquellos con coraje suficiente podrán descubrirlo."

// PrologueService handles generation of campaign prologue narratives.
type PrologueService struct {
	baseDir   string
	canonRepo repository.CanonRepository
}

// NewPrologueService creates a new prologue generation service.
func NewPrologueService(baseDir string, canonRepo repository.CanonRepository) *PrologueService {
	return &PrologueService{
		baseDir:   baseDir,
		canonRepo: canonRepo,
	}
}

// GeneratePrologue creates a 4-part narrative prologue for a campaign.
// It reads lore and introduction from the filesystem, loads canon entities,
// and assembles tone-appropriate prose for each part.
// Returns the Prologue struct, warnings, and error.
func (s *PrologueService) GeneratePrologue(ctx context.Context, campaignID, tone string, characterHooks []string) (*domain.Prologue, []string, error) {
	var warnings []string

	if tone == "" {
		tone = "heroic"
	}

	campaignDir := filepath.Join(s.baseDir, campaignID)

	// Validate campaign exists
	if _, err := os.Stat(campaignDir); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("campaign not found: %s", campaignID)
	}

	// Read lore and introduction as context sources
	loreContent := readCampaignFile(campaignDir, "lore.md")
	introContent := readCampaignFile(campaignDir, "introduction.md")

	// Load canon document for entities
	canonDoc, err := s.canonRepo.Load(campaignID)
	var entities []domain.CanonEntity
	if err != nil {
		warnings = append(warnings, "could not load canon document, using generic references")
	} else {
		entities = canonDoc.Entities
	}

	// Find campaign name from lore/intro or use campaignID
	campaignName := pickCampaignName(loreContent, introContent, campaignID)
	settingDesc := pickSettingDescription(loreContent, introContent)

	// Select tone-appropriate templates
	hookTmpl := selectToneTemplate(tone, toneHooks, defaultHook)
	roadAheadTmpl := selectToneTemplate(tone, toneRoadAheads, defaultRoadAhead)

	// Pick canon reference for the hook
	canonRef := pickCanonReference(entities)
	if canonRef == "" {
		canonRef = "el mundo"
		warnings = append(warnings, "no villain or mcguffin defined, using generic hook")
	}

	// Build context (Part 2) from lore/intro
	contextText := buildContextText(settingDesc, loreContent, introContent, campaignName)

	// Build connections (Part 3)
	connectionsText := buildConnectionsText(entities, campaignName)
	if len(entities) == 0 {
		warnings = append(warnings, "no entities found for connections section")
	}

	// Apply hook template with canon reference
	hookText := fmt.Sprintf(hookTmpl, canonRef)
	roadAheadText := roadAheadTmpl

	now := time.Now()

	prologue := &domain.Prologue{
		CampaignID: campaignID,
		Tone:       tone,
		Parts: []domain.ProloguePart{
			{
				Order:       1,
				Title:       "Gancho Narrativo",
				Content:     hookText,
				IsReadAloud: true,
			},
			{
				Order:       2,
				Title:       "Trasfondo",
				Content:     contextText,
				IsReadAloud: false,
			},
			{
				Order:       3,
				Title:       "Conexiones",
				Content:     connectionsText,
				IsReadAloud: false,
			},
			{
				Order:       4,
				Title:       "El Camino por Delante",
				Content:     roadAheadText,
				IsReadAloud: true,
			},
		},
		GeneratedAt: now,
	}

	return prologue, warnings, nil
}

// readCampaignFile reads a markdown file from the campaign directory.
// Returns empty string if the file does not exist or cannot be read.
func readCampaignFile(campaignDir, filename string) string {
	path := filepath.Join(campaignDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// selectToneTemplate picks the template for the given tone, falling back to
// the default template if the tone is unknown.
func selectToneTemplate(tone string, templates map[string]string, defaultTmpl string) string {
	if tmpl, ok := templates[tone]; ok {
		return tmpl
	}
	return defaultTmpl
}

// pickCampaignName extracts a display name from lore/intro or falls back to campaignID.
func pickCampaignName(loreContent, introContent, campaignID string) string {
	// Try to find a title in the introduction (first # heading)
	for _, content := range []string{introContent, loreContent} {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "# ") {
				return strings.TrimPrefix(trimmed, "# ")
			}
		}
	}
	return campaignID
}

// pickSettingDescription extracts a setting description from lore/intro.
func pickSettingDescription(loreContent, introContent string) string {
	preferIntro := true
	source := introContent
	if strings.TrimSpace(source) == "" {
		source = loreContent
		preferIntro = false
	}
	if strings.TrimSpace(source) == "" {
		return ""
	}

	lines := strings.Split(source, "\n")
	var descLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ">") {
			continue
		}
		descLines = append(descLines, trimmed)
	}

	if len(descLines) == 0 {
		return ""
	}

	// Take up to 3 non-empty lines as setting description
	limit := 3
	if len(descLines) < limit {
		limit = len(descLines)
	}

	// For intro content, prefer the description paragraph right after the heading
	if preferIntro && len(descLines) > 0 {
		_ = 0 // use first meaningful lines
	}

	return strings.Join(descLines[:limit], " ")
}

// pickCanonReference selects the most important entity to reference in the hook.
func pickCanonReference(entities []domain.CanonEntity) string {
	if len(entities) == 0 {
		return ""
	}

	// Prefer mcguffin, villain, or ally roles
	var candidates []string
	for _, e := range entities {
		if e.Role == "mcguffin" || e.Role == "villain" || e.Role == "ally" {
			candidates = append(candidates, e.Name)
		}
	}

	if len(candidates) == 0 {
		for _, e := range entities {
			if e.Name != "" {
				candidates = append(candidates, e.Name)
			}
		}
	}

	if len(candidates) == 0 {
		return ""
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return candidates[rng.Intn(len(candidates))]
}

// buildContextText generates the context/background section (Part 2).
func buildContextText(settingDesc, loreContent, introContent, campaignName string) string {
	var b strings.Builder

	b.WriteString("Este prólogo presenta el trasfondo de la campaña. ")

	if settingDesc != "" {
		b.WriteString(settingDesc)
		b.WriteString(" ")
	}

	if strings.TrimSpace(loreContent) != "" {
		b.WriteString("La historia de esta tierra está marcada por eventos que pocos recuerdan y muchos temen. ")
		// Extract first substantive paragraph from lore
		lines := strings.Split(loreContent, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, ">") && len(trimmed) > 40 {
				b.WriteString(trimmed)
				b.WriteString(" ")
				break
			}
		}
	}

	if campaignName != "" {
		b.WriteString(fmt.Sprintf("El destino de %s pende de un hilo, y solo aquellos con el valor suficiente podrán cambiar el curso de los acontecimientos.", campaignName))
	}

	result := strings.TrimSpace(b.String())
	if result == "" {
		result = "Una nueva aventura aguarda. Los personajes se verán envueltos en una trama que trasciende sus propias historias, y sus decisiones moldearán el futuro del mundo que los rodea."
	}

	return result
}

// buildConnectionsText generates the connections section (Part 3).
func buildConnectionsText(entities []domain.CanonEntity, campaignName string) string {
	if len(entities) == 0 {
		return "El mundo está lleno de personas, lugares y secretos por descubrir. A medida que avance la campaña, los personajes forjarán sus propias alianzas y enfrentarán a sus enemigos."
	}

	var b strings.Builder
	b.WriteString("A continuación, algunas de las fuerzas y figuras que darán forma a esta historia:")

	// Group entities by role
	var villains, allies, mcguffins, others []domain.CanonEntity
	for _, e := range entities {
		switch e.Role {
		case "villain":
			villains = append(villains, e)
		case "ally":
			allies = append(allies, e)
		case "mcguffin":
			mcguffins = append(mcguffins, e)
		default:
			others = append(others, e)
		}
	}

	if len(villains) > 0 {
		b.WriteString("\n\n**Antagonistas:** ")
		var names []string
		for _, v := range villains {
			names = append(names, v.Name)
		}
		b.WriteString(strings.Join(names, ", "))
		b.WriteString(".")
	}

	if len(allies) > 0 {
		b.WriteString("\n\n**Aliados:** ")
		var names []string
		for _, a := range allies {
			names = append(names, a.Name)
		}
		b.WriteString(strings.Join(names, ", "))
		b.WriteString(".")
	}

	if len(mcguffins) > 0 {
		b.WriteString("\n\n**Objetos de Interés:** ")
		var names []string
		for _, m := range mcguffins {
			names = append(names, m.Name)
		}
		b.WriteString(strings.Join(names, ", "))
		b.WriteString(".")
	}

	if len(others) > 0 {
		b.WriteString("\n\n**Otras Entidades:** ")
		var names []string
		for _, o := range others {
			names = append(names, o.Name)
		}
		b.WriteString(strings.Join(names, ", "))
		b.WriteString(".")
	}

	return b.String()
}

// BuildPrologueMarkdown executes the prologue template with the given data
// and returns the rendered markdown.
func BuildPrologueMarkdown(prologue *domain.Prologue, tmplStr string) (string, error) {
	funcMap := template.FuncMap{}
	tmpl, err := template.New("prologue").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse prologue template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, prologue); err != nil {
		return "", fmt.Errorf("failed to execute prologue template: %w", err)
	}

	return buf.String(), nil
}
