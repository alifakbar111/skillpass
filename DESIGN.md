# SkillPass Design System

**Version:** 1.0.0  
**Last updated:** 2026-05-23  
**Stack:** Tailwind CSS v4 + DaisyUI 5  
**Storybook:** Run `bun --cwd web storybook` to open the component library at http://localhost:6006

## Philosophy

SkillPass follows a **utility-first, component-driven** approach. Every visual decision is made through DaisyUI semantic classes and Tailwind utility classes — there are zero lines of custom CSS. This keeps the design system lightweight, themeable, and easy to maintain.

**Three principles:**
1. **Semantic over absolute** — Use `bg-base-200` not `bg-gray-100`. This enables instant dark mode.
2. **Consistency over creativity** — Follow established patterns. Every page uses the same container, card, and form patterns.
3. **Minimalism over decoration** — No gratuitous gradients, shadows, or animations. Design serves the content.

---

## Typography

### Font Stack

| Role | Font | Fallback |
|---|---|---|
| Body & Headings | Outfit | `sans-serif` |
| Code & Monospace | Fira Code | `monospace` |

**Rationale:** Outfit provides a warm, geometric character that feels approachable for a job platform. Fira Code adds personality to code snippets without sacrificing readability. Both load efficiently from Google Fonts with a single `display=swap` request.

### Size Scale

| Token | Size | Weight | Used For |
|---|---|---|---|
| `text-xs` | 0.75rem | 500 | Labels, badges, error messages |
| `text-sm` | 0.875rem | 400 | Body text, descriptions, metadata |
| `text-base` | 1rem | 400 | Default paragraph text |
| `text-lg` | 1.125rem | 500 | Taglines, emphasized body |
| `text-xl` | 1.25rem | 600 | Section subheadings, navbar brand |
| `text-2xl` | 1.5rem | 700 | Page headings (h2) |
| `text-3xl` | 1.875rem | 700 | Section hero headings |
| `text-4xl` | 2.25rem | 700 | Landing primary heading |
| `text-5xl` | 3rem | 700 | Landing hero heading |

**Line height:** `1.15` for headings (`leading-tight`), `1.5` for body (`leading-normal`), `1.75` for long-form content (`leading-relaxed`).

---

## Color Palette

SkillPass uses **DaisyUI semantic color tokens exclusively**. The actual hex values depend on the active theme.

### Semantic Tokens (DaisyUI)

| Token | Usage |
|---|---|
| `bg-base-100` | Page backgrounds, navbar, card surfaces |
| `bg-base-200` | Card backgrounds, form containers, secondary surfaces |
| `bg-base-300` | Hover states on cards and clickable surfaces |
| `text-base-content` | Primary text color |
| `bg-neutral` | Avatar backgrounds |
| `text-neutral-content` | Avatar text |
| `text-error` | Error messages, destructive actions |
| `text-primary` | Primary accents, loading spinners |

### Component Color Roles

| Component | Primary | Secondary | Hover |
|---|---|---|---|
| Buttons | `btn-primary` | `btn-ghost` / `btn-outline` | DaisyUI default |
| Badges | `badge` | `badge-primary` / `badge-success` | N/A |
| Inputs | `input input-bordered` | — | `focus:input-bordered` |
| Links | `link link-primary` | — | DaisyUI default |

**Rationale:** Zero custom hex values means the palette is entirely theme-driven. Switching from `"winter"` to `"dark"` theme recolor the entire app automatically.

---

## Spacing

### Scale

SkillPass uses **Tailwind's default spacing scale** (4px base = 0.25rem).

| Token | Value | Usage |
|---|---|---|
| `1` | 0.25rem | Inline gaps between badges |
| `2` | 0.5rem | Inline action gaps, close icon spacing |
| `3` | 0.75rem | Form field gaps, badge rows |
| `4` | 1rem | **Default** — card padding, page padding, section spacing |
| `6` | 1.5rem | Page-level container spacing, outer card padding |
| `8` | 2rem | Loading/empty state centering |

### Page Layout Pattern

```tsx
<div className="max-w-2xl mx-auto p-4">
  {/* page content */}
</div>
```

| Container | Used For |
|---|---|
| `max-w-sm` | Login, Register (narrow forms) |
| `max-w-lg` | CompanyProfile, CompanyVerification (medium forms) |
| `max-w-2xl` | JobseekerProfile, JobDetail, Passport (detail views) |
| `max-w-3xl` | CompanyJobs, AdminVerifications, PublicJobs (list views) |
| `max-w-4xl` | CompanySearch (wide search results) |

---

## Border Radius

| Token | Value | Used For |
|---|---|---|
| `rounded-lg` | 0.75rem | Fieldset borders, card corners |
| `rounded-box` | var(--rounded-box) | Experience items, verification panels |
| `rounded-full` | 9999px | Avatars, theme toggle button |

**Rationale:** `rounded-box` delegates to DaisyUI's theme variable, keeping radius consistent with the active theme.

---

## Shadows

| Token | Value | Used For |
|---|---|---|
| `shadow-sm` | 0 1px 2px 0 rgb(0 0 0 / 0.05) | Navbar, dropdown menu |
| None | — | Cards intentionally have no shadow (use `bg-base-200` for separation) |

**Rationale:** Shadows are reserved for elevated elements (navbar, dropdowns). Cards use background color contrast instead of shadows, keeping the UI flat and clean.

---

## Breakpoints

| Breakpoint | Width | Behavior |
|---|---|---|
| Default | < 640px | Single column, stacked layout |
| `sm` | 640px+ | Wider containers |
| `md` | 768px+ | Multi-column grids begin |
| `lg` | 1024px+ | Sidebar layouts, wider cards |
| `xl` | 1280px+ | Maximum container width |

**Note:** The app currently uses single-column centered layout for all pages. Responsive multi-column layouts are only used on the landing page (`md:grid-cols-3`). This is acceptable for an internal tool MVP.

---

## Dark Mode

**Mechanism:** `data-theme` attribute on `<html>` toggles between `"winter"` (light) and `"dark"` (dark). Preference is persisted in `localStorage`.

**Implementation:** `web/src/components/ThemeToggle.tsx`

All DaisyUI semantic tokens adapt automatically — no `dark:` Tailwind prefixes, no custom CSS, no media queries.

---

## Accessibility

### Current State
- All form inputs use `<label className="form-control">` with `<span className="label-text">` — properly associated ✅
- Focus states provided by DaisyUI default styles ✅
- Contrast meets WCAG AA via DaisyUI theme defaults ✅
- Touch targets are 44px+ via `btn` class ✅

### Gaps & Fixes
| Issue | Fix | Priority |
|---|---|---|
| No `role="alert"` on dynamic errors | Add `role="alert"` to error `<p>` elements | Low |
| No `aria-label` on icon-only buttons | Add `aria-label="Toggle theme"` to ThemeToggle | Low |
| No skip-to-content link | Add hidden skip link in `RootLayout` | Low |

---

## Animation & Motion

### Current State
The app uses almost no animation — only `hover:bg-base-300 transition-colors` on clickable cards and DaisyUI's default button hover states.

### Proposed Additions (future)
| Element | Animation | Priority |
|---|---|---|
| Page transitions | Fade-in on route change with CSS `@keyframes fadeIn` | Low |
| Button hover | Subtle `translateY(-1px)` + shadow | Low |
| Card hover | `border-color` transition | Low |

**Principle:** Animation should be purposeful and subtle — never gratuitous. No parallax, no scroll-triggered reveals, no bouncing elements.

---

## Component Patterns

### Card
```tsx
<div className="card bg-base-200 p-4">
  {/* content */}
</div>
```

### Form Input
```tsx
<label className="form-control w-full">
  <span className="label-text">Label</span>
  <input className="input input-bordered w-full" />
  {error && <span className="text-error text-xs mt-1">{error}</span>}
</label>
```

### Button
```tsx
<button className="btn btn-primary">Primary</button>
<button className="btn btn-ghost">Ghost</button>
<button className="btn btn-outline">Outline</button>
```

### Badge
```tsx
<span className="badge badge-sm">Label</span>
<span className="badge badge-sm badge-primary">Primary</span>
<span className="badge badge-sm badge-success">Success</span>
```

### Loading State
```tsx
<div className="flex min-h-[60vh] items-center justify-center">
  <span className="loading loading-spinner loading-lg text-primary" />
</div>
```

### Empty State
```tsx
<p className="text-center opacity-50 py-8">No items found</p>
```

### Error State
```tsx
{error && <p className="text-error text-sm" role="alert">{error}</p>}
```

### Fieldset
```html
<fieldset class="fieldset">
  <legend class="fieldset-legend">Title</legend>
  <!-- form elements -->
</fieldset>
```
