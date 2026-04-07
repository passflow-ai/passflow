# Passflow Design System

Dark, industrial-technical aesthetic. Terminal interfaces, not SaaS dashboards.

---

## CSS Custom Properties

```css
:root {
  /* --------------------------------------------------------
     COLORS
     -------------------------------------------------------- */

  /* Backgrounds */
  --bg-primary:       #0A0A0F;
  --bg-elevated:      #111118;
  --bg-surface:       #16161F;
  --bg-overlay:       rgba(10, 10, 15, 0.85);

  /* Accent: Terminal Green */
  --accent-primary:       #00FF87;
  --accent-primary-rgb:   0, 255, 135;
  --accent-primary-muted: #00CC6A;
  --accent-primary-dim:   #00994F;
  --accent-primary-ghost: rgba(0, 255, 135, 0.08);
  --accent-primary-glow:  rgba(0, 255, 135, 0.25);

  /* Accent: Electric Blue */
  --accent-secondary:       #0066FF;
  --accent-secondary-rgb:   0, 102, 255;
  --accent-secondary-muted: #0052CC;
  --accent-secondary-dim:   #003D99;
  --accent-secondary-ghost: rgba(0, 102, 255, 0.08);
  --accent-secondary-glow:  rgba(0, 102, 255, 0.25);

  /* Text */
  --text-primary:   #E8E8F0;
  --text-secondary: #6B7280;
  --text-tertiary:  #4B5563;
  --text-disabled:  #374151;
  --text-inverse:   #0A0A0F;
  --text-accent:    #00FF87;

  /* Borders */
  --border-primary:   #1E1E2E;
  --border-secondary: #2A2A3A;
  --border-hover:     #3A3A4A;
  --border-focus:     #00FF87;
  --border-error:     #FF3B5C;

  /* Semantic */
  --color-success:      #00FF87;
  --color-success-dim:  #00CC6A;
  --color-error:        #FF3B5C;
  --color-error-dim:    #CC2F4A;
  --color-warning:      #FFB800;
  --color-warning-dim:  #CC9300;
  --color-info:         #0066FF;
  --color-info-dim:     #0052CC;

  /* Gradients */
  --gradient-hero:      linear-gradient(180deg, #0A0A0F 0%, #111118 50%, #0A0A0F 100%);
  --gradient-radial:    radial-gradient(ellipse at center, rgba(0, 255, 135, 0.06) 0%, transparent 70%);
  --gradient-card:      linear-gradient(135deg, #16161F 0%, #111118 100%);
  --gradient-accent:    linear-gradient(135deg, #00FF87 0%, #0066FF 100%);
  --gradient-text:      linear-gradient(135deg, #00FF87 0%, #0066FF 100%);
  --gradient-scanline:  repeating-linear-gradient(
                          0deg,
                          transparent,
                          transparent 2px,
                          rgba(0, 255, 135, 0.03) 2px,
                          rgba(0, 255, 135, 0.03) 4px
                        );
  --gradient-noise:     url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='0.04'/%3E%3C/svg%3E");

  /* --------------------------------------------------------
     TYPOGRAPHY
     -------------------------------------------------------- */

  --font-display: 'Playfair Display', Georgia, serif;
  --font-mono:    'JetBrains Mono', 'Fira Code', 'Courier New', monospace;
  --font-body:    'Inter', system-ui, -apple-system, sans-serif;

  /* Type Scale (1.25 ratio -- Major Third) */
  --text-xs:   0.75rem;    /* 12px */
  --text-sm:   0.875rem;   /* 14px */
  --text-base: 1rem;       /* 16px */
  --text-lg:   1.125rem;   /* 18px */
  --text-xl:   1.25rem;    /* 20px */
  --text-2xl:  1.5rem;     /* 24px */
  --text-3xl:  1.875rem;   /* 30px */
  --text-4xl:  2.25rem;    /* 36px */
  --text-5xl:  3rem;       /* 48px */
  --text-6xl:  3.75rem;    /* 60px */
  --text-7xl:  4.5rem;     /* 72px */

  /* Line Heights */
  --leading-none:    1;
  --leading-tight:   1.15;
  --leading-snug:    1.3;
  --leading-normal:  1.5;
  --leading-relaxed: 1.65;
  --leading-loose:   2;

  /* Letter Spacing */
  --tracking-tighter: -0.04em;
  --tracking-tight:   -0.02em;
  --tracking-normal:   0;
  --tracking-wide:     0.02em;
  --tracking-wider:    0.05em;
  --tracking-widest:   0.1em;

  /* Font Weights */
  --weight-regular:  400;
  --weight-medium:   500;
  --weight-semibold: 600;
  --weight-bold:     700;
  --weight-black:    900;

  /* --------------------------------------------------------
     SPACING (8px base grid)
     -------------------------------------------------------- */

  --space-0:   0;
  --space-px:  1px;
  --space-0-5: 0.125rem;  /* 2px  */
  --space-1:   0.25rem;   /* 4px  */
  --space-1-5: 0.375rem;  /* 6px  */
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
  --space-24:  6rem;      /* 96px */
  --space-32:  8rem;      /* 128px */

  /* --------------------------------------------------------
     LAYOUT
     -------------------------------------------------------- */

  --max-width-content: 1200px;
  --max-width-narrow:  800px;
  --max-width-wide:    1400px;

  --section-padding-y:        var(--space-20);   /* 80px mobile  */
  --section-padding-y-desktop: var(--space-24);   /* 96px desktop */
  --container-padding-x:       var(--space-4);    /* 16px mobile  */
  --container-padding-x-tablet: var(--space-6);   /* 24px tablet  */
  --container-padding-x-desktop: var(--space-8);  /* 32px desktop */

  /* --------------------------------------------------------
     BORDERS & RADII
     -------------------------------------------------------- */

  --radius-sm:   4px;
  --radius-md:   8px;
  --radius-lg:   12px;
  --radius-xl:   16px;
  --radius-full: 9999px;

  /* --------------------------------------------------------
     SHADOWS & EFFECTS
     -------------------------------------------------------- */

  --shadow-sm:    0 1px 2px rgba(0, 0, 0, 0.4);
  --shadow-md:    0 4px 12px rgba(0, 0, 0, 0.5);
  --shadow-lg:    0 8px 24px rgba(0, 0, 0, 0.6);
  --shadow-glow-green: 0 0 20px rgba(0, 255, 135, 0.15),
                       0 0 60px rgba(0, 255, 135, 0.05);
  --shadow-glow-blue:  0 0 20px rgba(0, 102, 255, 0.15),
                       0 0 60px rgba(0, 102, 255, 0.05);

  /* --------------------------------------------------------
     TRANSITIONS
     -------------------------------------------------------- */

  --duration-fast:   150ms;
  --duration-normal: 250ms;
  --duration-slow:   400ms;
  --duration-slower:  600ms;
  --easing-default:  cubic-bezier(0.4, 0, 0.2, 1);
  --easing-in:       cubic-bezier(0.4, 0, 1, 1);
  --easing-out:      cubic-bezier(0, 0, 0.2, 1);
  --easing-spring:   cubic-bezier(0.34, 1.56, 0.64, 1);

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

| Token                  | Hex / Value             | Usage                                  |
|------------------------|-------------------------|----------------------------------------|
| `--bg-primary`         | `#0A0A0F`               | Page background, body                  |
| `--bg-elevated`        | `#111118`               | Cards, raised surfaces                 |
| `--bg-surface`         | `#16161F`               | Input backgrounds, code blocks         |
| `--bg-overlay`         | `rgba(10,10,15,0.85)`   | Modal overlays, backdrop               |
| `--accent-primary`     | `#00FF87`               | CTAs, highlights, active states        |
| `--accent-secondary`   | `#0066FF`               | Links, secondary actions, info         |
| `--text-primary`       | `#E8E8F0`               | Headings, body copy                    |
| `--text-secondary`     | `#6B7280`               | Descriptions, metadata                 |
| `--text-tertiary`      | `#4B5563`               | Placeholders, captions                 |
| `--border-primary`     | `#1E1E2E`               | Card borders, dividers                 |
| `--border-focus`       | `#00FF87`               | Focus rings, active borders            |

### Semantic Colors

- **Success:** `#00FF87` -- shares the terminal green accent.
- **Error:** `#FF3B5C` -- high contrast red for destructive states.
- **Warning:** `#FFB800` -- amber for caution indicators.
- **Info:** `#0066FF` -- shares the electric blue accent.

---

## Typography

### Font Stack

| Role          | Family                                       | Load Source    |
|---------------|----------------------------------------------|----------------|
| Headlines     | `Playfair Display` (700, 900)                | Google Fonts   |
| Code / UI     | `JetBrains Mono` (400, 500, 700)             | Google Fonts   |
| Body          | `Inter` (400, 500, 600, 700)                 | Google Fonts   |

### Type Scale

| Element   | Font            | Size         | Weight | Line Height | Letter Spacing | Usage                          |
|-----------|-----------------|--------------|--------|-------------|----------------|--------------------------------|
| `h1`      | Playfair Display | `--text-7xl` | 900    | 1.0         | -0.04em        | Hero headline only             |
| `h2`      | Playfair Display | `--text-5xl` | 700    | 1.15        | -0.02em        | Section titles                 |
| `h3`      | Inter            | `--text-3xl` | 700    | 1.3         | -0.02em        | Subsection titles              |
| `h4`      | Inter            | `--text-2xl` | 600    | 1.3         | 0              | Card titles                    |
| `h5`      | JetBrains Mono   | `--text-lg`  | 500    | 1.5         | 0.02em         | Labels, overlines              |
| `h6`      | JetBrains Mono   | `--text-sm`  | 500    | 1.5         | 0.1em          | Uppercase labels, tags         |
| `body`    | Inter            | `--text-lg`  | 400    | 1.65        | 0              | Paragraphs                     |
| `body-sm` | Inter            | `--text-base`| 400    | 1.5         | 0              | Secondary body copy            |
| `small`   | Inter            | `--text-sm`  | 400    | 1.5         | 0              | Captions, footnotes            |
| `code`    | JetBrains Mono   | `--text-sm`  | 400    | 1.65        | 0              | Inline code, terminal output   |

### Responsive Scaling

```
Mobile  (< 768px):  h1 = 2.25rem,  h2 = 1.875rem
Tablet  (>= 768px): h1 = 3rem,     h2 = 2.25rem
Desktop (>= 1024px): h1 = 4.5rem,  h2 = 3rem
```

---

## Spacing System

Base unit: **8px**. All spacing derives from the scale in `--space-*` variables.

### Section Layout

| Context          | Mobile    | Tablet    | Desktop   |
|------------------|-----------|-----------|-----------|
| Section padding Y | 80px      | 80px      | 96px      |
| Container padding X | 16px   | 24px      | 32px      |
| Card padding     | 24px      | 32px      | 32px      |
| Card gap         | 16px      | 24px      | 24px      |
| Component gap (vertical) | 12px | 16px   | 16px      |

---

## Breakpoints

| Name    | Min Width | Target             |
|---------|-----------|--------------------|
| Mobile  | 0         | Phones             |
| Tablet  | 768px     | Tablets, small laptops |
| Desktop | 1024px    | Laptops and up     |
| Wide    | 1400px    | Large monitors (optional clamp) |

---

## Components

### Card

```
Background: var(--bg-elevated)
Border: 1px solid var(--border-primary)
Border radius: var(--radius-lg)            -- 12px
Padding: var(--space-6) to var(--space-8)  -- 24-32px
Transition: border-color 250ms, box-shadow 250ms

Hover:
  border-color: var(--accent-primary-muted)
  box-shadow: var(--shadow-glow-green)
```

### Buttons

Three variants. All use `JetBrains Mono` at `--text-sm`, `--weight-semibold`, uppercase, `--tracking-wider`.

**Primary**
```
Background: var(--accent-primary)
Color: var(--text-inverse)                  -- #0A0A0F
Padding: 14px 32px
Border radius: var(--radius-md)
Position: relative; overflow: hidden

Pseudo ::after (scanline effect):
  content: ""
  position: absolute; inset: 0
  background: var(--gradient-scanline)
  pointer-events: none

Hover:
  background: var(--accent-primary-muted)
  box-shadow: var(--shadow-glow-green)
  transform: translateY(-1px)

Active:
  transform: translateY(0)
```

**Secondary**
```
Background: transparent
Color: var(--accent-primary)
Border: 1px solid var(--accent-primary)
Padding: 14px 32px

Hover:
  background: var(--accent-primary-ghost)
  box-shadow: var(--shadow-glow-green)
```

**Ghost**
```
Background: transparent
Color: var(--text-secondary)
Padding: 14px 32px
Border: 1px solid var(--border-primary)

Hover:
  color: var(--text-primary)
  border-color: var(--border-hover)
  background: var(--bg-surface)
```

### Input

```
Background: transparent
Border: none
Border-bottom: 1px solid var(--border-primary)
Color: var(--text-primary)
Font: JetBrains Mono, --text-base
Padding: var(--space-3) 0
Transition: border-color 250ms

Focus:
  border-bottom-color: var(--accent-primary)
  outline: none

Placeholder:
  color: var(--text-tertiary)
```

### Tabs (Code Examples)

```
Container:
  border: 1px solid var(--border-primary)
  border-radius: var(--radius-lg)
  overflow: hidden

Tab bar:
  background: var(--bg-surface)
  display: flex
  border-bottom: 1px solid var(--border-primary)

Tab item:
  font: JetBrains Mono, --text-sm, --weight-medium
  color: var(--text-secondary)
  padding: var(--space-3) var(--space-4)
  border-bottom: 2px solid transparent

Tab item (active):
  color: var(--accent-primary)
  border-bottom-color: var(--accent-primary)

Tab content:
  background: var(--bg-elevated)
  padding: var(--space-6)
  font: JetBrains Mono, --text-sm
  color: var(--text-primary)
  overflow-x: auto
```

### Comparison Table

```
Table:
  width: 100%
  border-collapse: collapse

th:
  font: JetBrains Mono, --text-xs, --weight-semibold
  text-transform: uppercase
  letter-spacing: var(--tracking-widest)
  color: var(--text-secondary)
  padding: var(--space-3) var(--space-4)
  border-bottom: 1px solid var(--border-primary)
  text-align: left

td:
  font: Inter, --text-sm
  color: var(--text-primary)
  padding: var(--space-4)
  border-bottom: 1px solid var(--border-primary)

tr:hover td:
  background: var(--accent-primary-ghost)

Checkmark icon color: var(--accent-primary)
Cross icon color: var(--text-tertiary)
```

### Pricing Card

```
Base: same as Card component

Recommended variant (add):
  border-color: var(--accent-primary)
  box-shadow: var(--shadow-glow-green)
  position: relative

Recommended badge:
  position: absolute
  top: -12px; left: 50%; transform: translateX(-50%)
  background: var(--accent-primary)
  color: var(--text-inverse)
  font: JetBrains Mono, --text-xs, --weight-semibold
  text-transform: uppercase
  letter-spacing: var(--tracking-widest)
  padding: var(--space-1) var(--space-4)
  border-radius: var(--radius-full)

Price number:
  font: Playfair Display, --text-5xl, --weight-bold
  color: var(--text-primary)

Price period:
  font: Inter, --text-sm, --weight-regular
  color: var(--text-secondary)
```

### Terminal Window

```
Container:
  border: 1px solid var(--border-primary)
  border-radius: var(--radius-lg)
  overflow: hidden

Title bar:
  background: var(--bg-surface)
  padding: var(--space-3) var(--space-4)
  display: flex; align-items: center; gap: var(--space-2)
  border-bottom: 1px solid var(--border-primary)

Traffic lights (3 circles):
  width: 12px; height: 12px; border-radius: 50%
  Colors: #FF5F57, #FEBC2E, #28C840

Title text:
  font: JetBrains Mono, --text-xs
  color: var(--text-tertiary)
  margin-left: var(--space-3)

Body:
  background: var(--bg-primary)
  padding: var(--space-6)
  font: JetBrains Mono, --text-sm
  color: var(--accent-primary)
  line-height: var(--leading-relaxed)
  min-height: 300px

Prompt prefix:
  color: var(--text-secondary)
  content: "$ " or ">"

Cursor (blinking block):
  display: inline-block
  width: 8px; height: 18px
  background: var(--accent-primary)
  animation: blink 1s step-end infinite

@keyframes blink {
  50% { opacity: 0; }
}
```

### Badge / Pill

```
display: inline-flex; align-items: center
font: JetBrains Mono, --text-xs, --weight-medium
text-transform: uppercase
letter-spacing: var(--tracking-wider)
padding: var(--space-1) var(--space-3)
border-radius: var(--radius-full)

Variant "green":
  background: var(--accent-primary-ghost)
  color: var(--accent-primary)
  border: 1px solid rgba(0, 255, 135, 0.2)

Variant "blue":
  background: var(--accent-secondary-ghost)
  color: var(--accent-secondary)
  border: 1px solid rgba(0, 102, 255, 0.2)

Variant "neutral":
  background: var(--bg-surface)
  color: var(--text-secondary)
  border: 1px solid var(--border-primary)
```

### Sticky Header

```
position: fixed; top: 0; left: 0; right: 0
z-index: var(--z-sticky)
background: var(--bg-overlay)
backdrop-filter: blur(12px) saturate(180%)
-webkit-backdrop-filter: blur(12px) saturate(180%)
border-bottom: 1px solid var(--border-primary)
padding: var(--space-3) 0
transition: background 300ms, border-color 300ms

Inner container:
  max-width: var(--max-width-content)
  margin: 0 auto
  padding: 0 var(--container-padding-x)
  display: flex; align-items: center; justify-content: space-between

Nav links:
  font: JetBrains Mono, --text-sm, --weight-medium
  color: var(--text-secondary)
  transition: color var(--duration-fast)

Nav links hover:
  color: var(--accent-primary)

Logo text:
  font: JetBrains Mono, --text-lg, --weight-bold
  color: var(--text-primary)
```

---

## Effects & Animations

### Background Grain Noise

Apply to `body::before` or a fixed overlay element:

```css
.grain-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  pointer-events: none;
  opacity: 0.4;
  background-image: var(--gradient-noise);
  background-repeat: repeat;
}
```

### Radial Glow

Used behind hero section and accent areas:

```css
.radial-glow {
  position: absolute;
  width: 600px;
  height: 600px;
  border-radius: 50%;
  background: radial-gradient(
    circle,
    rgba(0, 255, 135, 0.08) 0%,
    transparent 70%
  );
  filter: blur(80px);
  pointer-events: none;
}
```

### Animated Dot Grid (Hero)

```css
.dot-grid {
  position: absolute;
  inset: 0;
  background-image: radial-gradient(
    circle,
    var(--border-primary) 1px,
    transparent 1px
  );
  background-size: 32px 32px;
  opacity: 0.4;
  mask-image: radial-gradient(ellipse at center, black 30%, transparent 70%);
}
```

### Typewriter Effect (Terminal)

```css
.typewriter {
  overflow: hidden;
  white-space: nowrap;
  border-right: 2px solid var(--accent-primary);
  animation:
    typing 3s steps(40) 1s forwards,
    cursor-blink 0.75s step-end infinite;
  width: 0;
}

@keyframes typing {
  to { width: 100%; }
}

@keyframes cursor-blink {
  50% { border-color: transparent; }
}
```

### Scroll-Triggered Fade-In

Use Intersection Observer in JS. CSS classes:

```css
.reveal {
  opacity: 0;
  transform: translateY(24px);
  transition: opacity var(--duration-slow) var(--easing-out),
              transform var(--duration-slow) var(--easing-out);
}

.reveal.visible {
  opacity: 1;
  transform: translateY(0);
}
```

### Stagger Animation (Architecture Diagram, Feature Cards)

```css
.stagger-item {
  opacity: 0;
  transform: translateY(16px);
  transition: opacity var(--duration-slow) var(--easing-out),
              transform var(--duration-slow) var(--easing-out);
}

.stagger-item.visible {
  opacity: 1;
  transform: translateY(0);
}

/* Apply via JS: each child gets delay = index * 100ms */
.stagger-item:nth-child(1) { transition-delay: 0ms; }
.stagger-item:nth-child(2) { transition-delay: 100ms; }
.stagger-item:nth-child(3) { transition-delay: 200ms; }
.stagger-item:nth-child(4) { transition-delay: 300ms; }
.stagger-item:nth-child(5) { transition-delay: 400ms; }
.stagger-item:nth-child(6) { transition-delay: 500ms; }
```

### Counter Animation (Metrics)

Implement via JS with `requestAnimationFrame`. Duration: 2 seconds. Easing: ease-out. Trigger on scroll into view. Format numbers with locale separators.

### Scanline Button Effect

```css
.btn-scanline::after {
  content: "";
  position: absolute;
  inset: 0;
  background: var(--gradient-scanline);
  pointer-events: none;
  opacity: 0.5;
}
```

### Reduced Motion

```css
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
    scroll-behavior: auto !important;
  }

  .typewriter {
    width: 100%;
    border-right: none;
  }

  .reveal,
  .stagger-item {
    opacity: 1;
    transform: none;
  }
}
```

---

## Layout

### Container

```css
.container {
  width: 100%;
  max-width: var(--max-width-content);
  margin: 0 auto;
  padding-left: var(--container-padding-x);
  padding-right: var(--container-padding-x);
}

@media (min-width: 768px) {
  .container {
    padding-left: var(--container-padding-x-tablet);
    padding-right: var(--container-padding-x-tablet);
  }
}

@media (min-width: 1024px) {
  .container {
    padding-left: var(--container-padding-x-desktop);
    padding-right: var(--container-padding-x-desktop);
  }
}
```

### Section

```css
.section {
  padding-top: var(--section-padding-y);
  padding-bottom: var(--section-padding-y);
}

@media (min-width: 1024px) {
  .section {
    padding-top: var(--section-padding-y-desktop);
    padding-bottom: var(--section-padding-y-desktop);
  }
}
```

### Grid System

```css
.grid-2 { display: grid; gap: var(--space-6); }
.grid-3 { display: grid; gap: var(--space-6); }
.grid-4 { display: grid; gap: var(--space-6); }

@media (min-width: 768px) {
  .grid-2 { grid-template-columns: repeat(2, 1fr); }
  .grid-3 { grid-template-columns: repeat(2, 1fr); }
  .grid-4 { grid-template-columns: repeat(2, 1fr); }
}

@media (min-width: 1024px) {
  .grid-3 { grid-template-columns: repeat(3, 1fr); }
  .grid-4 { grid-template-columns: repeat(4, 1fr); }
}
```

### Reading Flow

- **Hero:** Z-pattern -- logo top-left, CTA top-right, headline center, action bottom.
- **Feature sections:** F-pattern -- heading left-aligned, content scanning left to right.
- **Alternating rows:** Image-left/text-right, then swap, to maintain engagement.
- **Final CTA:** Centered, single focal point.

---

## Font Loading

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&family=Playfair+Display:wght@700;900&family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
```

For Next.js, use `next/font/google`:

```tsx
import { JetBrains_Mono, Playfair_Display, Inter } from 'next/font/google';

const jetbrains = JetBrains_Mono({
  subsets: ['latin'],
  variable: '--font-mono',
  weight: ['400', '500', '700'],
});

const playfair = Playfair_Display({
  subsets: ['latin'],
  variable: '--font-display',
  weight: ['700', '900'],
});

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-body',
  weight: ['400', '500', '600', '700'],
});
```

---

## Accessibility Notes

- All text meets WCAG AA contrast on `--bg-primary`. The `--accent-primary` (#00FF87) on `--bg-primary` (#0A0A0F) yields ~11:1 contrast ratio.
- `--text-secondary` (#6B7280) on `--bg-primary` yields ~5.5:1 -- passes AA for normal text.
- `--text-tertiary` (#4B5563) on `--bg-primary` yields ~3.8:1 -- use only for large text (18px+) or decorative elements.
- Focus states use `--border-focus` (#00FF87) with a 2px offset ring.
- All interactive elements must have visible focus indicators.
- `prefers-reduced-motion` disables all animations.
- `prefers-color-scheme` is not applicable (always dark).
