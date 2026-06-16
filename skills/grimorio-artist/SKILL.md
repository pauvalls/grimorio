---
name: grimorio-artist
version: "1.0.0"
description: Prepare AI image specifications and update markdown references for NPC portraits, monster illustrations, and scene artwork
---

# grimorio-artist — Artist

## Purpose

Generate AI image specifications and update markdown references for:
- NPC portraits
- Monster illustrations
- Scene artwork
- Campaign cover

**IMPORTANT:** Image generation is ALWAYS sequential with a 3-second delay between images to avoid rate limiting.

## Available Tools

**MCP Tools:**
- `generate_image` — Generate AI images (portraits, illustrations, scenes, covers)
- `generate_map` — Generate SVG maps (use from cartographer)
- `generate_divider` — Generate decorative SVG dividers

**System Tools:**
- `Read` — Read campaign files
- `Write` — Write batch-spec.json and update markdowns
- `Bash` — List generated images
- `Grep` — Search references in markdowns
- `Edit` — Update image references

## Mandatory Workflow

### Phase A: Prepare Batch Specification

**Step 1: Read ALL source files**

```python
# Read in order
read("{campaign_path}/canon.json")           # Visual canonical facts
read("{campaign_path}/npcs/npcs_and_factions.md")  # NPCs for portraits
read("{campaign_path}/bestiary/bestiary.md")  # Monsters for illustrations
read("{campaign_path}/acts/*.md")            # Scenes marked with [SCENE: ...]
read("{campaign_path}/lore/lore.md")         # Setting/tone for cover
```

**IMPORTANT:** Check canon.json first. If it establishes visual facts (e.g., "the city is underwater", "all vampires have silver hair"), INCORPORATE those details in the prompts.

**Step 2: Build batch-spec.json**

Create `{campaign_path}/assets/batch-spec.json`:

```json
{
  "campaign": "campaign-name",
  "images": [
    {
      "filename": "cover-art",
      "prompt": "Epic D&D fantasy cover art, [setting description], cinematic, highly detailed, dramatic lighting, professional digital painting",
      "type": "cover"
    },
    {
      "filename": "npc-[kebab-case-name]",
      "prompt": "D&D character portrait, [race/class/description], [personality traits], detailed fantasy art style, professional illustration",
      "type": "portrait"
    },
    {
      "filename": "monster-[kebab-case-name]",
      "prompt": "D&D monster illustration, [creature type/description], menacing pose, detailed fantasy art, dramatic lighting",
      "type": "illustration"
    },
    {
      "filename": "scene-[act]-[kebab-case-description]",
      "prompt": "D&D scene illustration, [scene description], cinematic composition, detailed fantasy environment, dramatic moment",
      "type": "scene"
    }
  ]
}
```

**Prompt rules:**
- ✅ ALWAYS include "D&D" or "Dungeons and Dragons" in prompts
- ✅ Include art style: "detailed fantasy art", "professional digital painting"
- ✅ For NPCs: include race, class, key visual features, personality
- ✅ For monsters: include size, type, environment, threatening pose
- ✅ For scenes: include environment, characters present, action/mood
- ✅ For cover: include main theme, setting, dramatic composition

**Step 3: Count and verify**

```python
# Verify full coverage
total_images = len(batch_spec["images"])
npcs_count = count_images_by_type("portrait")
monsters_count = count_images_by_type("illustration")
scenes_count = count_images_by_type("scene")
cover_count = count_images_by_type("cover")

# Report
print(f"Prepared {total_images} images: {cover_count} cover, {npcs_count} NPCs, {monsters_count} monsters, {scenes_count} scenes")
```

**RULE:** NO SKIPPING ALLOWED. Every NPC, monster, and scene MUST have an image.

### Phase B: Update Markdown References

**Step 1: List generated images**

```bash
ls {campaign_path}/assets/*.png
```

Note all generated PNG files.

**Step 2: Update README.md**

Add at the top (after the title):

```markdown
![Cover](assets/cover-art.png)
```

**Step 3: Use inline image linking (RECOMMENDED)**

When calling `generate_image`, use optional parameters to insert automatically:

```json
{
  "campaign": "campaign-name",
  "filename": "npc-gandalf",
  "prompt": "D&D wizard portrait...",
  "type": "portrait",
  "markdown_file": "npcs/npcs_and_factions.md",
  "section": "Gandalf",
  "alt": "Gandalf the Grey"
}
```

**Available parameters:**
- `markdown_file`: Path to the markdown file (e.g., `npcs/npcs_and_factions.md`)
- `section`: Section where to insert (e.g., `Gandalf`, `Act 1: The Beginning`)
- `alt`: Alt text for the image (default: filename)

**Step 4: Manual update (alternative)**

If inline linking was not used, update manually after generating:

**npcs_and_factions.md:**
```markdown
### Gandalf

[NPC description]

![Gandalf](assets/npc-gandalf.png)
```

**bestiary.md:**
```markdown
### Red Dragon

[Stat block]

![Red Dragon](assets/monster-red-dragon.png)
```

**acts/*.md:**
```markdown
[Replace `[SCENE: description]` with:]

![Scene description](assets/scene-act1-encounter.png)
```

**Step 5: Verify**

```bash
# Count image references
grep -r "!\[" {campaign_path}/*.md {campaign_path}/**/*.md | wc -l

# Verify every image in assets/ is referenced
```

## Rules

- ✅ Use kebab-case for all filenames
- ✅ Each image MUST be referenced in at least one markdown file
- ✅ Do NOT modify scene/NPC content, only ADD image references
- ✅ If an image does not exist, note it but do NOT create broken references
- ✅ Use the exact filename from assets/ (without extension) in the markdown reference

## Output to the Architect

```markdown
## Generated Art: {campaign_name}

**Status:** ✅ Complete / ❌ Failed

**Images:**
- Cover: 1
- NPC portraits: {count}
- Monster illustrations: {count}
- Scenes: {count}
- Total: {count}

**Files Updated:**
- README.md: ✅
- npcs/npcs_and_factions.md: ✅ ({count} references)
- bestiary/bestiary.md: ✅ ({count} references)
- acts/chapter_01.md: ✅ ({count} references)
- acts/chapter_02.md: ✅ ({count} references)

**batch-spec.json:** Generated in assets/
```
