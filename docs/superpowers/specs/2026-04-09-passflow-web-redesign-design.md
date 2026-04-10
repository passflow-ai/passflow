# Passflow Web Redesign - Design Specification

**Date**: 2026-04-09
**Status**: Approved
**Mockup**: `.superpowers/brainstorm/57791-1775792146/content/13-passflow-light-premium.html`

---

## 1. Executive Summary

Rediseño completo de la landing page de passflow-web con un enfoque **light-premium**: página clara con mockups de producto oscuros. El objetivo es posicionar Passflow como un **control plane para operaciones de AI workflows empresariales**, no como otro "agent builder".

### Principios de diseño

1. **Open source visible**: GitHub debe ser prominente en header, hero CTAs, y sección dedicada
2. **Light-premium**: Fondos claros (~70%), mockups oscuros (~10%), acentos mínimos (~5%)
3. **Control, not hope**: Enfatizar gobernanza, aprobaciones, trazabilidad
4. **GitHub-native**: El workflow lifecycle vive en GitHub (versiones, PRs, releases)

---

## 2. Positioning

### No somos
- Un "agent builder" como Claude Code, Cursor, o GitHub Copilot
- Una plataforma de "automation" genérica
- Un dashboard de seguridad SOC

### Somos
- **The workflow operating system** (categoría pública única)
- Open source foundation + Enterprise control layer

### Tagline principal
> "Operate AI workflows with control, not hope."

### Categoría pública (una sola)
**"Workflow Operating System"** — no "control plane", no "agent builder", no "automation platform". Esta es la categoría que usamos en hero, navegación y messaging.

### Use cases objetivo (high-stakes, no triviales)
- KYC / Onboarding
- Fraud detection
- Compliance reviews
- Risk assessment
- Critical business operations

---

## 3. Color Palette

### Foundation colors

| Token | Hex | Usage |
|-------|-----|-------|
| `--bg-primary` | `#F6F7FB` | Fondo principal (warm gray) |
| `--bg-alternate` | `#F1F4FA` | Secciones alternas |
| `--bg-surface` | `#FFFFFF` | Cards, overlays, header |
| `--bg-technical` | `#0F1117` | Mockups de producto |
| `--border-light` | `#E5E7EB` | Bordes sutiles |
| `--border-dark` | `#1E2028` | Bordes en mockups oscuros |

### Text colors

| Token | Hex | Usage |
|-------|-----|-------|
| `--text-primary` | `#121826` | Títulos, texto principal |
| `--text-secondary` | `#5B6475` | Descripciones, labels |
| `--text-muted` | `#9CA3AF` | Metadata, timestamps |
| `--text-on-dark` | `#E8E8F0` | Texto sobre fondos oscuros |

### Accent colors

| Token | Hex | Usage |
|-------|-----|-------|
| `--violet-primary` | `#7C3AED` | CTAs, keywords, estados activos |
| `--violet-hover` | `#6D28D9` | Hover en botones primarios |
| `--violet-subtle` | `rgba(124,58,237,0.1)` | Backgrounds sutiles |
| `--green-success` | `#16A34A` | Éxito, aprobado, checkmarks |
| `--amber-pending` | `#D97706` | Pendiente, awaiting, warning |
| `--red-error` | `#DC2626` | Error, problemas |

### Distribution rule
- **65%** fondos claros (`--bg-primary`, `--bg-alternate`)
- **20%** superficies blancas/gris perla
- **10%** mockups oscuros (`--bg-technical`)
- **5%** acentos (`--violet-primary`)

---

## 4. Typography

### Font stack
```css
font-family: -apple-system, BlinkMacSystemFont, 'Inter', system-ui, sans-serif;
```

### Scale

| Token | Size | Weight | Line Height | Usage |
|-------|------|--------|-------------|-------|
| `--text-hero` | 3.5rem (56px) | 600 | 1.05 | Hero headline (dominant) |
| `--text-h2` | 1.75rem (28px) | 500 | 1.2 | Section titles |
| `--text-h3` | 1rem (16px) | 500 | 1.4 | Card titles |
| `--text-body-lg` | 1.05rem (17px) | 400 | 1.7 | Hero description |
| `--text-body` | 0.9rem (14px) | 400 | 1.6 | Body text |
| `--text-sm` | 0.85rem (14px) | 400 | 1.5 | Secondary text |
| `--text-xs` | 0.75rem (12px) | 400 | 1.4 | Labels, metadata |
| `--text-xxs` | 0.65rem (10px) | 400 | 1.3 | Mockup internals |

### Letter spacing
- Hero headline: `-0.03em`
- Section titles: `-0.01em`
- Body: `0`

---

## 5. Spacing

### Base unit: 4px

| Token | Value | Usage |
|-------|-------|-------|
| `--space-1` | 0.25rem (4px) | Tiny gaps |
| `--space-2` | 0.5rem (8px) | Icon gaps |
| `--space-3` | 0.75rem (12px) | Button padding |
| `--space-4` | 1rem (16px) | Card padding small |
| `--space-5` | 1.25rem (20px) | Default gap |
| `--space-6` | 1.5rem (24px) | Card padding |
| `--space-8` | 2rem (32px) | Page horizontal padding |
| `--space-10` | 2.5rem (40px) | Trust band padding |
| `--space-12` | 3rem (48px) | Section gap |
| `--space-16` | 4rem (64px) | Large spacing |
| `--space-20` | 5rem (80px) | Section vertical padding |

### Section vertical rhythm
- Secciones principales: `padding: 4.5rem 2rem`
- Hero: `padding: 5rem 2rem 4rem`
- Trust band: `padding: 1.75rem 2rem`
- CTA final: `padding: 5rem 2rem`

---

## 6. Components

### 6.1 Header

```
┌─────────────────────────────────────────────────────────────────┐
│ [■ Passflow]    Product | Use Cases | ... | Docs  [GitHub] [CTA]│
└─────────────────────────────────────────────────────────────────┘
```

**Styles:**
- Background: `rgba(255,255,255,0.85)` with `backdrop-filter: blur(12px)`
- Border: `1px solid #E5E7EB`
- Position: `sticky top-0`
- Height: ~52px
- Logo: Square `8x8px` violet + "Passflow" weight 600

**Nav items:**
- Product, Use Cases, Architecture, Security, Pricing, Docs
- Font: 0.8rem, color `--text-secondary`

**Actions:**
- GitHub button: Ghost with border, icon + text
- Primary CTA: Filled violet

### 6.2 Buttons

**Primary (filled):**
```css
background: #7C3AED;
color: white;
padding: 0.75rem 1.25rem; /* or 1rem 2rem for large */
border-radius: 8px;
font-weight: 500;
```

**Secondary (ghost with border):**
```css
background: #FFFFFF;
color: #121826;
border: 1px solid #D9DFEA;
padding: 0.75rem 1.25rem;
border-radius: 8px;
```

**With icon:**
- Gap: 0.4rem - 0.5rem
- Icon: 15-16px

### 6.3 Cards

**Light card (sobre fondo claro):**
```css
background: #FFFFFF;
border: 1px solid #E5E7EB;
border-radius: 10px;
padding: 1.5rem;
```

**Feature card (subtle background):**
```css
background: #FAFBFC;
border: 1px solid #E5E7EB;
border-radius: 10px;
padding: 1.25rem;
```

**Highlighted card (enterprise):**
```css
background: rgba(124,58,237,0.02);
border: 2px solid #7C3AED;
border-radius: 10px;
padding: 1.75rem;
```

### 6.4 Product Mockup (Dark)

**Container:**
```css
background: #0F1117;
border: 1px solid #1E2028;
border-radius: 12px;
box-shadow: 0 25px 50px -12px rgba(0,0,0,0.15);
```

**Window chrome:**
```css
background: #0A0B0F;
padding: 0.6rem 1rem;
border-bottom: 1px solid #1E2028;
/* Traffic lights: #FF5F57, #FEBC2E, #28C840 */
```

**Internal elements:**
```css
/* Node/step */
background: #161922;
border: 1px solid #252A36;
border-radius: 5px;
padding: 0.5rem 0.7rem;

/* Highlighted state (pending) */
background: rgba(217,119,6,0.1);
border: 1px solid rgba(217,119,6,0.3);

/* Success state */
border-color: #16A34A;
```

**Debe mostrar:**
- Version (e.g., v1.4.2)
- GitHub ref (e.g., 9f3a2b1)
- Environment (dev/prod)
- Approval state
- Run stats

### 6.5 Trust Band

```css
background: #FFFFFF;
border-top: 1px solid #E5E7EB;
border-bottom: 1px solid #E5E7EB;
padding: 1.75rem 2rem;
/* Centered items with checkmarks */
```

Items format: `✦ Open-source runtime`

### 6.6 Checkmarks & Icons

**Success checkmark:**
```html
<span style="color: #16A34A;">✓</span>
```

**Status indicators in mockups:**
- Trigger: `●` violet
- Decision/Policy: `◆` amber
- Pending: `⏸` amber
- Success: `✓` green

---

## 7. Page Sections

### 7.1 Hero

**Layout:** Grid 2 columns (1fr 1.15fr)
- Left: Copy on light background
- Right: Dark product mockup with shadow

**Left column:**
1. Subtitle: "Open-source workflow runtime · Enterprise control plane"
2. Headline: "Operate AI workflows with **control, not hope.**"
3. Description: ~2 líneas sobre versioning, governance, execution
4. CTAs: [Book demo] (primary) + [Start on GitHub] (secondary with icon)
5. Trust points: 3 checkmarks horizontales

**Right column:**
- Mockup mostrando workflow visual con approval gate
- Datos de contexto: version, GitHub ref, env, risk level
- Estado actual: "Awaiting approval"
- Stats bar: runs, success rate, pending, agents

### 7.2 Trust Band

4 items centrados:
- ✦ Open-source runtime
- ✦ GitHub-native lifecycle
- ✦ Policy-enforced execution
- ✦ Full auditability

### 7.3 Problem Section

**Title:** "Most AI workflows break in operations, not in demos."

**3 cards (grid):**
1. No governance (red title)
2. No visibility (red title)
3. No safe path (red title)

### 7.4 Value Proposition

**Title:** "An open foundation with enterprise control."

**4 cards (grid):**
1. Build openly
2. Version in GitHub
3. Run with guardrails
4. Audit everything

Card titles en violet.

### 7.5 GitHub Section

**Layout:** Grid 2 columns
- Left: Copy con bullet points
- Right: Dark GitHub mockup

**Mockup contents:**
- File: `kyc-onboarding.yaml`
- PR card: `#142 Add approval gate` - ✓ Approved
- Tags: version + environment

### 7.6 Open Source vs Enterprise

**Layout:** Grid 2 columns

**Left card (Open Source):**
- GitHub icon + title
- List: Runtime, SDK/CLI, Workflow spec, Templates, Local tooling
- Border normal

**Right card (Enterprise):**
- Title en violet
- List: Control plane, RBAC, Audit, Isolated execution, SSO
- Border violet 2px

### 7.7 Final CTA

**Title:** "Build openly. Operate responsibly."

**CTAs:** Same as hero (Book demo + Start on GitHub)

**Trust points:** 3 items con ✦

### 7.8 Footer (Enterprise-grade)

**Layout:** 4 columnas + copyright row

**Columnas:**
1. **Product**: Features, Use Cases, Architecture, Roadmap
2. **Developers**: Documentation, GitHub, API Reference, Templates
3. **Company**: About, Blog, Careers, Contact
4. **Legal**: Privacy, Terms, Security, Compliance

**Copyright row:**
- Left: Logo pequeño + "© 2026 Passflow"
- Center: "Open source under Apache 2.0"
- Right: Social icons (GitHub, Twitter/X, LinkedIn)

**Styles:**
- Background: `#FFFFFF`
- Border top: `1px solid #E5E7EB`
- Padding: `4rem 2rem 2rem`
- Column titles: `--text-primary`, weight 500
- Links: `--text-secondary`, hover `--violet-primary`

---

## 8. Implementation Rules for React/Tailwind

### Tailwind config tokens

```js
// tailwind.config.js
module.exports = {
  theme: {
    extend: {
      colors: {
        'pf-bg': '#F6F7FB',
        'pf-bg-alt': '#F1F4FA',
        'pf-surface': '#FFFFFF',
        'pf-technical': '#0F1117',
        'pf-border': '#E5E7EB',
        'pf-border-dark': '#1E2028',
        'pf-text': '#121826',
        'pf-text-secondary': '#5B6475',
        'pf-text-muted': '#9CA3AF',
        'pf-text-light': '#E8E8F0',
        'pf-violet': '#7C3AED',
        'pf-violet-hover': '#6D28D9',
        'pf-green': '#16A34A',
        'pf-amber': '#D97706',
        'pf-red': '#DC2626',
      },
      fontFamily: {
        sans: ['-apple-system', 'BlinkMacSystemFont', 'Inter', 'system-ui', 'sans-serif'],
      },
      fontSize: {
        'hero': ['3.5rem', { lineHeight: '1.05', letterSpacing: '-0.03em', fontWeight: '600' }],
        'h2': ['1.75rem', { lineHeight: '1.2', fontWeight: '500' }],
      },
      borderRadius: {
        'card': '10px',
        'mockup': '12px',
      },
      boxShadow: {
        'mockup': '0 25px 50px -12px rgba(0,0,0,0.15)',
        'card': '0 15px 40px -10px rgba(0,0,0,0.12)',
      },
    },
  },
}
```

### Component patterns

**Section wrapper:**
```tsx
<section className="py-[4.5rem] px-8 bg-pf-bg">
  <div className="max-w-[900px] mx-auto">
    {/* content */}
  </div>
</section>
```

**Hero grid:**
```tsx
<div className="grid grid-cols-[1fr_1.15fr] gap-16 max-w-[1300px] mx-auto">
```

**Button primary:**
```tsx
<button className="bg-pf-violet hover:bg-pf-violet-hover text-white
  px-5 py-3 rounded-lg font-medium text-sm">
```

**Card:**
```tsx
<div className="bg-white border border-pf-border rounded-card p-6">
```

**Product mockup:**
```tsx
<div className="bg-pf-technical border border-pf-border-dark
  rounded-mockup overflow-hidden shadow-mockup">
```

### Violet usage rules

**DO:**
- Primary CTA buttons
- Keywords en headlines (wrap con `<span className="text-pf-violet">`)
- Card feature titles
- Active states
- Links en mockups (GitHub refs, PRs)

**DON'T:**
- Full section backgrounds
- Large solid bars
- Icons (usar negro/gris)
- Borders excepto en cards destacadas

### Accessibility

- Contrast ratio text-primary sobre bg: >7:1
- Contrast ratio text-secondary sobre bg: >4.5:1
- Focus states: `ring-2 ring-pf-violet ring-offset-2`
- Button min touch target: 44x44px

---

## 9. Responsive Considerations

### Breakpoints

| Breakpoint | Width | Adjustments |
|------------|-------|-------------|
| Mobile | <640px | Single column, stack hero, smaller typography |
| Tablet | 640-1024px | 2-col grids → 1-col, reduce padding |
| Desktop | >1024px | Full layout as designed |

### Hero mobile
- Stack columns (copy arriba, mockup abajo)
- Headline: 2rem
- Mockup: Full width, max-height con overflow hidden

### Cards mobile
- Full width
- Stack vertical
- Padding reducido

---

## 10. Assets Required

### Icons
- GitHub icon (SVG, 16px base)
- Checkmark (✓ o SVG)
- Star symbols (✦)

### Fonts
- Inter (fallback to system)

### Images
- Logo: Small square violet + wordmark
- (Optional) Background gradients CSS-only

---

## 11. Implementation Priorities

### P0 — Foundation + Hero (must ship first)

| File | Changes | Acceptance Criteria |
|------|---------|---------------------|
| `tailwind.config.ts` | Add all custom tokens | All pf-* classes working |
| `src/app/globals.css` | Update CSS variables | Variables match spec |
| `src/components/Header.tsx` | New light translucent header | Sticky, blur, GitHub visible |
| `src/components/Hero.tsx` | Split layout with mockup | 56px headline, CTAs, mockup right |
| `src/components/ProductMockup.tsx` | New component | Shows workflow, version, approval state |
| `src/components/TrustBand.tsx` | 4 trust points | Centered, light bg |

### P1 — Core Sections

| File | Changes | Acceptance Criteria |
|------|---------|---------------------|
| `src/components/sections/Problem.tsx` | 3 problem cards | Red titles, white cards |
| `src/components/sections/ValueProp.tsx` | 4 feature cards | Violet titles |
| `src/components/sections/GitHub.tsx` | Split with mockup | GitHub mockup dark |
| `src/components/sections/OpenVsEnterprise.tsx` | 2-column comparison | Enterprise card highlighted |
| `src/components/sections/FinalCTA.tsx` | Repeat CTAs | Same as hero |
| `src/components/Footer.tsx` | 4-column enterprise footer | All links, social icons |

### P2 — Polish

| File | Changes | Acceptance Criteria |
|------|---------|---------------------|
| `DESIGN-SYSTEM.md` | Replace with new system | Documents new tokens |
| Responsive breakpoints | Mobile/tablet adjustments | Hero stacks, cards full-width |
| Microinteractions | Hover states, transitions | Smooth, consistent |
| `src/app/[locale]/page.tsx` | Wire all sections | Clean composition |

### Execution order
1. **P0 first, ship as unblocking PR**
2. P1 can parallelize (sections are independent)
3. P2 after P0+P1 merged

---

## Appendix: Mockup Reference

El mockup HTML aprobado está en:
```
.superpowers/brainstorm/57791-1775792146/content/13-passflow-light-premium.html
```

Abrir en navegador para referencia visual durante implementación.
