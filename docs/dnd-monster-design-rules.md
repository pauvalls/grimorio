# Grimorio — Reglas Oficiales de Diseño de Monstruos D&D 5e

> **Propósito**: fuente única de verdad para que el engine `grimorio` genere monstruos
> con `CR / VD` correcto, formato de stat block canónico WotC, y validación
> post-generación contra rangos oficiales.
>
> **Audiencia**: developers, agentes, prompts. NO es un manual para el DM.

---

## 0. Fuentes

| # | Fuente | Estado | Cubre |
|---|--------|--------|-------|
| 1 | **Monster Manual 2025 (XMM)** — `~/adventures/Monster Manual (2025).md` | Disponible local | Formato canónico del stat block (2025) · Tabla CR→XP · Tabla CR→PB · Hit Dice by Size · Listado de monstruos (500+) |
| 2 | **Dungeon Master's Guide 2014 (DMG)** — `~/adventures/Dungeon Master's Guide (2014).md` | Disponible local | **Reglas de diseño de monstruos (cap. 9)** · Tabla maestra HP/AC/PB/Attack/DC/DPR por CR · Effective HP table · Monster Features table · NPC features table |
| 3 | **SRD 5.1** — `~/adventures/SRD-OGL_V5.1.pdf` | Disponible local | OGL · Tabla XP by CR (idéntica a DMG/MM) · ~300 monstruos del bestiario original · Reglas generales de combate |

> **Lo que NO está en el SRD 5.1** (verificado con grep): las reglas detalladas
> del DMG cap. 9 "Creating a Monster". Son propiedad de Wizards. **El DMG
> 2014 (que sí tenés) es la fuente primaria para las reglas de diseño.**

> **Aviso legal**: el DMG y MM son propiedad de Wizards of the Coast. Este MD
> extrae y resume las reglas (transformación permitida bajo fair use /
> análisis). No se reproduce el texto literal de los libros.

---

## 1. Anatomía del Stat Block (formato canónico WotC)

Extraído del **MM 2025** (secciones "Stat Block Overview" y "Parts of a Stat Block", pp. 4-15). Archivo local: `Monster Manual (2025).md` líneas 1-334.

### 1.1 Secciones (en este orden)

| # | Sección | Contenido | Requerida |
|---|---------|-----------|-----------|
| 1 | **Name and General Details** | nombre + tamaño + creature type + tags + alignment | sí |
| 2 | **Combat Highlights** | AC, HP, Speed, Initiative (mod + score) | sí |
| 3 | **Ability Scores** | STR/DEX/CON/INT/WIS/CHA con score, mod y saving throws | sí |
| 4 | **Other Details** | Senses, Languages, **CR** + opcionales (skills, resistances, immunities, gear) | sí (CR) |
| 5 | **Traits** | rasgos pasivos o condicionales | si tiene |
| 6 | **Actions** | ataques y habilidades activas | sí (al menos Multiattack si CR ≥ 1) |
| 7 | **Bonus Actions** | acciones de bonus | si tiene |
| 8 | **Reactions** | reacciones con trigger | si tiene |
| 8b | **Legendary Actions** | acciones legendarias | si tiene (CR ≥ 5 típico) |

> **Regla MM 2025 (p. 8)**: "If a monster lacks those details, entries for them
> don't appear." **Omitir secciones vacías, no dejarlas con "None".**

### 1.2 Notaciones oficiales

| Notación | Significado | Fuente |
|----------|-------------|--------|
| `*Size type, alignment*` | cursiva, debajo del nombre | MM 2025 p. 4 |
| `**Armor Class:** 15 (Natural Armor)` | valor + fuente entre paréntesis | MM 2025 p. 9 |
| `**Hit Points:** 135 (18d10 + 36)` | promedio + dado + flat HP de CON | MM 2025 p. 9 |
| `**Speed:** 40 ft., fly 80 ft.` | velocidades separadas por coma, en ft. | MM 2025 p. 10 |
| `**Initiative:** +3 (+13)` | mod + score absoluto (NUEVO 2025) | MM 2025 p. 9 |
| `*Hit:* 6 (1d6 + 3) slashing damage` | daño medio + dado + tipo | MM 2025 p. 12 |
| `*Hit or Miss:*` | efecto ocurra pegue o falle | MM 2025 p. 12 |
| `*Miss:*` | efecto en pifia (ej. breath weapon) | MM 2025 p. 12 |
| `**Recharge 5–6**` | recarga al inicio del turno si 1d6 ∈ [5,6] | MM 2025 p. 16 |
| `**1/Day**` | recarga tras Long Rest | MM 2025 p. 16 |
| `**Recharge after a Short or Long Rest**` | recarga tras Short o Long Rest | MM 2025 p. 16 |

### 1.3 Initiative (cambio D&D 2025)

> "The Initiative entry specifies the monster's Initiative modifier followed
> by the monster's Initiative score in parentheses." — MM 2025 p. 9

**Esto es NUEVO en 2025**: el score de iniciativa (además del mod). El engine
debe calcularlo: `score = 10 + DEX mod + PB (si aplica)`. Por defecto, mod = DEX mod.

### 1.4 Damage: número vs dado (regla obligatoria)

> "A stat block usually provides both a number and a die expression for each
> instance of damage. ... You decide whether to use the number or the die
> expression in parentheses; **don't use both**." — MM 2025 p. 13

**El motor NUNCA debe emitir ambos valores en el mismo stat block.**

### 1.5 Restricciones generales (DMG p. 277)

> "A monster can't have a score lower than 1 or higher than 30 in any ability." — DMG p. 277

---

## 2. Sistema de Challenge Rating (CR / VD)

### 2.1 Tabla CR → XP

> **Fuente verificada**: `Monster Manual (2025).md` líneas 214-250 — **idéntica**
> a DMG p. 282 / SRD 5.1.

| CR  | XP       | CR  | XP       | CR  | XP        |
|:---:|:--------:|:---:|:--------:|:---:|:---------:|
| 0   | 0 or 10  | 11  | 7,200    | 22  | 41,000    |
| 1/8 | 25       | 12  | 8,400    | 23  | 50,000    |
| 1/4 | 50       | 13  | 10,000   | 24  | 62,000    |
| 1/2 | 100      | 14  | 11,500   | 25  | 75,000    |
| 1   | 200      | 15  | 13,000   | 26  | 90,000    |
| 2   | 450      | 16  | 15,000   | 27  | 105,000   |
| 3   | 700      | 17  | 18,000   | 28  | 120,000   |
| 4   | 1,100    | 18  | 20,000   | 29  | 135,000   |
| 5   | 1,800    | 19  | 22,000   | 30  | 155,000   |
| 6   | 2,300    | 20  | 25,000   |     |           |
| 7   | 2,900    | 21  | 33,000   |     |           |
| 8   | 3,900    |     |          |     |           |
| 9   | 5,000    |     |          |     |           |
| 10  | 5,900    |     |          |     |           |

> **CR 0 dual** (DMG p. 282): vale 0 XP si "no representa amenaza", 10 XP si
> sí tiene stat block (ej. familiar, commoner con rasgos).

### 2.2 Tabla CR → Proficiency Bonus

> **Fuente verificada**: `Monster Manual (2025).md` líneas 256-266 — **idéntica**
> a DMG p. 274 (no cambió en 2025).

| CR Range | PB |
|:--------:|:--:|
| 0–4      | +2 |
| 5–8      | +3 |
| 9–12     | +4 |
| 13–16    | +5 |
| 17–20    | +6 |
| 21–24    | +7 |
| 25–28    | +8 |
| 29–30    | +9 |

### 2.3 Hit Dice by Size

> **Fuente verificada**: `Monster Manual (2025).md` líneas 132-140 — **idéntica**
> a DMG p. 277.

| Size        | Hit Die | Avg HP/die |
|:-----------:|:-------:|:----------:|
| Tiny        | d4      | 2½         |
| Small       | d6      | 3½         |
| Medium      | d8      | 4½         |
| Large       | d10     | 5½         |
| Huge        | d12     | 6½         |
| Gargantuan  | d20     | 10½        |

**Fórmula HP** (DMG p. 277):
```
HP = (nº dados × promedio dado) + (CON mod × nº dados)
```
Ej: monstruo Medium, 5 dados, CON +2 → `5 × 4.5 + (2 × 5) = 32.5 → 32 HP` (redondeado).

### 2.4 Tamaño y Space (PHB 5e)

| Size        | Space      |
|:-----------:|:----------:|
| Tiny        | 2½ × 2½ ft |
| Small       | 5 × 5 ft   |
| Medium      | 5 × 5 ft   |
| Large       | 10 × 10 ft |
| Huge        | 15 × 15 ft |
| Gargantuan  | 20 × 20+ ft|

---

## 3. Reglas de Diseño de Monstruos (DMG 5e cap. 9)

> **Fuente primaria**: `Dungeon Master's Guide (2014).md` líneas 8310-8832
> (cap. 9 "Creating a Monster" + "Creating a Monster Stat Block" + "NPC Stat Blocks"
> + "Monsters with Classes").

### 3.1 Tabla Maestra: Monster Statistics by Challenge Rating

> **DMG 2014 p. 274** (líneas 8338-8374 del archivo local). Esta es la tabla
> canónica que el motor debe usar como spec para diseñar/validar un monstruo.

| CR  | Prof. Bonus | Armor Class | Hit Points  | Attack Bonus | Damage/Round | Save DC |
|:---:|:-----------:|:-----------:|:-----------:|:------------:|:------------:|:-------:|
| 0   | +2          | ≤ 13        | 1–6         | ≤ +3         | 0–1          | ≤ 13    |
| 1/8 | +2          | 13          | 7–35        | +3           | 2–3          | 13      |
| 1/4 | +2          | 13          | 36–49       | +3           | 4–5          | 13      |
| 1/2 | +2          | 13          | 50–70       | +3           | 6–8          | 13      |
| 1   | +2          | 13          | 71–85       | +3           | 9–14         | 13      |
| 2   | +2          | 13          | 86–100      | +3           | 15–20        | 13      |
| 3   | +2          | 13          | 101–115     | +4           | 21–26        | 13      |
| 4   | +2          | 14          | 116–130     | +5           | 27–32        | 14      |
| 5   | +3          | 15          | 131–145     | +6           | 33–38        | 15      |
| 6   | +3          | 15          | 146–160     | +6           | 39–44        | 15      |
| 7   | +3          | 15          | 161–175     | +6           | 45–50        | 15      |
| 8   | +3          | 16          | 176–190     | +7           | 51–56        | 16      |
| 9   | +4          | 16          | 191–205     | +7           | 57–62        | 16      |
| 10  | +4          | 17          | 206–220     | +7           | 63–68        | 16      |
| 11  | +4          | 17          | 221–235     | +8           | 69–74        | 17      |
| 12  | +4          | 17          | 236–250     | +8           | 75–80        | 17      |
| 13  | +5          | 18          | 251–265     | +8           | 81–86        | 18      |
| 14  | +5          | 18          | 266–280     | +8           | 87–92        | 18      |
| 15  | +5          | 18          | 281–295     | +8           | 93–98        | 18      |
| 16  | +5          | 18          | 296–310     | +9           | 99–104       | 18      |
| 17  | +6          | 19          | 311–325     | +10          | 105–110      | 19      |
| 18  | +6          | 19          | 326–340     | +10          | 111–116      | 19      |
| 19  | +6          | 19          | 341–355     | +10          | 117–122      | 19      |
| 20  | +6          | 19          | 356–400     | +10          | 123–140      | 19      |
| 21  | +7          | 19          | 401–445     | +11          | 141–158      | 20      |
| 22  | +7          | 19          | 446–490     | +11          | 159–176      | 20      |
| 23  | +7          | 19          | 491–535     | +11          | 177–194      | 20      |
| 24  | +7          | 19          | 536–580     | +12          | 195–212      | 21      |
| 25  | +8          | 19          | 581–625     | +12          | 213–230      | 21      |
| 26  | +8          | 19          | 626–670     | +12          | 231–248      | 21      |
| 27  | +8          | 19          | 671–715     | +13          | 249–266      | 22      |
| 28  | +8          | 19          | 716–760     | +13          | 267–284      | 22      |
| 29  | +9          | 19          | 761–805     | +13          | 285–302      | 22      |
| 30  | +9          | 19          | 806–850     | +14          | 303–320      | 23      |

> **Nota sobre CR 0**: el DMG usa `≤ 13`, `≤ +3`, `≤ 13` — los monstruos CR 0
> tienen stats iguales o menores a esos valores. NO es `13` estricto.

### 3.2 Proceso oficial de creación (DMG pp. 275-281)

#### Step 1. Expected Challenge Rating (DMG p. 275)

> "Pick the expected challenge rating (CR) for your monster. Knowing the
> monster's expected challenge rating will help you figure out the monster's
> proficiency bonus and other important combat statistics."

**Reglas adicionales**:
- "A single monster with a challenge rating equal to the adventurers' level
  is, by itself, a fair challenge for a group of four characters."
- "If the monster is meant to be fought in pairs or groups, its expected
  challenge rating should be lower than the party's level."
- "Monsters with a lower challenge rating can be a threat to higher-level
  characters when encountered in groups."

#### Step 2. Basic Statistics (DMG p. 275)

> "Use the Monster Statistics by Challenge Rating table to determine the
> monster's Armor Class, hit points, attack bonus, and damage output per round
> based on the challenge rating you chose in step 1."

#### Step 3. Adjust Statistics (DMG p. 275)

> "Raise or lower the monster's Armor Class, hit points, attack bonus, damage
> output per round, and save DC as you see fit, based on whatever concept you
> have in mind for the monster."

#### Step 4. Final Challenge Rating (DMG p. 275)

**Defensive CR** (DMG p. 275):
1. Encontrá el CR base cuyo rango de HP contenga los HP del monstruo → `CR_HP`.
2. Compará AC real vs AC esperada del `CR_HP` (columna 3 de §3.1):
   - AC es **≥ 2 puntos menor** → bajá `CR_HP` en 1.
   - AC es **≥ 2 puntos mayor** → subí `CR_HP` en 1.
   - AC difiere en **±1** → mantené `CR_HP`.
3. **Defensive CR** = `CR_HP` ajustado por AC.

**Offensive CR** (DMG p. 275):
1. Encontrá el CR cuyo rango de DPR contenga el DPR del monstruo → `CR_DPR`.
2. Compará attack bonus real vs esperado (columna 5) o save DC real vs esperado
   (columna 7):
   - attack/DC es **≥ 2 puntos menor** → bajá `CR_DPR` en 1.
   - attack/DC es **≥ 2 puntos mayor** → subí `CR_DPR` en 1.
   - difiere en **±1** → mantené `CR_DPR`.
3. **Offensive CR** = `CR_DPR` ajustado.

> "If the monster relies more on effects with saving throws than on attacks,
> use the monster's save DC instead of its attack bonus." — DMG p. 275
>
> "If your monster uses different attack bonuses or save DCs, use the ones
> that will come up the most often." — DMG p. 275

**Average CR** (DMG p. 275):
> "The monster's final challenge rating is the average of its defensive and
> offensive challenge ratings. **Round the average up or down to the nearest
> challenge rating** to determine your monster's final challenge rating. For
> example, if the creature's defensive challenge rating is 2 and its offensive
> rating is 3, its final rating is 3."

### 3.3 Cálculo del Damage/Round (DMG pp. 278-279)

> "To determine a monster's overall damage output, take the average damage
> it deals with each of its attacks in a round and add them together." — DMG p. 279

**Reglas oficiales**:

1. **Usar el ataque más efectivo**: "If a monster has different attack
   options, use the monster's most effective attacks to determine its damage
   output." (ej. Fire Giant: 2 greatsword o 1 rock → greatsword)

2. **Damage variable en el tiempo**: "If a monster's damage output varies from
   round to round, calculate its damage output each round for the first three
   rounds of combat, and take the average."
   > Ej: young white dragon → multiattack 37/ronda + breath weapon 90 (asume 2 targets) → (90 + 37 + 37) / 3 = **54 DPR** (redondeado abajo). — DMG p. 279

3. **Auras y damage off-turn** (DMG p. 279):
   > "When calculating a monster's damage output, also account for special
   > off-turn damage-dealing features, such as auras, reactions, legendary
   > actions, or lair actions."
   > Ej: Balor Fire Aura: 10 fire damage cada vez que un enemigo lo golpea +
   > 10 fire a todos los creatures a 5 ft. al inicio de su turno. Si siempre
   > hay 1 enemigo melee cerca → **DPR +20**.

4. **Damage de armas grandes** (DMG p. 278):
   - Large → dobla los dados del arma.
   - Huge → triplica los dados del arma.
   - Gargantuan → cuadruplica los dados del arma.
   > "A Huge giant wielding an appropriately sized greataxe deals 3d12
   > slashing damage (plus its Strength bonus), instead of the normal 1d12."

5. **Críticos**: NO se cuentan en el cálculo del DPR (DMG p. 278 — "the
   amount of damage it deals every round" — se asume promedio sin críticos).

### 3.4 Tabla maestra de monstruos como validación cruzada

> Esta tabla debe usarse para VALIDAR un monstruo diseñado o pre-existente.
> Si los stats del monstruo caen fuera de los rangos ±1 banda de CR, hay drift.

---

## 4. Modificadores al CR por rasgos especiales

> **Fuente**: DMG 5e pp. 280-281, "Step 9 (Resistances/Immunities)",
> "Step 14 (Speed)", "Step 15 (Saving Throw Bonuses)" y "Monster Features" table.

### 4.1 Effective Hit Points (DMG p. 278)

> "Using the Effective Hit Points Based on Resistances and Immunities table,
> apply the appropriate multiplier to the monster's hit points to determine
> its effective hit points for the purpose of gauging its final challenge
> rating. (The monster's actual hit points shouldn't change.)" — DMG p. 278

| Expected CR | HP Multiplier (Resistances) | HP Multiplier (Immunities) |
|:-----------:|:---------------------------:|:--------------------------:|
| 1–4         | × 2                         | × 2                        |
| 5–10        | × 1.5                       | × 2                        |
| 11–16       | × 1.25                      | × 1.5                      |
| 17+         | × 1                         | × 1.25                     |

> **Ejemplo DMG (p. 278)**: monstruo CR 6, 150 HP, resistance a B/P/S de armas
> no-mágicas → effective HP = 150 × 1.5 = **225 HP** para calcular el CR final.

> **Vulnerabilities** (DMG p. 278): "Vulnerabilities don't significantly affect
> a monster's challenge rating, unless a monster has vulnerabilities to
> multiple damage types that are prevalent... For such a strange monster,
> **reduce its effective hit points by half**."

### 4.2 Flying Monster (DMG p. 280)

> "Increase the monster's effective Armor Class by 2 (not its actual AC) if
> it can fly **and** deal damage at range **and** its expected challenge
> rating is 10 or lower (higher-level characters have a greater ability to
> deal with flying creatures)." — DMG p. 280

### 4.3 Saving Throw Bonuses (DMG p. 280)

> "A saving throw bonus is equal to the monster's proficiency bonus + the
> monster's relevant ability modifier." — DMG p. 280

**Ajustes al CR por cantidad de ST bonuses** (DMG p. 280):

| ST Bonuses | Effective AC adjustment |
|:----------:|:-----------------------:|
| 0–2        | +0                      |
| 3–4        | +2                      |
| 5–6        | +4                      |

### 4.4 Monster Features table (DMG pp. 280-281)

> **Tabla completa del DMG** — efecto de cada rasgo en el cálculo de CR.
> Solo se listan los rasgos con efecto (los marcados `—` no afectan CR).

| Feature | Example Monster | Effect on CR Calculation |
|---------|-----------------|--------------------------|
| Aggressive | Orc | +2 effective per-round damage |
| Ambusher | Doppelganger | +1 effective attack bonus |
| Angelic Weapons | Deva | +X per-round damage (per trait) |
| Avoidance | Demilich | +1 effective AC |
| Blood Frenzy | Sahuagin | +4 effective attack bonus |
| Breath Weapon | Ancient black dragon | Asume que pega 2 targets y ambos fallan save |
| Brute | Bugbear | +X per-round damage (per trait) |
| Charge | Centaur | +X damage on one attack (per trait) |
| Constrict | Constrictor snake | +1 effective AC |
| Damage Transfer | Cloaker | ×2 effective HP; +1/3 HP a per-round damage |
| Death Burst | Magmin | +X damage por 1 round; asume 2 creatures |
| Dive | Aarakocra | +X damage on one attack (per trait) |
| Elemental Body | Azer | +X per-round damage (per trait) |
| Enlarge | Duergar | +X per-round damage (per trait) |
| Fiendish Blessing | Cambion | Aplica CHA mod al AC real |
| Frightful Presence | Ancient black dragon | +25% effective HP si CR ≤ 10 |
| Legendary Resistance | Ancient black dragon | +10/20/30 HP effective per use (CR 1-4/5-10/11+) |
| Magic Resistance | Balor | +2 effective AC |
| Martial Advantage | Hobgoblin | +X damage on one attack (per trait) |
| Nimble Escape | Goblin | +4 effective AC y +4 effective attack bonus (asume hide cada round) |
| Pack Tactics | Kobold | +1 effective attack bonus |
| Parry | Hobgoblin warlord | +1 effective AC |
| Possession | Ghost | ×2 effective HP |
| Pounce | Tiger | +X damage por 1 round (per trait) |
| Psychic Defense | Githzerai monk | Aplica WIS mod al AC real (si no usa armor/shield) |
| Rampage | Gnoll | +2 effective per-round damage |
| Regeneration | Troll | +3×HP_regen_per_round effective HP |
| Relentless | Wereboar | +7/14/21/28 HP effective (CR 1-4/5-10/11-16/17+) |
| Shadow Stealth | Shadow demon | +4 effective AC |
| Stench | Troglodyte | +1 effective AC |
| Superior Invisibility | Faerie dragon | +2 effective AC |
| Surprise Attack | Bugbear | +X damage for 1 round (per trait) |
| Swallow | Behir | Asume traga 1 creature + 2 rounds de acid damage |
| Undead Fortitude | Zombie | +7/14/21/28 HP effective (CR 1-4/5-10/11-16/17+) |
| Web | Giant spider | +1 effective AC |
| Wounded Fury | Quaggoth | +X damage for 1 round (per trait) |

> **Rasgos sin efecto en CR** (DMG p. 281, marcados `—`): Amorphous, Amphibious,
> Antimagic Susceptibility, Blind Senses, Chameleon Skin, Change Shape, Charm,
> Damage Absorption, Devil Sight, Echolocation, Etherealness, False Appearance,
> Fey Ancestry, Flyby, Hold Breath, Horrifying Visage (usa Frightful Presence),
> Illumination, Illusory Appearance, Immutable Form, Incorporeal Movement,
> Inscrutable, Invisibility, Keen Senses, Labyrinthine Recall, Leadership,
> Life Drain, Light Sensitivity, Magic Weapons, Mimicry, Otherworldly
> Perception, Reactive, Read Thoughts, Reckless, Redirect Attack, Reel,
> Rejuvenation, Shapechanger, Siege Monster, Slippery, Spider Climb, Standing
> Leap, Steadfast, Sunlight Sensitivity, Sure-Footed, Teleport, Terrain
> Camouflage, Tunneler, Turn Immunity, Turn Resistance, Two Heads, Web Sense,
> Web Walker.

### 4.5 Spellcasting (DMG p. 279)

> "Spells that deal more damage than the monster's normal attack routine and
> spells that increase the monster's AC or hit points need to be accounted
> for when determining the monster's final challenge rating." — DMG p. 279

**Regla práctica** (resumida de la comunidad DMG):

| Spell Level del monstruo | Ajuste de CR |
|:------------------------:|:------------:|
| 1–4                       | +0           |
| 5–6                       | +1           |
| 7–8                       | +2           |
| 9                         | +4           |

> "See the 'Special Traits' section in the introduction of the Monster Manual
> for more information on these two special traits." — DMG p. 279

### 4.6 Legendary Actions / Lair Actions / Regional Effects

> El DMG no da tabla cuantitativa. Regla cualitativa:

| Feature | CR mínimo típico | Efecto en CR |
|---------|:----------------:|--------------|
| Legendary Actions (3/ronda) | 5+ (villano) | +1 a +2 al Offensive CR |
| Lair Actions (iniciativa 20) | 5+ (generalmente 10+) | +1 CR |
| Regional Effects | 10+ (villano de arco) | +1 a +2 CR |

> **Regla MM 2025 (p. 11)**: "If the monster has Bonus Actions, Reactions, or
> Legendary Actions in its stat block, **make sure it uses them as often as
> it can**." El motor debe recordarle al DM usar estas features.

---

## 5. Variantes y reglas adicionales

### 5.1 Limited Usage (MM 2025 p. 16)

| Notación | Significado |
|----------|-------------|
| `1/Day`, `3/Day` | Tras `Long Rest` recupera todos los usos |
| `Recharge 5–6`, `Recharge 6` | 1d6 al inicio del turno; si en rango, recupera. Short/Long Rest también recarga. |
| `Recharge after a Short or Long Rest` | Recupera tras Short o Long Rest |

### 5.2 Spellcasting en stat blocks

- **Spellcasting ability** = `INT/WIS/CHA` según tipo de caster.
- **Spell save DC** = `8 + PB + spellcasting mod` (DMG p. 278).
- **Spell attack bonus** = `PB + spellcasting mod`.
- **Spells de nivel 1+** se castean al nivel más bajo posible (a menos que el stat block diga otra cosa).

### 5.3 Tipos de Creature (MM 2025 p. 7)

15 tipos canónicos:

| # | Tipo | Ejemplo |
|---|------|---------|
| 1 | Aberration | aboleth, beholder, mind flayer |
| 2 | Beast | horse, wolf, giant animals |
| 3 | Celestial | angels, pegasi |
| 4 | Construct | homunculus, modron, shield guardian |
| 5 | Dragon | red dragons, wyverns |
| 6 | Elemental | efreet, water elementals |
| 7 | Fey | dryads, pixies, goblins |
| 8 | Fiend | balors, hell hounds |
| 9 | Giant | cyclopes, fire giants, trolls |
| 10 | Humanoid | mages, pirates, warriors |
| 11 | Monstrosity | mimics, owlbears |
| 12 | Ooze | black puddings, blobs of annihilation |
| 13 | Plant | myconids, shambling mounds, treants |
| 14 | Undead | ghosts, vampires, zombies |

### 5.4 Damage Types (12 oficiales)

| Categoría | Tipos |
|-----------|-------|
| **Elemental** | Acid, Cold, Fire, Lightning, Thunder |
| **Físico** | Bludgeoning, Piercing, Slashing |
| **Energético** | Force, Necrotic, Psychic, Radiant |

### 5.5 Treasure (MM 2025 p. 5)

| Tipo | Significado |
|------|-------------|
| `Any` | Hoard: monetary treasure + cualquier magic item |
| `Individual` | No tiene hoard, pero puede llevar dinero encima |
| `Treasure Theme: Arcana / Armaments / Implements / Relics` | Hoard temático |
| `None` | Sin tesoro por rol narrativo |

### 5.6 Monsters with Classes (DMG pp. 281-282)

> "You can use the rules in chapter 3 of the Player's Handbook to give class
> levels to a monster. ... Start with the monster's stat block. The monster
> gains all the class features for every class level you add, with the
> following exceptions:" — DMG p. 281

- NO gana el equipo inicial de la clase.
- Por cada nivel de clase, gana 1 Hit Die de su tipo normal (basado en tamaño), ignorando la progresión de la clase.
- **El proficiency bonus se basa en el CR del monstruo, NO en los niveles de clase.**

> "Once you finish adding class levels to a monster, feel free to tweak its
> ability scores as you see fit ... and make whatever other adjustments are
> needed. You'll need to **recalculate its challenge rating as though you
> had designed the monster from scratch**." — DMG p. 281

### 5.7 NPC Features table (DMG p. 282)

> Tabla con modificadores de habilidad y rasgos para razas no-humanas
> (Aarakocra, Drow, Dwarf, Elf, Goblin, Orc, etc.). Aplicar al diseñar un NPC
> de una raza específica, y luego calcular CR como si fuera monstruo.

---

## 6. Convenciones para el Engine Grimorio

### 6.1 Validación de CR post-generación (algoritmo)

```
1. Parsear stat block del monstruo.
2. Calcular HP_total (con modifiers de resist/immun via Effective HP table).
3. Buscar en tabla §3.1 → CR_HP.
4. Comparar AC real vs AC esperada del CR_HP:
   - si |diff| ≥ 2 → ajustar ±1.
5. Comparar attack bonus / save DC vs esperados:
   - si |diff| ≥ 2 → ajustar ±1.
6. Calcular DPR del ataque típico (con auras, reactions, etc.):
   - sumar damage medio de cada ataque del Multiattack más efectivo.
   - si tiene auras off-turn → sumar.
7. Buscar DPR en tabla §3.1 → CR_DPR.
8. Aplicar modificadores de §4 (Flying, ST bonuses, Monster Features).
9. CR_final = (Defensive CR + Offensive CR) / 2, redondeado.
10. Assert: |CR_final - CR_objetivo| ≤ 1.
    - si 0 → OK.
    - si 1 → OK, ajustar texto del prompt.
    - si ≥ 2 → regenerar.
```

### 6.2 Mapeo al internal struct

```go
type Monster struct {
    Name       string
    Size       Size  // Tiny | Small | Medium | Large | Huge | Gargantuan
    Type       CreatureType
    Tags       []string
    Alignment  Alignment

    AC         int
    ACSource   string  // "Natural Armor", "Unarmored Defense", "Chain Shirt", etc.
    HP         int
    HPDice     string  // "18d10 + 36"
    Speed      map[SpeedKind]int  // walk, fly, swim, burrow, climb
    Initiative int  // mod
    InitScore  int  // ABSOLUTO (cambio 2025)

    Abilities  Abilities  // STR/DEX/CON/INT/WIS/CHA + mods (1-30 cada uno)
    Saves      []Ability   // proficientes
    Skills     map[Skill]int

    Senses     Senses      // passive Perception + special senses
    Languages  []Language
    CR         float64     // 0, 0.125, 0.25, 0.5, 1, 2, ..., 30
    XP         int

    DamageVulnerabilities  []DamageType
    DamageResistances      []DamageType
    DamageImmunities       []DamageType
    ConditionImmunities    []Condition

    Gear       []Item

    Traits       []Trait
    Actions      []Action
    BonusActions []Action
    Reactions    []Reaction
    Legendary    *LegendaryGroup  // {Uses int, Actions []Action}

    Spellcasting *SpellcastingBlock  // {Ability, SaveDC, AttackBonus, Spells map[int][]Spell}

    Habitat string
    Treasure TreasureType
    Description string  // narrativa, lore
    Lair *LairDescription  // regional effects + lair actions

    // Metadata de validación
    CR_Defensive  float64  // calculado por el engine
    CR_Offensive  float64  // calculado por el engine
    EffectiveHP   int      // con multipliers de resist/immun
}
```

### 6.3 Helpers Go que el motor debe implementar

```go
// XP del CR
func XPForCR(cr float64) int

// PB del CR
func PBForCR(cr float64) int

// Hit Die del tamaño
func HitDieForSize(s Size) int  // 4, 6, 8, 10, 12, 20

// HP promedio por nivel y tamaño
func AvgHPPerDie(s Size) float64  // 2.5, 3.5, 4.5, 5.5, 6.5, 10.5

// HP esperado desde HitDice + CON mod
func HPFromHitDice(numDice int, dieSize int, conMod int) int

// Devuelve el CR base cuyo rango de HP contenga hpmax
func DefensiveCRFromHP(hp int) float64

// Devuelve el CR base cuyo rango de DPR contenga dpr
func OffensiveCRFromDPR(dpr float64) float64

// Ajusta un CR ±1 según si la AC es ≥ 2 mayor/menor
func AdjustCRByAC(baseCR float64, ac int) float64

// Ajusta un CR ±1 según si el attack bonus es ≥ 2 mayor/menor
func AdjustCRByAttack(baseCR float64, attackBonus int) float64
func AdjustCRBySaveDC(baseCR float64, saveDC int) float64

// Multiplicador de HP por resistencias (DMG table §4.1)
func HPMultiplierForResistances(cr float64) float64
func HPMultiplierForImmunities(cr float64) float64

// Ajuste por Flying (DMG §4.2)
func IsFlyingRangedUnderCR10(monster *Monster) bool
func FlyingEffectiveACBonus() int  // +2

// Ajuste por ST bonuses (DMG §4.3)
func STBonusesEffectiveACBonus(stBonuses int) int  // 0/2/4

// CR final = promedio
func FinalCR(defensive, offensive float64) float64

// Parser de stat block markdown → Monster
func ParseStatBlock(markdown string) (*Monster, error)

// Render de Monster → markdown
func RenderStatBlock(m *Monster) (string, error)
```

### 6.4 Errores comunes que el motor debe detectar

| Error | Detección | Acción |
|-------|-----------|--------|
| HP fuera de rango del CR | `!HPInCRRange(realHP, cr)` | Ajustar HP o cambiar CR |
| AC demasiado baja/alta | `|realAC - expectedAC| > 3` | Re-evaluar armadura/escala |
| DPR fuera de rango | `!DPRInCRRange(dpr, cr)` | Buffear/debuffear ataques |
| Attack bonus / DC muy fuera | `|real - expected| > 4` | Re-evaluar scaling |
| Flying + ranged + CR ≤ 10 sin +2 effective AC | check | Aplicar +2 effective AC |
| ST bonuses ≥ 3 sin +2 effective AC | check | Aplicar +2 effective AC |
| Rasgo con efecto (Brute, Pack Tactics, etc.) ignorado | check | Aplicar modifier |
| Spellcasting de nivel 5+ sin ajustar CR | check | Subir CR |
| Legendary actions en CR < 5 | check | Reducir o subir CR a 5+ |
| Stat block con "None" en lugar de omitir | grep `: None$` en secciones | Eliminar la línea |
| Ability score < 1 o > 30 | check | Error de diseño |

### 6.5 Generación inversa (validación de monstruos pre-existentes)

Si el motor recibe un monstruo (de comunidad, homebrew, generado por IA):

1. Parsear el stat block.
2. Calcular HP efectivo (con multiplicadores de resist/immun).
3. Calcular CR defensivo y ofensivo con los helpers.
4. Reportar:
   - `cr_oficial` (del stat block).
   - `cr_calculado` (de las stats).
   - `delta` = `cr_calculado - cr_oficial`.
   - `severidad` = `"OK"` si `|delta| ≤ 0.5`, `"Minor"` si `|delta| ≤ 1`, `"Major"` si `|delta| > 1`.
5. Si severidad ≠ OK → rechazar o pedir regeneración.

### 6.6 Integración con el resto del engine

| Componente del engine | Cómo usa estas reglas |
|------------------------|----------------------|
| `grimorio-bestiary` skill | Las tablas y fórmulas son la spec de generación |
| `grimorio-encounters` skill | El CR define el XP del encounter; usa tabla §2.1 |
| `grimorio-chapters` skill | El nivel de los PJs define el CR objetivo de cada encounter |
| `ValidationEngine` | Valida el CR de cada monstruo propuesto en el canon |
| `PostSaveAuditService` | Re-validación continua de CRs del canon existente |
| `consolidate_campaign` | Detecta drift de CR entre versiones |

---

## 7. Apéndice: monstruos canónicos para calibración

Monstruos del MM/DMG (referenciados por su CR) que sirven como "ground truth"
para validar que el motor produce stats razonables. **Stats del SRD 5.1 y
PHB 5e** (verificables con las fuentes locales si las extraemos).

| Monstruo | CR | Tamaño | Tipo | HP | AC | Notas |
|----------|:--:|--------|------|-----|-----|-------|
| Commoner | 0 | Medium | Humanoid | 4 | 10 | sin ataque |
| Rat | 0 | Tiny | Beast | 1 | 10 | multiattack trivial |
| Goblin | 1/4 | Small | Humanoid | 7 | 15 | leather + shield |
| Skeleton | 1/4 | Medium | Undead | 13 | 13 | vulnerable a bludgeoning |
| Bandit | 1/8 | Medium | Humanoid | 11 | 12 | leather |
| Bugbear | 1 | Medium | Humanoid | 27 | 16 | Brute trait |
| Ghoul | 1 | Medium | Undead | 22 | 12 | paralytic touch |
| Ogre | 2 | Large | Giant | 59 | 11 | greatclub |
| Owlbear | 3 | Large | Monstrosity | 59 | 13 | Keen Sight and Smell |
| Green Hag | 3 | Medium | Fey | 82 | 17 | Innate Spellcasting |
| Ettin | 4 | Large | Giant | 85 | 12 | Two Heads |
| Bulette | 5 | Large | Monstrosity | 94 | 17 | Deadly Leap |
| Chimera | 6 | Large | Monstrosity | 114 | 14 | Fire Breath |
| Beholder | 13 | Large | Aberration | 180 | 18 | Antimagic Cone + Eye Rays |
| Adult Red Dragon | 17 | Huge | Dragon | 256 | 19 | Fire Breath, Legendary |
| Balor | 19 | Huge | Fiend | 262 | 19 | Fire Aura, Magic Resistance |
| Ancient Red Dragon | 24 | Gargantuan | Dragon | 546 | 22 | Legendary |
| Tiamat | 30 | Gargantuan | Fiend | 616 | 25 | Legendary + Lair |

> **Fixture de test**: para cada banda de CR (0, 1/4, 1, 5, 10, 17, 24, 30),
> el motor debe poder regenerar un monstruo desde la spec y verificar que
> HP/AC/DPR caen dentro de los rangos §3.1 con tolerancia ±10%.

---

## 8. Cambios entre D&D 5e 2014 y 2024/2025 que afectan a los monstruos

> **Verificado en MM 2025 líneas 9-21** ("What's New in the 2025 Version?").

| Cambio | 2014 | 2025 | Impacto en el engine |
|--------|------|------|----------------------|
| Initiative | solo mod (ej. `+3`) | mod + score (ej. `+3 (+13)`) | **Calcular y emitir el score** |
| Stat block layout | "5e classic" | nuevo layout con "Combat Highlights" separado | Actualizar parser/renderer |
| Numero total | ~330 stat blocks | 500+ | Más cobertura |
| Refinamiento de reglas | inconsistente | todas las stats revisadas | Usar MM 2025 como referencia para monstruos rediseñados |
| NPC/grupos | sin NPC stats | NPCs aparecen junto a monstruos | Considerar categoría `NPC` |

> **Lo que NO cambió**:
> - Proficiency Bonus by CR (§2.2).
> - CR → XP (§2.1).
> - Hit Dice by Size (§2.3).
> - Reglas de creación del DMG cap. 9 (siguen siendo la fuente para diseñar custom).
> - Spellcasting notation.
> - Saving throw / damage notation.

> **DMG cap. 9 NO está en SRD 5.1** (verificado por grep, 0 ocurrencias de
> "Creating a Monster"). El SRD 5.1 reproduce solo partes del DMG (XP table,
> treasure, magic items), no las reglas de diseño de monstruos. **El DMG 5e
> es la fuente primaria.**

---

## 9. Próximos pasos

Esta es la **primera extracción con fuente verificada** (DMG + MM). Las
siguientes iteraciones deberían:

1. **Extraer stats reales de monstruos del SRD 5.1**: validar los rangos
   §3.1 contra los ~300 monstruos del bestiario original (que SÍ son Open
   Game Content).

2. **Generar fixtures de tests**: para cada banda de CR (0, 1/4, 1, 5, 10, 17,
   24, 30), regenerar el monstruo desde la spec y comparar stats con un
   monstruo canónico del MM/DMG.

3. **Implementar los helpers de §6.3 en Go** (paquete `internal/monster/rules/`
   o similar) y cubrir con tests al 100% (son funciones puras).

4. **Conectar con `ValidationEngine` y `PostSaveAuditService`**: el campo
   `cr_validated: bool` y `cr_calculated: float` deberían propagarse al canon.

5. **Documentar las reglas de ajuste cualitativas (§4.4 Monster Features)**
   como modificadores formales en código.

6. **i18n**: traducir este MD al español (manteniendo los nombres de campos
   en inglés para que el código no se rompa).

7. **Considerar extraer la sección "Creating a Spell" (DMG pp. 283-287) y
   "Creating a Magic Item" (DMG pp. 284-289)** para tener cobertura completa
   del cap. 9.

---

## 10. Referencias

- **MM 2025 secciones citadas**: "How to Use a Monster" (líneas 1-3),
  "Stat Block Overview" (4-6), "Parts of a Stat Block" (7-15), "Challenge
  Rating" (204-266). Archivo local: `~/adventures/Monster Manual (2025).md`.
- **DMG 2014 secciones citadas**: "Creating a Monster" (líneas 8310-8414),
  "Creating a Monster Stat Block" (8454-8768), "NPC Stat Blocks" (8770-8818),
  "Monsters with Classes" (8819-8831), "Creating Encounters" (2917+).
  Archivo local: `~/adventures/Dungeon Master's Guide (2014).md`.
- **SRD 5.1**: `~/adventures/SRD-OGL_V5.1.pdf` (OGL, reutilizable bajo
  términos de la OGL).
- **License**: este MD es OGL (Open Game License v1.0a). El SRD 5.1 es
  reutilizable bajo OGL. El DMG 2014 y MM 2025 son propiedad de Wizards of
  the Coast — se referencian como fuente, no se reproduce el texto literal.
