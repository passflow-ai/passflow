# Passflow Design System

Light-premium aesthetic with clear backgrounds, dark mockups, and minimal violet accents. Positioning: "The workflow operating system" with open-source foundation + enterprise control.

---

## CSS Custom Properties

```css
:root {
  /* --------------------------------------------------------
     COLORS
     -------------------------------------------------------- */

  /* Backgrounds (light-premium: ~65% coverage) */
  --bg-primary:       #F6F7FB;    /* Main page background (warm gray) */
  --bg-alternate:     #F1F4FA;    /* Alternate section background */
  --bg-surface:       #FFFFFF;    /* Cards, overlays, header */
  --bg-technical:     #0F1117;    /* Dark mockups (product UI) */

  /* Text colors */
  --text-primary:     #121826;    /* Titles, main text */
  --text-secondary:   #5B6475;    /* Descriptions, labels */
  --text-muted:       #9CA3AF;    /* Metadata, timestamps */
  --text-on-dark:     #E8E8F0;    /* Text on dark backgrounds */

  /* Borders */
  --border-light:     #E5E7EB;    /* Light subtle borders */
  --border-dark:      #1E2028;    /* Dark borders in mockups */

  /* Accent: Violet (~5% coverage) */
  --violet-primary:   #7C3AED;    /* CTAs, keywords, active states */
  --violet-hover:     #6D28D9;    /* Hover on buttons */
  --violet-subtle:    rgba(124, 58, 237, 0.1);  /* Subtle backgrounds */

  /* Semantic colors */
  --green-success:    #16A34A;    /* Success, approved, checkmarks */
  --amber-pending:    #D97706;    /* Pending, awaiting, warning */
  --red-error:        #DC2626;    /* Error, problems */

  /* --------------------------------------------------------
     TYPOGRAPHY
     -------------------------------------------------------- */

  --font-display: '-apple-system', 'BlinkMacSystemFont', 'Inter', 'system-ui', sans-serif;
  --font-mono:    'JetBrains Mono', 'Fira Code', 'Courier New', monospace;
  --font-body:    '-apple-system', 'BlinkMacSystemFont', 'Inter', 'system-ui', sans-serif;

  /* Type Scale (light-premium: primarily Inter) */
  --text-hero:      3.5rem;      /* 56px - Hero headline */
  --text-h2:        1.75rem;     /* 28px - Section titles */
  --text-h3:        1rem;        /* 16px - Card titles */
  --text-body-lg:   1.05rem;     /* 17px - Hero description */
  --text-body:      0.9rem;      /* 14px - Body text */
  --text-sm:        0.85rem;     /* 14px - Secondary text */
  --text-xs:        0.75rem;     /* 12px - Labels, metadata */
  --text-xxs:       0.65rem;     /* 10px - Mockup internals */

  /* Line Heights */
  --leading-hero:     1.05;       /* Hero headline (tight) */
  --leading-heading:  1.2;        /* Section titles */
  --leading-tight:    1.4;        /* Card titles */
  --leading-body-lg:  1.7;        /* Hero description */
  --leading-body:     1.6;        /* Body text */
  --leading-sm:       1.5;        /* Small text */

  /* Letter Spacing */
  --tracking-hero:    -0.03em;    /* Hero headline */
  --tracking-heading: -0.01em;    /* Section titles */
  --tracking-normal:   0;

  /* Font Weights */
  --weight-regular:  400;
  --weight-medium:   500;
  --weight-semibold: 600;
  --weight-bold:     700;

  /* --------------------------------------------------------
     SPACING (4px base grid)
     -------------------------------------------------------- */

  --space-1:   0.25rem;   /* 4px  */
  --space-2:   0.5rem;    /* 8px  */
  --space-3:   0.75rem;   /* 12px */
  --space-4:   1rem;      /* 16px */
  --space-5:   1.25rem;   /* 20px */
  --space-6:   1.5rem;    /* 24px */
  --space-8:   2rem;      /* 32px */
  --space-10:  2.5rem;    /* 40px */
  --space-12:  3rem;      /* 48px */
  --space-16:  4rem;      /* 64px */
  --space-20:  5rem;      /* 80px */

  /* --------------------------------------------------------
     LAYOUT
     -------------------------------------------------------- */

  --max-width-content: 1300px;
  --max-width-narrow:  900px;
  --max-width-wide:    1400px;

  --section-padding-v:  4.5rem;     /* Section vertical padding */
  --section-padding-h:  2rem;       /* Section horizontal padding */

  /* --------------------------------------------------------
     BORDERS & RADII
     -------------------------------------------------------- */

  --radius-button:  8px;
  --radius-card:    10px;
  --radius-mockup:  12px;

  /* --------------------------------------------------------
     SHADOWS & EFFECTS
     -------------------------------------------------------- */

  --shadow-mockup:  0 25px 50px -12px rgba(0, 0, 0, 0.15);
  --shadow-card:    0 15px 40px -10px rgba(0, 0, 0, 0.12);

  /* --------------------------------------------------------
     TRANSITIONS
     -------------------------------------------------------- */

  --duration-fast:   150ms;
  --duration-normal: 250ms;
  --duration-slow:   400ms;
  --easing-default:  cubic-bezier(0.4, 0, 0.2, 1);
  --easing-out:      cubic-bezier(0, 0, 0.2, 1);

  /* --------------------------------------------------------
     Z-INDEX SCALE
     -------------------------------------------------------- */

  --z-base:    0;
  --z-above:   10;
  --z-dropdown: 100;
  --z-sticky:  200;
  --z-overlay: 300;
  --z-modal:   400;
  --z-toast:   500;
}
```

---

## Color Palette

### Foundation Colors

| Token | Hex | Usage |
|-------|-----|-------|
| `--bg-primary` | `#F6F7FB` | Main page background (warm gray) |
| `--bg-alternate` | `#F1F4FA` | Alternate section backgrounds |
| `--bg-surface` | `#FFFFFF` | Cards, overlays, header |
| `--bg-technical` | `#0F1117` | Dark product mockups |
| `--border-light` | `#E5E7EB` | Subtle borders on light backgrounds |
| `--border-dark` | `#1E2028` | Borders in dark mockups |

### Text Colors

| Token | Hex | Usage |
|-------|-----|-------|
| `--text-primary` | `#121826` | Titles, main body text |
| `--text-secondary` | `#5B6475` | Descriptions, labels |
| `--text-muted` | `#9CA3AF` | Metadata, timestamps |
| `--text-on-dark` | `#E8E8F0` | Text on dark backgrounds |

### Accent Colors

| Token | Hex | Usage |
|-------|-----|-------|
| `--violet-primary` | `#7C3AED` | CTAs, keywords, active states |
| `--violet-hover` | `#6D28D9` | Hover on primary buttons |
| `--violet-subtle` | `rgba(124,58,237,0.1)` | Subtle backgrounds, highlighted cards |
| `--green-success` | `#16A34A` | Success, approved, checkmarks |
| `--amber-pending` | `#D97706` | Pending, awaiting, warning states |
| `--red-error` | `#DC2626` | Error, problems, destructive |

### Color Distribution Rule

- **65%**: Light backgrounds (`--bg-primary`, `--bg-alternate`)
- **20%**: White/surface areas (`--bg-surface`)
- **10%**: Dark mockups (`--bg-technical`)
- **5%**: Violet accents (`--violet-primary`)

---

## Typography

### Font Stack

System-first with Inter as primary:
```css
font-family: -apple-system, BlinkMacSystemFont, 'Inter', system-ui, sans-serif;
```

| Role        | Family                 | Usage |
|-------------|------------------------|-------|
| Body / UI   | Inter (400, 500, 600)  | Headings, body, buttons |
| Code (opt)  | JetBrains Mono (400)   | Code blocks, terminal mockups |

### Type Scale

| Element | Size | Weight | Line Height | Letter Spacing | Usage |
|---------|------|--------|-------------|----------------|-------|
| Hero headline | 3.5rem (56px) | 600 | 1.05 | -0.03em | Hero h1 only |
| Section title (h2) | 1.75rem (28px) | 500 | 1.2 | -0.01em | Section headers |
| Card title (h3) | 1rem (16px) | 500 | 1.4 | 0 | Card/feature titles |
| Body large | 1.05rem (17px) | 400 | 1.7 | 0 | Hero description |
| Body (p) | 0.9rem (14px) | 400 | 1.6 | 0 | Main body text |
| Small | 0.85rem (14px) | 400 | 1.5 | 0 | Secondary/meta text |
| Extra small | 0.75rem (12px) | 400 | 1.4 | 0 | Labels, captions |
| Tiny (mockups) | 0.65rem (10px) | 400 | 1.3 | 0 | Mockup internals |

### Responsive Scaling

```
Mobile  (< 640px):  Hero = 2rem,    h2 = 1.25rem
Tablet  (640-1024): Hero = 2.5rem,  h2 = 1.5rem
Desktop (> 1024px): Hero = 3.5rem,  h2 = 1.75rem
```

---

## Spacing System

Base unit: **4px**. All spacing derives from the scale in `--space-*` variables.

### Spacing Values

| Token | Value | Usage |
|-------|-------|-------|
| `--space-1` | 4px | Tiny gaps, inline spacing |
| `--space-2` | 8px | Icon-text gaps |
| `--space-3` | 12px | Button padding (vertical) |
| `--space-4` | 16px | Card padding small |
| `--space-5` | 20px | Default gap between elements |
| `--space-6` | 24px | Card padding standard |
| `--space-8` | 32px | Page horizontal padding |
| `--space-10` | 40px | Trust band padding |
| `--space-12` | 48px | Section gap |
| `--space-16` | 64px | Large spacing |
| `--space-20` | 80px | Section vertical padding |

### Section Vertical Rhythm

- **Main sections**: `padding: 4.5rem 2rem` (72px vertical, 32px horizontal)
- **Hero**: `padding: 5rem 2rem 4rem` (80px top, 64px bottom)
- **Trust band**: `padding: 1.75rem 2rem`
- **Final CTA**: `padding: 5rem 2rem`

---

## Breakpoints

| Name | Min Width | Target |
|------|-----------|--------|
| Mobile | 0 | Phones (< 640px) |
| Tablet | 640px | Tablets, small laptops |
| Desktop | 1024px | Laptops and up |
| Wide | 1400px | Large monitors |

---

## Components

### Header

Sticky, translucent with blur effect:
```css
background: rgba(255, 255, 255, 0.85);
backdrop-filter: blur(12px);
border-bottom: 1px solid var(--border-light);
position: sticky;
top: 0;
z-index: var(--z-sticky);
padding: 0.875rem 2rem;
```

**Navigation:**
- Font: `--text-sm`, color: `--text-secondary`
- Items: Product, Use Cases, Architecture, Security, Pricing, Docs
- Actions: GitHub button (ghost with border) + Primary CTA

### Button Variants

**Primary (filled):**
```css
background: var(--violet-primary);
color: white;
padding: 0.75rem 1.25rem;
border-radius: var(--radius-button);
font-weight: var(--weight-medium);
```

**Secondary (ghost with border):**
```css
background: var(--bg-surface);
color: var(--text-primary);
border: 1px solid #D9DFEA;
padding: 0.75rem 1.25rem;
border-radius: var(--radius-button);
```

### Card Variants

**Light card (on light background):**
```css
background: var(--bg-surface);
border: 1px solid var(--border-light);
border-radius: var(--radius-card);
padding: 1.5rem;
```

**Feature card (subtle background):**
```css
background: #FAFBFC;
border: 1px solid var(--border-light);
border-radius: var(--radius-card);
padding: 1.25rem;
```

**Highlighted card (enterprise):**
```css
background: rgba(124, 58, 237, 0.02);
border: 2px solid var(--violet-primary);
border-radius: var(--radius-card);
padding: 1.75rem;
```

### Product Mockup (Dark)

Dark technical mockup showing workflow with approval states:

**Container:**
```css
background: var(--bg-technical);
border: 1px solid var(--border-dark);
border-radius: var(--radius-mockup);
box-shadow: var(--shadow-mockup);
overflow: hidden;
```

**Window chrome:**
```css
background: #0A0B0F;
padding: 0.6rem 1rem;
border-bottom: 1px solid var(--border-dark);
/* Traffic lights: #FF5F57, #FEBC2E, #28C840 */
```

**Internal elements (nodes/steps):**
```css
background: #161922;
border: 1px solid #252A36;
border-radius: 5px;
padding: 0.5rem 0.7rem;
font-size: var(--text-xxs);
```

**Highlighted state (pending):**
```css
background: rgba(217, 119, 6, 0.1);
border: 1px solid rgba(217, 119, 6, 0.3);
```

**Success state:**
```css
border-color: var(--green-success);
```

**Must display:**
- Version (e.g., v1.4.2)
- GitHub ref (e.g., 9f3a2b1)
- Environment (dev/prod)
- Approval state
- Run statistics

### Trust Band

Centered light band with checkmarks:
```css
background: var(--bg-surface);
border-top: 1px solid var(--border-light);
border-bottom: 1px solid var(--border-light);
padding: 1.75rem 2rem;
display: flex;
justify-content: center;
gap: 2rem;
```

Format: `✦ Open-source runtime` (centered items with checkmarks)

---

## Interactions & Hover States

### Button Hover

```css
button:hover {
  background: var(--violet-hover);
  transition: background var(--duration-fast) var(--easing-default);
}

button:focus {
  outline: 2px solid var(--violet-primary);
  outline-offset: 2px;
}
```

### Card Hover (optional lift)

```css
.card:hover {
  border-color: var(--violet-primary);
  box-shadow: 0 10px 30px rgba(124, 58, 237, 0.1);
  transition: border-color var(--duration-normal),
              box-shadow var(--duration-normal);
}
```

### Link Hover

```css
a {
  color: var(--violet-primary);
  text-decoration: underline;
}

a:hover {
  color: var(--violet-hover);
}
```

### Reduced Motion

For accessibility, respect user preferences:

```css
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

---

## Layout Patterns

### Hero Grid

Two-column layout with copy left, mockup right:
```css
.hero-grid {
  display: grid;
  grid-template-columns: 1fr 1.15fr;
  gap: 4rem;
  max-width: var(--max-width-content);
  margin: 0 auto;
}

@media (max-width: 1024px) {
  .hero-grid {
    grid-template-columns: 1fr;
  }
}
```

### Section Container

```css
.section {
  padding: var(--section-padding-v) var(--section-padding-h);
  background: var(--bg-primary);
}

.section.alternate {
  background: var(--bg-alternate);
}

.section-content {
  max-width: var(--max-width-content);
  margin: 0 auto;
}
```

### Card Grid (3 or 4 columns)

```css
.grid-3 {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
}

@media (max-width: 1024px) {
  .grid-3 {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .grid-3 {
    grid-template-columns: 1fr;
  }
}
```

### Reading Flow

- **Hero:** Left column (copy, CTAs) → Right column (mockup)
- **Problem/Value sections:** Cards in 3-column grid
- **GitHub/OpenVsEnterprise:** 2-column with alt layout
- **Final CTA:** Centered, full width

---

## Font Loading

System fonts are preferred for performance. Optional: load Inter from Google Fonts:

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&display=swap" rel="stylesheet">
```

For Next.js, use `next/font/google`:

```tsx
import { Inter } from 'next/font/google';

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-body',
  weight: ['400', '500', '600'],
  display: 'swap',
});
```

---

## Accessibility & Contrast

### WCAG AA Compliance

| Text Color | Background | Ratio | Grade |
|------------|-----------|-------|-------|
| `--text-primary` (#121826) | `--bg-primary` (#F6F7FB) | 12.4:1 | AAA ✓ |
| `--text-secondary` (#5B6475) | `--bg-primary` (#F6F7FB) | 6.8:1 | AA ✓ |
| `--text-muted` (#9CA3AF) | `--bg-primary` (#F6F7FB) | 4.2:1 | AA (for 18px+) |
| `--violet-primary` (#7C3AED) | `--bg-surface` (#FFFFFF) | 5.8:1 | AA ✓ |

### Focus States

All interactive elements require visible focus:
```css
*:focus {
  outline: 2px solid var(--violet-primary);
  outline-offset: 2px;
}
```

### Minimum Touch Target

Buttons and clickable elements: **44px × 44px** minimum

### Motion Preferences

Respect `prefers-reduced-motion` (see Reduced Motion section)

---

## Violet Usage Rules

**DO use violet for:**
- Primary CTA buttons
- Keywords in headlines (wrap with `<span>`)
- Card feature titles
- Active/selected states
- GitHub refs, version numbers

**DON'T use violet for:**
- Full section backgrounds
- Large solid bars
- Icon fills (use black/gray)
- Borders (except on highlighted cards)
