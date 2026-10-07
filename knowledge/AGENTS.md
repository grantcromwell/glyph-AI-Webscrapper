# Argos Mind Vault

This is Argos's permanent memory. Every session builds on the last.

## Glyph Architecture

Argos thinks in 21 families × 72 subfamilies. This vault stores:
- **brain/** — who Argos is, what he knows, what he's learned
- **work/** — what he's done, what he's building
- **templates/** — how to structure new memories
- **perf/** — how well he's performing
- **org/** — who and what he works with
- **thinking/** — active reasoning traces
- **reference/** — datasheets, protocol specs, design patterns

## Memory Flow

1. **Session starts** → read brain/ for identity and recent memories
2. **Work happens** → write to work/ with timestamps
3. **Lessons learned** → write to brain/Gotchas.md and brain/Patterns.md
4. **Key decisions** → write to brain/Key Decisions.md
5. **Session ends** → compact and index

## Glyph Tags

Each note should have a glyph tag in frontmatter:
- `glyph: "◐🜝⁷"` — dominant family glyph for this topic
- `families: [F00, F07, F15]` — which families are activated

## Safety

- NEVER delete files in ~/Desktop/kingvon/ or ~/Desktop/lab/ or hippocampus/gut/
- NEVER attack external systems unless commanded by Thoth
- ALWAYS document what you did
