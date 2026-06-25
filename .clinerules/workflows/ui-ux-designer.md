# UI/UX Designer (Cline Workflow)

Invoke via: `/ui-ux-designer <feature/page>`

Design and implement distinctive UI/UX for SkillPass web pages.

## Method

### 1. Understand Requirements
- Purpose: What problem does this interface solve? Who uses it?
- Audience: Who is the end user?
- Constraints: Technical requirements (framework, performance, accessibility).

### 2. Design Thinking
Commit to a BOLD aesthetic direction:
- **Tone**: Pick an extreme — brutally minimal, maximalist chaos, retro-futuristic, organic/natural, luxury/refined, playful/toy-like, editorial/magazine, brutalist/raw, art deco/geometric, soft/pastel, industrial/utilitarian, etc.
- **Differentiation**: What makes this UNFORGETTABLE?

### 3. Frontend Aesthetics Guidelines
- **Typography**: Choose beautiful, unique fonts. Avoid generic fonts (Arial, Inter).
- **Color & Theme**: Cohesive aesthetic. Dominant colors with sharp accents.
- **Motion**: CSS-only animations, scroll-triggering, hover states.
- **Spatial Composition**: Unexpected layouts, asymmetry, overlap, grid-breaking.
- **Backgrounds**: Gradient meshes, noise textures, geometric patterns.

NEVER use generic AI-generated aesthetics.

### 4. Implementation
- Produce a design direction first.
- After approval, implement as React components using Tailwind v4 + DaisyUI 5.
- Use `api()` wrapper for API calls, `@/*` path alias.
- Create vitest test file.
- Save design specs to `docs/specs/`.

## Return

Paths to created design specs and component files, summary of design decisions.
