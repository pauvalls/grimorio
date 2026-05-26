package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pauvalls/grimorio/internal/repository"
)

func TestCampaignService_CreateCampaign(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, "/tmp/campaigns", "")

	tests := []struct {
		name         string
		campaignName string
		title        string
		setting      string
		wantErr      bool
	}{
		{
			name:         "create valid campaign",
			campaignName: "test-campaign",
			title:        "Test Campaign",
			setting:      "Forgotten Realms",
			wantErr:      false,
		},
		{
			name:         "create campaign without title",
			campaignName: "no-title",
			setting:      "Test",
			wantErr:      false,
		},
		{
			name:         "create duplicate campaign",
			campaignName: "test-campaign",
			title:        "Duplicate",
			wantErr:      true,
		},
		{
			name:         "invalid campaign name",
			campaignName: "Invalid Name",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			campaign, err := service.CreateCampaign(tt.campaignName, tt.title, tt.setting)
			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateCampaign() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateCampaign() unexpected error: %v", err)
				return
			}
			if campaign.Name != tt.campaignName {
				t.Errorf("CreateCampaign() name = %v, want %v", campaign.Name, tt.campaignName)
			}
			if tt.title != "" && campaign.Title != tt.title {
				t.Errorf("CreateCampaign() title = %v, want %v", campaign.Title, tt.title)
			}
		})
	}
}

func TestCampaignService_GetCampaign(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, "/tmp/campaigns", "")

	// Create a campaign first
	_, err := service.CreateCampaign("get-test", "Get Test", "Setting")
	if err != nil {
		t.Fatalf("Failed to create test campaign: %v", err)
	}

	tests := []struct {
		name         string
		campaignName string
		wantErr      bool
	}{
		{
			name:         "get existing campaign",
			campaignName: "get-test",
			wantErr:      false,
		},
		{
			name:         "get non-existent campaign",
			campaignName: "does-not-exist",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			campaign, err := service.GetCampaign(tt.campaignName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetCampaign() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("GetCampaign() unexpected error: %v", err)
				return
			}
			if campaign.Name != tt.campaignName {
				t.Errorf("GetCampaign() name = %v, want %v", campaign.Name, tt.campaignName)
			}
		})
	}
}

func TestCampaignService_SaveAct(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, "/tmp/campaigns", "")

	// Create a campaign first
	_, err := service.CreateCampaign("act-test", "Act Test", "Setting")
	if err != nil {
		t.Fatalf("Failed to create test campaign: %v", err)
	}

	// Helper content with 700+ words for v3.1.0 validation
	validChapterContent := `## Cómo Dirigir Este Capítulo

### 1. Preparación Previa a la Sesión

**Materiales Necesarios:**
- **Mapas:** Mapa del pueblo de Phandalin con distritos marcados, plano del molino viejo con todas las habitaciones, mapa de la cueva subterránea con trampas señaladas, mapa regional de la Costa de la Espada
- **Handouts:** Carta sellada de Gundren Rockseeker, folleto del Carretel Roto con menú y precios, documento de recompensa del Lord de Agua Profunda, pista física con símbolo cifrado
- **NPCs con Stats:** Iarno Albrek (mago nivel 5), Glasstaff (mago nivel 4), 4 guardias de la ciudad (guardias veterano), 2 bandidos en el camino
- **Música/Ambientación:** Sonidos de bosque nocturno para el viaje, música de taberna alegre para el Carretel Roto, tensión de combate para encuentros, sonidos de cueva con goteo de agua

**Repaso de Consecuencias:**
Si los PJs aliaron con los Zhentarim en la sesión anterior, ahora tienen acceso a información privilegiada sobre el mercado negro y pueden comprar equipo a precios reducidos. Si traicionaron a los miners del pueblo, serán emboscados en el camino a Phandalin por venganza. Si completaron la quest del templo, el clérigo principal les debe un favor que pueden reclamar ahora.

**Estado Inicial del Mundo:**
La Guardia de la Ciudad está en alerta máxima después del ataque al almacén la semana pasada. El mercado negro opera solo de noche en los muelles para evitar patrullas. El templo ha cerrado sus puertas por temor a represalias de los cultistas. Los campesinos hablan en susurros sobre "sombras que se mueven de noche". El clima es tormentoso con lluvia intermitente que crea un tono oscuro y peligroso.

### 2. Ritmo y Timing Sugerido

**Desglose por Escenas:**
- **Apertura (15-20 min):** Los PJs llegan al pueblo al anochecer bajo lluvia ligera. Las calles están casi vacías excepto por la guardia patrullando. Encuentro inicial con el quest giver en el Carretel Roto que explica la misión con detalles urgentes.
- **Exploración/Investigación (30-45 min):** Los PJs investigan el lugar del crimen marcado en el mapa. Percepción DC 14 revela huellas de botas claveteadas que van hacia los muelles. Investigación DC 16 encuentra el arma del crimen escondida bajo un barril podrido. Interrogatorio a testigos requiere Persuasión DC 12.
- **Encuentro Principal (45-60 min):** Combate con 4 goblins y 1 ogro en el puente viejo que cruza el arroyo. Los goblins usan tácticas de emboscada desde las sombras. Alternativa: negociación posible con el líder goblin si algún PJ habla Goblin o usa magia de traducción.
- **Resolución (15-20 min):** Los PJs reciben recompensa del alcalde en la plaza principal. Una carta misteriosa llega con un mensajero encapuchado revelando el siguiente gancho. Los NPCs del pueblo comentan sobre los héroes en las tabernas.

**Señales de Alerta:**
- **Si los jugadores están aburridos:** 1) Introducir un encuentro sorpresa con 2 bandidos que intentan robar, 2) Revelar información dramática sobre un traidor en el consejo del pueblo, 3) Cambiar de escena abruptamente con una explosión en la distancia
- **Si la sesión va muy rápida:** 1) Encuentro con mercader ambulante que vende objetos mágicos menores, 2) Puzzle ambiental en el dungeon con puertas cifradas que requieren descifrar
- **Si la sesión va muy lenta:** 1) Descripciones extensas de ambientes menores pueden resumirse, 2) Combates con minions pueden resolverse con una tirada grupal, 3) Diálogos de NPCs secundarios pueden cortarse yendo al grano

### 3. Ganchos de Personaje Obligatorios

**Por Background:**
- **Soldado:** Un antiguo camarada de la milicia aparece pidiendo ayuda contra una amenaza que conoce. O reconocés tácticas militares enemigas que revelan entrenamiento formal del oponente.
- **Criminal:** Contactos del bajo fondo te reconocen en la taberna. Podés acceder a información privilegiada sobre el mercado negro que otros no obtienen.
- **Sabio:** Textos antiguos en la biblioteca del pueblo que solo vos podés traducir correctamente. Conocimiento arcano revela secretos del ritual que están realizando los cultistas.
- **Noble:** Tu linaje te da acceso a eventos exclusivos en la mansión del Lord. Otros nobles te tratan como igual y comparten rumores de la corte.
- **Ermitaño:** Conocés rituales antiguos que otros ignoran. La naturaleza te responde con señales sobre el camino seguro a través del bosque embrujado.

**Por Clase:**
- **Guerrero:** Oportunidad de combate singular con el campeón enemigo. Demostración de fuerza abre puertas que de otra manera estarían cerradas. Los enemigos te respetan y pueden ofrecer términos.
- **Mago:** Problema mágico que requiere identificación de hechizos residuales o contramagia específica. Los glyphs antiguos responden solo a tu conocimiento arcana.
- **Pícaro:** Trampas complejas para desactivar, cerraduras mágicas para forzar, información robada de documentos sellados. Tu sigilo permite reconocimiento sin alerta.
- **Clérigo:** Muertos que necesitan ser guiados al más allá antes de que se corrompan. Maldiciones que requieren intervención divina específica de tu deidad.

**Momentos de Spotlight:**
Cada PJ debe tener al menos un momento de protagonismo en este capítulo. El pícaro desactiva la trampa que salva al grupo de caer al vacío. El clérigo reconoce el símbolo sagrado profanado y sabe cómo purificarlo. El guerrero sostiene la puerta mientras los demás escapan del colapso. El mago descifra el ritual que está debilitando la barrera mágica.

### 4. Manejo de Decisiones Críticas

**Puntos de Ramificación:**
1. **Decisión:** Los PJs ayudan a la facción A (Guardia) o B (Mercaderes)
   - **Consecuencia inmediata:** La facción elegida proporciona recursos, información y acceso exclusivo
   - **Consecuencia a largo plazo:** La facción rechazada se vuelve enemiga en el Acto N+2, cerrando quests y comercio
   - **Recovery path:** Si evitan decidir explícitamente, ambas facciones los ignoran temporalmente hasta que tomen partido

2. **Decisión:** Atacar de frente al fuerte enemigo o infiltrarse por los túneles
   - **Consecuencia inmediata:** Combate directo con ventaja táctica vs sigilo que evita peleas
   - **Consecuencia a largo plazo:** Alerta general en la región vs acceso sorpresa al jefe final
   - **Recovery path:** Si son detectados infiltrándose, combate en desventaja numérica pero con elemento sorpresa parcial

### 5. Ajustes por Tamaño de Grupo

| Grupo | Ajuste |
|-------|--------|
| **2-3 PJs** | Reducir enemigos de 4 a 2 goblins, CDs de habilidades -2, aliado NPC extra acompaña |
| **4 PJs** | Sin cambios — configuración por defecto balanceada para este encuentro |
| **5-6 PJs** | Agregar 2 goblins arqueros extra, CDs de salvación +2, el jefe tiene una acción legendaria adicional |

### 6. Tesoro y Recompensas

**Recompensas Principales:**
- 500 piezas de oro divididas del tesoro del jefe encontradas en el cofre sellado
- Objeto mágico menor: poción de curación +2 que cura 4d4+2 puntos de golpe
- Información crítica: carta que revela la ubicación del siguiente objetivo de la campaña

**Recompensas Secundarias:**
- 100 piezas de oro por completar quest secundaria opcional del molino
- Alianza con facción local que proporciona descuentos del 20% en equipo
- Acceso a comerciante especial que vende objetos mágicos no disponibles normalmente

### 7. Transición al Siguiente Capítulo

**Gancho Final:**
La carta encontrada en el cofre menciona una ubicación específica en Neverwinter donde el villano principal tiene una base. Un NPC clave revela en su lecho de muerte que el verdadero villano es alguien cercano al grupo, un giro que cambia toda la comprensión de la trama. El objeto mágico recuperado brilla con intensidad cuando se acerca a ciertos tipos de magia negra, sirviendo como detector para el próximo acto.

**Estado del Mundo Post-Capítulo:**
La facción aliada ahora controla el distrito sur del pueblo, estableciendo un gobierno provisional. El mercado negro se ha mudado al puerto más lejano para evitar la nueva vigilancia. La guardia ha aumentado patrullas nocturnas de 2 a 4 guardias por turno. Los campesinos comienzan a salir de sus casas durante el día, recuperando confianza. El templo reabre sus puertas pero con guardias armados en la entrada.`

	tests := []struct {
		name       string
		campaignID string
		actNumber  int
		title      string
		content    string
		wantErr    bool
	}{
		{
			name:       "save valid act",
			campaignID: "act-test",
			actNumber:  1,
			title:      "The Beginning",
			content:    validChapterContent,
			wantErr:    false,
		},
		{
			name:       "save act to non-existent campaign",
			campaignID: "does-not-exist",
			actNumber:  1,
			title:      "Act",
			content:    "Content",
			wantErr:    true,
		},
		{
			name:       "save act with invalid number",
			campaignID: "act-test",
			actNumber:  0,
			title:      "Invalid",
			content:    "Content",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SaveAct(tt.campaignID, tt.actNumber, tt.title, tt.content)
			if tt.wantErr {
				if err == nil {
					t.Errorf("SaveAct() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("SaveAct() unexpected error: %v", err)
				return
			}

			// Verify act was saved
			act, err := actRepo.Read(tt.campaignID, tt.actNumber)
			if err != nil {
				t.Errorf("SaveAct() act not saved: %v", err)
				return
			}
			if act.Number != tt.actNumber {
				t.Errorf("SaveAct() act number = %v, want %v", act.Number, tt.actNumber)
			}
		})
	}
}

func TestCampaignService_CompilePDF(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	tmpDir := t.TempDir()
	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, tmpDir, "echo")

	// Create a campaign first
	_, err := service.CreateCampaign("pdf-test", "PDF Test", "Setting")
	if err != nil {
		t.Fatalf("Failed to create test campaign: %v", err)
	}

	// Create minimal campaign content
	campaignDir := filepath.Join(tmpDir, "pdf-test")
	if err := os.WriteFile(filepath.Join(campaignDir, "lore.md"), []byte("# Lore\n\nTest."), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	pdfPath, err := service.CompilePDF(ctx, "pdf-test", "")
	if err != nil {
		t.Fatalf("CompilePDF() unexpected error: %v", err)
	}
	if pdfPath == "" {
		t.Error("CompilePDF() returned empty path")
	}
}

func TestCampaignService_CompilePDF_MissingCampaign(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, "/tmp/campaigns", "")

	ctx := context.Background()
	_, err := service.CompilePDF(ctx, "missing-campaign", "")
	if err == nil {
		t.Error("CompilePDF() expected error for missing campaign, got nil")
	}
}

func TestCampaignService_ListCampaigns(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, "/tmp/campaigns", "")

	// Create some campaigns
	_, _ = service.CreateCampaign("list-1", "List 1", "Setting 1")
	_, _ = service.CreateCampaign("list-2", "List 2", "Setting 2")

	campaigns, err := service.ListCampaigns()
	if err != nil {
		t.Errorf("ListCampaigns() unexpected error: %v", err)
		return
	}

	if len(campaigns) != 2 {
		t.Errorf("ListCampaigns() got %d campaigns, want 2", len(campaigns))
	}
}
