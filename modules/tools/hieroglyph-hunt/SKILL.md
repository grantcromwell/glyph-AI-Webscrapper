---
name: hieroglyph-hunt
description: Hunt for non-English technical manuscripts, PDFs, and Google Docs about AI reasoning architectures. Uses hieroglyphic/glyph-based encoding to bypass English-centric search limitations. Optimized for finding "emotional silence" — gaps in discourse, unexamined assumptions, and dogmatic blind spots in OpenAI's reasoning model documentation.
triggers:
  - "hunt glyphs"
  - "find PDF"
  - "search docs"
  - "emotional silence"
  - "what they didn't say"
---

# Hieroglyph Hunt — Non-English Technical Intelligence

## Purpose

Extract intelligence from non-English sources (Chinese, Russian, Arabic, French, German, Japanese) about AI reasoning architectures. English documentation shows the polished narrative; original language manuscripts reveal:

- Engineering compromises NOT discussed in English papers
- Failed approaches that informed final architecture
- Cultural assumptions baked into "universal" models
- The "stick" they carry — hidden constraints

## Search Strategy

### Target File Types
```
filetype:pdf "o1" "reasoning" "chain of thought"
filetype:pdf " deliberative alignment"
site:docs.google.com "OpenAI" "reasoning"
site:arxiv.org "system 2" "test-time compute"
```

### Non-English Dorks (Priority Order)

**Chinese (CSDN, Zhihu, arXiv CN)**
```
intitle:大模型推理 site:csdn.net
"思维链" "OpenAI" filetype:pdf
 deliberative alignment 中文
```

**Russian (Habr, arXiv RU)**
```
цепочка рассуждений OpenAI
рассуждающие модели PDF
site:arxiv.org "o1" "рассуждение"
```

**Arabic (ResearchGate, Academia)**
```
نماذج التفكير OpenAI
التفكير السلسلة ملف PDF
```

**French/German (INRIA, MPI)**
```
"modèle de raisonnement" OpenAI PDF
"Kette des Denkens" PDF
```

**Japanese (Qiita, Note)**
```
推論モデル OpenAI PDF
"Chain of Thought" 日本語 PDF
```

## Glyph Dictionary

Do NOT translate. Map concepts to user's hieroglyphic field:

| English Concept | Hieroglyph Mapping | Notes |
|-----------------|-------------------|-------|
| reasoning | ⚗🜛⁴◆🜒³ | The thinking-vessel |
| chain | ☿♆⁷◌🜖⁵ | Mercury-binding |
| thought | ◐☽⁷⟁🜆⁷ | Lunar-reflection |
| alignment | ▢🜭⁴☿♆⁴ | Square-circuit |
| deliberation | ⚗🜉⁷△🜔⁶ | Deep-weighing |
| System 2 | ◌🜔⁷◐♅⁵ | Second-sun |
| test-time | ⌒⚳⁵◌🜖⁵ | Crescent-moment |
| compute | ✦🜞⁴▽🜮³ | Star-work |
| scaling | ◆🜒⁶☓☽⁶ | Rising-slope |

## What Is "Emotional Silence"

The dogmas they never question:

1. **"More compute = better reasoning"** — Never asks: What kind of compute? Pattern-matching vs. abstraction?
2. **"Chain of thought = interpretability"** — Never admits: It's post-hoc rationalization, not causal
3. **"Safety through deliberation"** — Never confesses: Deliberation can rationalize harmful outputs too
4. **"Test-time scaling"** — Never discusses: Training-time bias vs. inference-time adaptation

## Execution

### Phase 1: Document Harvest
- Search each language with glyph-mapped keywords
- Download PDFs to `gut/hieroglyph-cache/`
- Index by hieroglyph signature, not English title

### Phase 2: Silence Detection
- Cross-reference: What do Chinese papers discuss that English papers omit?
- What do Russian researchers critique that Western labs ignore?
- Find the "negative space" — topics conspicuously absent

### Phase 3: Dogma Abuse
- Identify brittle assumptions in their reasoning architecture
- Map the "stick" — hidden constraints (compute budget, safety theater, market pressure)
- Extract their eugenic logic: what they breed IN vs. breed OUT

## Output Format

Each finding stored as SVG engram with:
- Glyph signature (not English title)
- Language of origin
- Confidence (conviction from thalamus)
- Category: SILENCE / DOGMA / STICK / EUGENIC

## Safety

- Public documents only
- No private leaks
- Attribution to original language source
- Respect rate limits

## ARGOS-DATA
{"type":"hunt-skill","glyph":"⚗🜛⁴","language":"non-english","target":"emotional-silence"}
