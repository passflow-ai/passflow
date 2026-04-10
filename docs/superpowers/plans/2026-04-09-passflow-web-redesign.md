# Passflow Web Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign passflow-web from dark terminal aesthetic to light-premium with dark product mockups, repositioning as "Workflow Operating System"

**Architecture:** Replace CSS variables in globals.css with new light-premium palette. Rewrite all section components with new visual language. Add ProductMockup component for dark product screenshots. Consolidate 10 sections to 8.

**Tech Stack:** Next.js 16, Tailwind CSS 4, next-intl, CSS variables

**Spec:** `docs/superpowers/specs/2026-04-09-passflow-web-redesign-design.md`

---

## File Structure

### Modified Files

| File | Responsibility |
|------|----------------|
| `src/app/globals.css` | CSS variables (colors, typography, spacing, components) |
| `src/components/Header.tsx` | Light translucent header with GitHub button |
| `src/components/Hero.tsx` | Split layout: copy left, ProductMockup right |
| `src/components/Footer.tsx` | 4-column enterprise footer |
| `src/components/FinalCTA.tsx` | Repeat CTAs with trust points |
| `src/app/[locale]/page.tsx` | Wire new sections, remove deprecated ones |

### New Files

| File | Responsibility |
|------|----------------|
| `src/components/ProductMockup.tsx` | Dark mockup showing workflow with approval gate |
| `src/components/TrustBand.tsx` | 4 trust points (OSS, GitHub-native, Policy, Audit) |
| `src/components/Problem.tsx` | 3 problem cards (replaces Urgency) |
| `src/components/ValueProp.tsx` | 4 feature cards (replaces HowItWorks) |
| `src/components/GitHub.tsx` | GitHub integration section with mockup |
| `src/components/OpenVsEnterprise.tsx` | OSS vs Enterprise comparison |

### Deprecated Files (delete after migration)

| File | Replaced By |
|------|-------------|
| `src/components/Urgency.tsx` | Problem.tsx |
| `src/components/HowItWorks.tsx` | ValueProp.tsx |
| `src/components/Differentiation.tsx` | OpenVsEnterprise.tsx |
| `src/components/Trust.tsx` | TrustBand.tsx |
| `src/components/OpenSource.tsx` | GitHub.tsx + OpenVsEnterprise.tsx |
| `src/components/UseCases.tsx` | (removed - use cases integrated into Problem/ValueProp) |
| `src/components/Pricing.tsx` | (removed for V1 - will be re-added later) |

---

## P0 — Foundation + Hero

### Task 1: Update CSS Variables

**Files:**
- Modify: `src/app/globals.css:1-200`

- [ ] **Step 1: Read current globals.css structure**

Verify the file starts with `@import "tailwindcss"` and has `:root` block.

- [ ] **Step 2: Replace color variables in :root**

Replace lines 3-67 (COLORS section) with:

```css
:root {
  /* --------------------------------------------------------
     COLORS - Light Premium System
     -------------------------------------------------------- */

  /* Backgrounds */
  --bg-primary: #F6F7FB;
  --bg-alternate: #F1F4FA;
  --bg-surface: #FFFFFF;
  --bg-technical: #0F1117;
  --bg-overlay: rgba(255, 255, 255, 0.85);

  /* Borders */
  --border-light: #E5E7EB;
  --border-dark: #1E2028;
  --border-hover: #D1D5DB;
  --border-focus: #7C3AED;

  /* Text */
  --text-primary: #121826;
  --text-secondary: #5B6475;
  --text-muted: #9CA3AF;
  --text-on-dark: #E8E8F0;
  --text-inverse: #FFFFFF;

  /* Accent: Violet */
  --accent-primary: #7C3AED;
  --accent-primary-hover: #6D28D9;
  --accent-primary-subtle: rgba(124, 58, 237, 0.1);
  --accent-primary-ghost: rgba(124, 58, 237, 0.02);

  /* Semantic */
  --color-success: #16A34A;
  --color-warning: #D97706;
  --color-error: #DC2626;

  /* Dark mockup internals */
  --mockup-bg: #0F1117;
  --mockup-surface: #161922;
  --mockup-border: #252A36;
  --mockup-chrome: #0A0B0F;
```

- [ ] **Step 3: Update typography variables**

Replace typography section with:

```css
  /* --------------------------------------------------------
     TYPOGRAPHY
     -------------------------------------------------------- */

  --font-sans: -apple-system, BlinkMacSystemFont, 'Inter', system-ui, sans-serif;
  --font-mono: 'JetBrains Mono', 'Fira Code', monospace;

  /* Type Scale */
  --text-hero: 3.5rem;
  --text-h2: 1.75rem;
  --text-h3: 1rem;
  --text-body-lg: 1.05rem;
  --text-body: 0.9rem;
  --text-sm: 0.85rem;
  --text-xs: 0.75rem;
  --text-xxs: 0.65rem;

  /* Line Heights */
  --leading-hero: 1.05;
  --leading-tight: 1.2;
  --leading-normal: 1.6;
  --leading-relaxed: 1.7;

  /* Letter Spacing */
  --tracking-tight: -0.03em;
  --tracking-normal: 0;

  /* Weights */
  --weight-regular: 400;
  --weight-medium: 500;
  --weight-semibold: 600;
```

- [ ] **Step 4: Update shadows and effects**

```css
  /* --------------------------------------------------------
     SHADOWS & EFFECTS
     -------------------------------------------------------- */

  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 12px rgba(0, 0, 0, 0.08);
  --shadow-mockup: 0 25px 50px -12px rgba(0, 0, 0, 0.15);
  --shadow-card: 0 15px 40px -10px rgba(0, 0, 0, 0.12);

  /* --------------------------------------------------------
     BORDERS & RADII
     -------------------------------------------------------- */

  --radius-sm: 6px;
  --radius-md: 8px;
  --radius-lg: 10px;
  --radius-xl: 12px;
}
```

- [ ] **Step 5: Update @theme inline block**

Replace the `@theme inline` block (~line 195):

```css
@theme inline {
  --color-background: var(--bg-primary);
  --color-foreground: var(--text-primary);
  --color-primary: var(--accent-primary);
  --font-sans: var(--font-sans);
  --font-mono: var(--font-mono);
}
```

- [ ] **Step 6: Update body styles**

```css
html {
  scroll-behavior: smooth;
}

body {
  background: var(--bg-primary);
  color: var(--text-primary);
  font-family: var(--font-sans);
}
```

- [ ] **Step 7: Update button classes**

Replace `.btn-primary`, `.btn-secondary`, `.btn-ghost` (~lines 256-333):

```css
/* --------------------------------------------------------
   BUTTONS
   -------------------------------------------------------- */

.btn-primary {
  background: var(--accent-primary);
  color: var(--text-inverse);
  padding: 0.75rem 1.25rem;
  border-radius: var(--radius-md);
  font-family: var(--font-sans);
  font-size: var(--text-sm);
  font-weight: var(--weight-medium);
  transition: background 150ms ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.btn-primary:hover {
  background: var(--accent-primary-hover);
}

.btn-secondary {
  background: var(--bg-surface);
  color: var(--text-primary);
  padding: 0.75rem 1.25rem;
  border-radius: var(--radius-md);
  border: 1px solid var(--border-light);
  font-family: var(--font-sans);
  font-size: var(--text-sm);
  font-weight: var(--weight-medium);
  transition: border-color 150ms ease, background 150ms ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.btn-secondary:hover {
  border-color: var(--border-hover);
  background: var(--bg-alternate);
}
```

- [ ] **Step 8: Update card classes**

Replace `.card` styles (~lines 339-362):

```css
/* --------------------------------------------------------
   CARDS
   -------------------------------------------------------- */

.card {
  background: var(--bg-surface);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
}

.card-feature {
  background: #FAFBFC;
  border: 1px solid var(--border-light);
  border-radius: var(--radius-lg);
  padding: 1.25rem;
}

.card-highlight {
  background: var(--accent-primary-ghost);
  border: 2px solid var(--accent-primary);
  border-radius: var(--radius-lg);
  padding: 1.75rem;
}
```

- [ ] **Step 9: Add new utility classes**

Add after cards section:

```css
/* --------------------------------------------------------
   SECTIONS
   -------------------------------------------------------- */

.section {
  padding: 4.5rem 2rem;
}

.section-alt {
  background: var(--bg-alternate);
}

/* --------------------------------------------------------
   TYPOGRAPHY CLASSES
   -------------------------------------------------------- */

.text-hero {
  font-size: var(--text-hero);
  font-weight: var(--weight-semibold);
  line-height: var(--leading-hero);
  letter-spacing: var(--tracking-tight);
  color: var(--text-primary);
}

.text-h2 {
  font-size: var(--text-h2);
  font-weight: var(--weight-medium);
  line-height: var(--leading-tight);
  color: var(--text-primary);
}

.text-accent {
  color: var(--accent-primary);
}

/* --------------------------------------------------------
   MOCKUP STYLES
   -------------------------------------------------------- */

.mockup-container {
  background: var(--mockup-bg);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-xl);
  overflow: hidden;
  box-shadow: var(--shadow-mockup);
}

.mockup-chrome {
  background: var(--mockup-chrome);
  padding: 0.6rem 1rem;
  border-bottom: 1px solid var(--border-dark);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.mockup-traffic-lights {
  display: flex;
  gap: 0.4rem;
}

.mockup-traffic-light {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.mockup-node {
  background: var(--mockup-surface);
  border: 1px solid var(--mockup-border);
  border-radius: 5px;
  padding: 0.5rem 0.7rem;
  color: var(--text-on-dark);
  font-size: var(--text-xxs);
}

.mockup-node-pending {
  background: rgba(217, 119, 6, 0.1);
  border-color: rgba(217, 119, 6, 0.3);
}

.mockup-node-success {
  border-color: var(--color-success);
}
```

- [ ] **Step 10: Remove dark-theme specific styles**

Delete or comment out these classes that are no longer needed:
- `.grain-overlay`
- `.radial-glow`
- `.dot-grid`
- `.gradient-*` references
- `.badge-green`, `.badge-blue` (replace with new badge styles)

- [ ] **Step 11: Run dev server to verify CSS compiles**

Run: `npm run dev`
Expected: No CSS compilation errors, page loads with light background

- [ ] **Step 12: Commit CSS foundation**

```bash
git add src/app/globals.css
git commit -m "style: replace dark terminal with light-premium CSS system

- New color palette: warm gray bg, violet accent
- Typography: 56px hero, Inter font
- Components: buttons, cards, mockups
- Remove dark-theme effects (grain, radial glow)
"
```

---

### Task 2: Create ProductMockup Component

**Files:**
- Create: `src/components/ProductMockup.tsx`

- [ ] **Step 1: Create ProductMockup.tsx**

```tsx
const ProductMockup = () => {
  return (
    <div className="mockup-container">
      {/* Chrome bar */}
      <div className="mockup-chrome">
        <div className="mockup-traffic-lights">
          <span className="mockup-traffic-light" style={{ background: '#FF5F57' }} />
          <span className="mockup-traffic-light" style={{ background: '#FEBC2E' }} />
          <span className="mockup-traffic-light" style={{ background: '#28C840' }} />
        </div>
        <span style={{ color: '#4B5563', fontSize: '0.65rem' }}>
          passflow — kyc-onboarding
        </span>
        <div style={{ display: 'flex', gap: '0.75rem', fontSize: '0.65rem' }}>
          <span style={{ color: '#4B5563' }}>Draft</span>
          <span style={{ color: 'var(--accent-primary)', fontWeight: 500 }}>Reviewed</span>
          <span style={{ color: '#4B5563' }}>Released</span>
        </div>
      </div>

      {/* Main content */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 240px', fontSize: '0.72rem', color: 'var(--text-on-dark)' }}>
        {/* Canvas */}
        <div style={{ padding: '1rem', borderRight: '1px solid var(--border-dark)' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.4rem' }}>
            <div className="mockup-node" style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
              <span style={{ color: 'var(--accent-primary)' }}>●</span> Trigger: New request
            </div>
            <div style={{ textAlign: 'center', color: '#3A3F4A', fontSize: '0.6rem' }}>↓</div>
            <div className="mockup-node">Validate inputs</div>
            <div style={{ textAlign: 'center', color: '#3A3F4A', fontSize: '0.6rem' }}>↓</div>
            <div className="mockup-node" style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
              <span style={{ color: 'var(--color-warning)' }}>◆</span> Risk policy
            </div>
            <div style={{ textAlign: 'center', color: '#3A3F4A', fontSize: '0.6rem' }}>↓</div>
            <div className="mockup-node mockup-node-pending" style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
              <span style={{ color: 'var(--color-warning)' }}>⏸</span> <strong>Approval gate</strong>
            </div>
            <div style={{ textAlign: 'center', color: '#3A3F4A', fontSize: '0.6rem' }}>↓</div>
            <div className="mockup-node mockup-node-success" style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
              <span style={{ color: 'var(--color-success)' }}>✓</span> Execute & log
            </div>
          </div>
        </div>

        {/* Side panel */}
        <div style={{ padding: '0.85rem', background: 'var(--mockup-chrome)', fontSize: '0.68rem' }}>
          <div style={{ marginBottom: '0.85rem' }}>
            <div style={{ color: '#4B5563', marginBottom: '0.3rem', fontSize: '0.6rem', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              Run details
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.3rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: '#6B7280' }}>Version</span>
                <span>v1.4.2</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: '#6B7280' }}>GitHub</span>
                <span style={{ color: 'var(--accent-primary)' }}>9f3a2b1</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: '#6B7280' }}>Env</span>
                <span style={{ color: 'var(--color-success)' }}>prod</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: '#6B7280' }}>Risk</span>
                <span style={{ color: 'var(--color-warning)' }}>High</span>
              </div>
            </div>
          </div>
          <div style={{ background: 'rgba(217,119,6,0.1)', color: 'var(--color-warning)', padding: '0.35rem 0.5rem', borderRadius: '4px', textAlign: 'center', fontSize: '0.65rem' }}>
            ⏸ Awaiting approval
          </div>
        </div>
      </div>

      {/* Stats bar */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', padding: '0.6rem 1rem', borderTop: '1px solid var(--border-dark)', background: 'var(--mockup-chrome)', fontSize: '0.65rem', textAlign: 'center', color: 'var(--text-on-dark)' }}>
        <div><strong>847</strong> <span style={{ color: '#4B5563' }}>runs</span></div>
        <div><strong style={{ color: 'var(--color-success)' }}>99.2%</strong> <span style={{ color: '#4B5563' }}>success</span></div>
        <div><strong style={{ color: 'var(--color-warning)' }}>3</strong> <span style={{ color: '#4B5563' }}>pending</span></div>
        <div><strong>12</strong> <span style={{ color: '#4B5563' }}>agents</span></div>
      </div>
    </div>
  );
};

export default ProductMockup;
```

- [ ] **Step 2: Verify component renders**

Add temporarily to page.tsx and check in browser.

- [ ] **Step 3: Commit ProductMockup**

```bash
git add src/components/ProductMockup.tsx
git commit -m "feat: add ProductMockup component

Dark workflow visualization with:
- Traffic light chrome
- Workflow nodes with approval gate
- Run details panel
- Stats bar
"
```

---

### Task 3: Create TrustBand Component

**Files:**
- Create: `src/components/TrustBand.tsx`

- [ ] **Step 1: Create TrustBand.tsx**

```tsx
const TrustBand = () => {
  const items = [
    "Open-source runtime",
    "GitHub-native lifecycle",
    "Policy-enforced execution",
    "Full auditability",
  ];

  return (
    <div style={{
      padding: '1.75rem 2rem',
      background: 'var(--bg-surface)',
      borderTop: '1px solid var(--border-light)',
      borderBottom: '1px solid var(--border-light)',
    }}>
      <div style={{
        maxWidth: '900px',
        margin: '0 auto',
        display: 'flex',
        justifyContent: 'center',
        gap: '2.5rem',
        color: 'var(--text-secondary)',
        fontSize: '0.8rem',
        flexWrap: 'wrap',
      }}>
        {items.map((item) => (
          <span key={item}>✦ {item}</span>
        ))}
      </div>
    </div>
  );
};

export default TrustBand;
```

- [ ] **Step 2: Commit TrustBand**

```bash
git add src/components/TrustBand.tsx
git commit -m "feat: add TrustBand component

4 trust points: OSS runtime, GitHub-native, Policy, Audit
"
```

---

### Task 4: Rewrite Header Component

**Files:**
- Modify: `src/components/Header.tsx`

- [ ] **Step 1: Update Header with light translucent style**

Replace entire file:

```tsx
"use client";

import { useState } from "react";
import { useTranslations, useLocale } from "next-intl";
import { Link, usePathname, useRouter } from "@/i18n/routing";

const GitHubIcon = () => (
  <svg style={{ width: 14, height: 14 }} viewBox="0 0 24 24" fill="currentColor">
    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
  </svg>
);

const Header = () => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const pathname = usePathname();
  const router = useRouter();
  const locale = useLocale();
  const t = useTranslations("Header");
  const isHomePage = pathname === "/";

  const scrollToSection = (id: string) => {
    if (isHomePage) {
      const element = document.getElementById(id);
      if (element) {
        element.scrollIntoView({ behavior: "smooth" });
      }
    } else {
      window.location.href = `/#${id}`;
    }
    setIsMenuOpen(false);
  };

  const switchLocale = (newLocale: "es" | "en") => {
    router.replace(pathname, { locale: newLocale });
  };

  const navLinks = [
    { label: "Product", id: "how-it-works" },
    { label: "Use Cases", id: "use-cases" },
    { label: "Architecture", id: "architecture" },
    { label: "Security", id: "security" },
    { label: "Docs", href: "/docs" },
  ];

  return (
    <header style={{
      position: 'sticky',
      top: 0,
      zIndex: 100,
      background: 'rgba(255,255,255,0.85)',
      backdropFilter: 'blur(12px)',
      borderBottom: '1px solid var(--border-light)',
    }}>
      <div style={{
        maxWidth: '1200px',
        margin: '0 auto',
        padding: '0 2rem',
      }}>
        <div style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          height: '52px',
        }}>
          {/* Logo */}
          <Link href="/" style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
            textDecoration: 'none',
          }}>
            <div style={{
              width: '8px',
              height: '8px',
              background: 'var(--accent-primary)',
              borderRadius: '2px',
            }} />
            <span style={{
              fontWeight: 600,
              fontSize: '0.95rem',
              color: 'var(--text-primary)',
            }}>
              Passflow
            </span>
          </Link>

          {/* Desktop nav */}
          <nav style={{
            display: 'flex',
            gap: '1.75rem',
          }} className="hidden md:flex">
            {navLinks.map((link) => (
              link.href ? (
                <Link
                  key={link.href}
                  href={link.href}
                  style={{
                    fontSize: '0.8rem',
                    color: 'var(--text-secondary)',
                    textDecoration: 'none',
                  }}
                >
                  {link.label}
                </Link>
              ) : (
                <button
                  key={link.id}
                  onClick={() => scrollToSection(link.id!)}
                  style={{
                    fontSize: '0.8rem',
                    color: 'var(--text-secondary)',
                    background: 'none',
                    border: 'none',
                    cursor: 'pointer',
                  }}
                >
                  {link.label}
                </button>
              )
            ))}
          </nav>

          {/* Actions */}
          <div style={{ display: 'flex', gap: '0.6rem', alignItems: 'center' }} className="hidden md:flex">
            <a
              href="https://github.com/passflow-ai/passflow"
              target="_blank"
              rel="noopener noreferrer"
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '0.35rem',
                padding: '0.5rem 0.75rem',
                border: '1px solid var(--border-light)',
                borderRadius: '6px',
                background: 'var(--bg-surface)',
                color: 'var(--text-primary)',
                fontSize: '0.8rem',
                textDecoration: 'none',
              }}
            >
              <GitHubIcon />
              GitHub
            </a>
            <a href="/demo" className="btn-primary" style={{ padding: '0.5rem 1rem', fontSize: '0.8rem' }}>
              Book demo
            </a>
          </div>

          {/* Language switcher */}
          <div className="hidden md:flex" style={{ marginLeft: '0.75rem', gap: '0.25rem', alignItems: 'center' }}>
            <button
              onClick={() => switchLocale("es")}
              style={{
                fontSize: '0.7rem',
                padding: '0.25rem 0.5rem',
                borderRadius: '4px',
                border: 'none',
                background: locale === "es" ? 'var(--accent-primary-subtle)' : 'transparent',
                color: locale === "es" ? 'var(--accent-primary)' : 'var(--text-secondary)',
                cursor: 'pointer',
              }}
            >
              ES
            </button>
            <span style={{ color: 'var(--text-muted)' }}>/</span>
            <button
              onClick={() => switchLocale("en")}
              style={{
                fontSize: '0.7rem',
                padding: '0.25rem 0.5rem',
                borderRadius: '4px',
                border: 'none',
                background: locale === "en" ? 'var(--accent-primary-subtle)' : 'transparent',
                color: locale === "en" ? 'var(--accent-primary)' : 'var(--text-secondary)',
                cursor: 'pointer',
              }}
            >
              EN
            </button>
          </div>

          {/* Mobile menu button */}
          <button
            onClick={() => setIsMenuOpen(!isMenuOpen)}
            className="md:hidden"
            style={{
              padding: '0.5rem',
              background: 'none',
              border: 'none',
              color: 'var(--text-primary)',
              cursor: 'pointer',
            }}
            aria-label={isMenuOpen ? "Close menu" : "Open menu"}
          >
            <svg width="24" height="24" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              {isMenuOpen ? (
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              ) : (
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              )}
            </svg>
          </button>
        </div>

        {/* Mobile menu */}
        {isMenuOpen && (
          <div className="md:hidden" style={{
            padding: '1rem 0',
            borderTop: '1px solid var(--border-light)',
          }}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
              {navLinks.map((link) => (
                link.href ? (
                  <Link
                    key={link.href}
                    href={link.href}
                    style={{ fontSize: '0.9rem', color: 'var(--text-secondary)', textDecoration: 'none' }}
                    onClick={() => setIsMenuOpen(false)}
                  >
                    {link.label}
                  </Link>
                ) : (
                  <button
                    key={link.id}
                    onClick={() => scrollToSection(link.id!)}
                    style={{ fontSize: '0.9rem', color: 'var(--text-secondary)', background: 'none', border: 'none', textAlign: 'left', cursor: 'pointer' }}
                  >
                    {link.label}
                  </button>
                )
              ))}
              <a
                href="https://github.com/passflow-ai/passflow"
                target="_blank"
                rel="noopener noreferrer"
                style={{ fontSize: '0.9rem', color: 'var(--text-secondary)', textDecoration: 'none' }}
              >
                GitHub
              </a>
              <a href="/demo" className="btn-primary" style={{ marginTop: '0.5rem', textAlign: 'center' }}>
                Book demo
              </a>
            </div>
          </div>
        )}
      </div>
    </header>
  );
};

export default Header;
```

- [ ] **Step 2: Verify header renders correctly**

Run: `npm run dev`
Expected: Light translucent header with GitHub button visible

- [ ] **Step 3: Commit Header**

```bash
git add src/components/Header.tsx
git commit -m "style: rewrite Header with light translucent design

- White translucent bg with blur
- GitHub button visible in header
- Violet accent for CTA
- Clean nav links
"
```

---

### Task 5: Rewrite Hero Component

**Files:**
- Modify: `src/components/Hero.tsx`

- [ ] **Step 1: Replace Hero with split layout**

```tsx
"use client";

import ProductMockup from "./ProductMockup";

const GitHubIcon = () => (
  <svg style={{ width: 15, height: 15 }} viewBox="0 0 24 24" fill="currentColor">
    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
  </svg>
);

const Hero = () => {
  const trustPoints = [
    "GitHub-native versioning",
    "Approval workflows",
    "Full audit trail",
  ];

  return (
    <section style={{
      background: 'linear-gradient(180deg, #FFFFFF 0%, var(--bg-primary) 100%)',
      paddingTop: '5rem',
      paddingBottom: '4rem',
    }}>
      <div style={{
        display: 'grid',
        gridTemplateColumns: '1fr 1.15fr',
        gap: '4rem',
        maxWidth: '1300px',
        margin: '0 auto',
        padding: '0 2rem',
        alignItems: 'center',
      }}>
        {/* Left: Copy */}
        <div>
          <div style={{
            color: 'var(--text-secondary)',
            fontSize: '0.8rem',
            marginBottom: '1.25rem',
            letterSpacing: '0.01em',
          }}>
            Open-source workflow runtime · Enterprise control plane
          </div>

          <h1 style={{
            fontSize: '3.5rem',
            fontWeight: 600,
            lineHeight: 1.05,
            marginBottom: '1.25rem',
            letterSpacing: '-0.03em',
            color: 'var(--text-primary)',
          }}>
            Operate AI workflows with{" "}
            <span style={{ color: 'var(--accent-primary)' }}>control, not hope.</span>
          </h1>

          <p style={{
            color: 'var(--text-secondary)',
            fontSize: '1.05rem',
            marginBottom: '2rem',
            lineHeight: 1.7,
            maxWidth: '480px',
          }}>
            Design, version, deploy, and govern AI workflows with GitHub-based delivery, approval gates, audit trails, and isolated execution.
          </p>

          <div style={{ display: 'flex', gap: '0.6rem', marginBottom: '1.5rem' }}>
            <a href="/demo" className="btn-primary" style={{ padding: '0.8rem 1.4rem', fontSize: '0.9rem' }}>
              Book demo
            </a>
            <a
              href="https://github.com/passflow-ai/passflow"
              target="_blank"
              rel="noopener noreferrer"
              className="btn-secondary"
              style={{ padding: '0.8rem 1.4rem', fontSize: '0.9rem' }}
            >
              <GitHubIcon />
              Start on GitHub
            </a>
          </div>

          <div style={{
            display: 'flex',
            gap: '1.25rem',
            color: 'var(--text-secondary)',
            fontSize: '0.75rem',
          }}>
            {trustPoints.map((point) => (
              <span key={point} style={{ display: 'flex', alignItems: 'center', gap: '0.3rem' }}>
                <span style={{ color: 'var(--color-success)' }}>✓</span> {point}
              </span>
            ))}
          </div>
        </div>

        {/* Right: Product mockup */}
        <ProductMockup />
      </div>
    </section>
  );
};

export default Hero;
```

- [ ] **Step 2: Verify Hero renders with mockup**

Run: `npm run dev`
Expected: Split layout with copy left, dark mockup right

- [ ] **Step 3: Commit Hero**

```bash
git add src/components/Hero.tsx
git commit -m "feat: rewrite Hero with split layout and ProductMockup

- 56px headline with violet accent
- Trust points with checkmarks
- GitHub CTA prominent
- ProductMockup showing workflow
"
```

---

### Task 6: Wire P0 Components in Page

**Files:**
- Modify: `src/app/[locale]/page.tsx`

- [ ] **Step 1: Update page.tsx with P0 components**

```tsx
import Header from "@/components/Header";
import Hero from "@/components/Hero";
import TrustBand from "@/components/TrustBand";
import Footer from "@/components/Footer";

export default function Home() {
  return (
    <>
      <Header />
      <main>
        <Hero />
        <TrustBand />
        {/* P1 sections will be added here */}
      </main>
      <Footer />
    </>
  );
}
```

- [ ] **Step 2: Verify page renders**

Run: `npm run dev`
Expected: Header, Hero with mockup, TrustBand, Footer

- [ ] **Step 3: Commit P0 page structure**

```bash
git add src/app/[locale]/page.tsx
git commit -m "feat: wire P0 components (Header, Hero, TrustBand)

Foundation complete. Ready for P1 sections.
"
```

---

## P1 — Core Sections

### Task 7: Create Problem Section

**Files:**
- Create: `src/components/Problem.tsx`

- [ ] **Step 1: Create Problem.tsx**

```tsx
const Problem = () => {
  const problems = [
    {
      title: "No governance",
      description: "Workflow logic changes without review, ownership, or release discipline.",
    },
    {
      title: "No visibility",
      description: "Teams can't inspect what ran, who approved it, or where it failed.",
    },
    {
      title: "No safe path",
      description: "A workflow that works in a demo becomes a liability in production.",
    },
  ];

  return (
    <section style={{ padding: '4.5rem 2rem', background: 'var(--bg-primary)' }}>
      <div style={{ maxWidth: '900px', margin: '0 auto' }}>
        <h2 style={{
          fontSize: '1.75rem',
          fontWeight: 500,
          marginBottom: '0.75rem',
          textAlign: 'center',
          color: 'var(--text-primary)',
        }}>
          Most AI workflows break in operations, not in demos.
        </h2>
        <p style={{
          color: 'var(--text-secondary)',
          textAlign: 'center',
          marginBottom: '3rem',
          maxWidth: '600px',
          marginLeft: 'auto',
          marginRight: 'auto',
        }}>
          Teams wire together agents quickly. What fails is everything around execution: safe releases, approvals, traceability, and accountability.
        </p>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(3, 1fr)',
          gap: '1.5rem',
        }}>
          {problems.map((problem) => (
            <div key={problem.title} className="card">
              <h3 style={{
                fontSize: '0.95rem',
                fontWeight: 500,
                marginBottom: '0.5rem',
                color: 'var(--color-error)',
              }}>
                {problem.title}
              </h3>
              <p style={{
                color: 'var(--text-secondary)',
                fontSize: '0.85rem',
              }}>
                {problem.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default Problem;
```

- [ ] **Step 2: Commit Problem**

```bash
git add src/components/Problem.tsx
git commit -m "feat: add Problem section with 3 pain point cards"
```

---

### Task 8: Create ValueProp Section

**Files:**
- Create: `src/components/ValueProp.tsx`

- [ ] **Step 1: Create ValueProp.tsx**

```tsx
const ValueProp = () => {
  const features = [
    {
      title: "Build openly",
      description: "Create workflows with reusable components and developer-friendly patterns.",
    },
    {
      title: "Version in GitHub",
      description: "Track changes, review logic, manage releases, and promote safely.",
    },
    {
      title: "Run with guardrails",
      description: "Apply policies, approvals, and environment rules before execution.",
    },
    {
      title: "Audit everything",
      description: "Capture logs, approvals, actions, and outcomes in a traceable record.",
    },
  ];

  return (
    <section style={{
      padding: '4.5rem 2rem',
      background: 'var(--bg-surface)',
      borderTop: '1px solid var(--border-light)',
    }}>
      <div style={{ maxWidth: '950px', margin: '0 auto' }}>
        <h2 style={{
          fontSize: '1.75rem',
          fontWeight: 500,
          marginBottom: '0.5rem',
          textAlign: 'center',
          color: 'var(--text-primary)',
        }}>
          An open foundation with enterprise control.
        </h2>
        <p style={{
          color: 'var(--text-secondary)',
          textAlign: 'center',
          marginBottom: '3rem',
        }}>
          Open development practices + operational safeguards for real business workflows.
        </p>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(4, 1fr)',
          gap: '1.25rem',
        }}>
          {features.map((feature) => (
            <div key={feature.title} className="card-feature">
              <div style={{
                color: 'var(--accent-primary)',
                fontSize: '0.85rem',
                fontWeight: 500,
                marginBottom: '0.4rem',
              }}>
                {feature.title}
              </div>
              <p style={{
                color: 'var(--text-secondary)',
                fontSize: '0.8rem',
              }}>
                {feature.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default ValueProp;
```

- [ ] **Step 2: Commit ValueProp**

```bash
git add src/components/ValueProp.tsx
git commit -m "feat: add ValueProp section with 4 feature cards"
```

---

### Task 9: Create GitHub Section

**Files:**
- Create: `src/components/GitHub.tsx`

- [ ] **Step 1: Create GitHub.tsx with mockup**

```tsx
const GitHubIcon = () => (
  <svg style={{ width: 16, height: 16, color: 'var(--text-on-dark)' }} viewBox="0 0 24 24" fill="currentColor">
    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
  </svg>
);

const GitHub = () => {
  const benefits = [
    "Workflow definitions tracked in GitHub",
    "Pull-request based review for critical changes",
    "Version history tied to execution records",
    "Promotion across dev, QA, and production",
  ];

  return (
    <section id="architecture" style={{
      padding: '4.5rem 2rem',
      background: 'var(--bg-primary)',
      borderTop: '1px solid var(--border-light)',
    }}>
      <div style={{
        display: 'grid',
        gridTemplateColumns: '1fr 1fr',
        gap: '3rem',
        maxWidth: '1100px',
        margin: '0 auto',
        alignItems: 'center',
      }}>
        {/* Copy */}
        <div>
          <h2 style={{
            fontSize: '1.75rem',
            fontWeight: 500,
            marginBottom: '1rem',
            color: 'var(--text-primary)',
          }}>
            Built for teams that already work in GitHub.
          </h2>
          <p style={{
            color: 'var(--text-secondary)',
            marginBottom: '1.5rem',
            lineHeight: 1.7,
          }}>
            Passflow treats workflows as versioned operational assets. Review changes, track history, and promote workflows with the same discipline as production software.
          </p>
          <ul style={{
            color: 'var(--text-secondary)',
            fontSize: '0.9rem',
            listStyle: 'none',
            display: 'flex',
            flexDirection: 'column',
            gap: '0.5rem',
          }}>
            {benefits.map((benefit) => (
              <li key={benefit}>
                <span style={{ color: 'var(--color-success)' }}>✓</span> {benefit}
              </li>
            ))}
          </ul>
        </div>

        {/* GitHub mockup */}
        <div style={{
          background: 'var(--mockup-bg)',
          border: '1px solid var(--border-dark)',
          borderRadius: '10px',
          padding: '1.25rem',
          fontSize: '0.75rem',
          color: 'var(--text-on-dark)',
          boxShadow: 'var(--shadow-card)',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem' }}>
            <GitHubIcon />
            <span>passflow-ai/workflows</span>
          </div>
          <div style={{
            background: 'var(--mockup-surface)',
            border: '1px solid var(--mockup-border)',
            borderRadius: '6px',
            padding: '0.75rem',
            marginBottom: '0.75rem',
          }}>
            <div style={{ color: '#9CA3AF', marginBottom: '0.2rem' }}>📄 kyc-onboarding.yaml</div>
            <div style={{ color: '#4B5563', fontSize: '0.68rem' }}>Updated 2 hours ago</div>
          </div>
          <div style={{
            background: 'var(--accent-primary-subtle)',
            border: '1px solid rgba(124,58,237,0.25)',
            borderRadius: '6px',
            padding: '0.75rem',
            marginBottom: '0.75rem',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.2rem' }}>
              <span style={{ color: 'var(--accent-primary)' }}>PR #142</span>
              <span>Add approval gate</span>
            </div>
            <div style={{ color: 'var(--color-success)', fontSize: '0.68rem' }}>✓ Approved · Ready to merge</div>
          </div>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <span style={{
              background: 'var(--color-success)',
              color: '#fff',
              padding: '0.2rem 0.5rem',
              borderRadius: '4px',
              fontSize: '0.62rem',
            }}>v1.4.2</span>
            <span style={{
              background: 'var(--mockup-border)',
              color: '#9CA3AF',
              padding: '0.2rem 0.5rem',
              borderRadius: '4px',
              fontSize: '0.62rem',
            }}>production</span>
          </div>
        </div>
      </div>
    </section>
  );
};

export default GitHub;
```

- [ ] **Step 2: Commit GitHub section**

```bash
git add src/components/GitHub.tsx
git commit -m "feat: add GitHub section with PR mockup"
```

---

### Task 10: Create OpenVsEnterprise Section

**Files:**
- Create: `src/components/OpenVsEnterprise.tsx`

- [ ] **Step 1: Create OpenVsEnterprise.tsx**

```tsx
const GitHubIcon = () => (
  <svg style={{ width: 16, height: 16 }} viewBox="0 0 24 24" fill="currentColor">
    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
  </svg>
);

const OpenVsEnterprise = () => {
  const openSource = [
    "Runtime base",
    "SDK and CLI",
    "Workflow spec",
    "Templates & connectors",
    "Local developer tooling",
  ];

  const enterprise = [
    "Control plane",
    "RBAC & approval policies",
    "Audit store & evidence",
    "Isolated execution",
    "SSO / enterprise security",
  ];

  return (
    <section style={{
      padding: '4.5rem 2rem',
      background: 'var(--bg-surface)',
      borderTop: '1px solid var(--border-light)',
    }}>
      <div style={{ maxWidth: '900px', margin: '0 auto' }}>
        <h2 style={{
          fontSize: '1.75rem',
          fontWeight: 500,
          marginBottom: '0.5rem',
          textAlign: 'center',
          color: 'var(--text-primary)',
        }}>
          Open where transparency matters.<br />Enterprise where control matters.
        </h2>
        <p style={{
          color: 'var(--text-secondary)',
          textAlign: 'center',
          marginBottom: '3rem',
        }}>
          Developer trust + business accountability in the same platform.
        </p>
        <div style={{
          display: 'grid',
          gridTemplateColumns: '1fr 1fr',
          gap: '1.5rem',
        }}>
          {/* Open Source */}
          <div className="card" style={{ padding: '1.75rem' }}>
            <h3 style={{
              fontSize: '1rem',
              fontWeight: 500,
              marginBottom: '1rem',
              display: 'flex',
              alignItems: 'center',
              gap: '0.5rem',
              color: 'var(--text-primary)',
            }}>
              <GitHubIcon />
              Open source foundation
            </h3>
            <ul style={{
              color: 'var(--text-secondary)',
              fontSize: '0.85rem',
              listStyle: 'none',
              display: 'flex',
              flexDirection: 'column',
              gap: '0.4rem',
            }}>
              {openSource.map((item) => (
                <li key={item}>• {item}</li>
              ))}
            </ul>
          </div>

          {/* Enterprise */}
          <div className="card-highlight">
            <h3 style={{
              fontSize: '1rem',
              fontWeight: 500,
              marginBottom: '1rem',
              color: 'var(--accent-primary)',
            }}>
              Enterprise control layer
            </h3>
            <ul style={{
              color: 'var(--text-secondary)',
              fontSize: '0.85rem',
              listStyle: 'none',
              display: 'flex',
              flexDirection: 'column',
              gap: '0.4rem',
            }}>
              {enterprise.map((item) => (
                <li key={item}>• {item}</li>
              ))}
            </ul>
          </div>
        </div>
      </div>
    </section>
  );
};

export default OpenVsEnterprise;
```

- [ ] **Step 2: Commit OpenVsEnterprise**

```bash
git add src/components/OpenVsEnterprise.tsx
git commit -m "feat: add OpenVsEnterprise comparison section"
```

---

### Task 11: Rewrite FinalCTA Component

**Files:**
- Modify: `src/components/FinalCTA.tsx`

- [ ] **Step 1: Replace FinalCTA**

```tsx
const GitHubIcon = () => (
  <svg style={{ width: 16, height: 16 }} viewBox="0 0 24 24" fill="currentColor">
    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
  </svg>
);

const FinalCTA = () => {
  const trustPoints = [
    "Open foundation",
    "GitHub-native",
    "Enterprise control",
  ];

  return (
    <section style={{
      padding: '5rem 2rem',
      background: 'linear-gradient(180deg, var(--bg-primary) 0%, var(--bg-alternate) 100%)',
      textAlign: 'center',
      borderTop: '1px solid var(--border-light)',
    }}>
      <h2 style={{
        fontSize: '2rem',
        fontWeight: 500,
        marginBottom: '0.75rem',
        color: 'var(--text-primary)',
      }}>
        Build openly. Operate responsibly.
      </h2>
      <p style={{
        color: 'var(--text-secondary)',
        marginBottom: '2rem',
        maxWidth: '480px',
        marginLeft: 'auto',
        marginRight: 'auto',
      }}>
        Move from experimental AI flows to governed business operations.
      </p>
      <div style={{
        display: 'flex',
        gap: '0.75rem',
        justifyContent: 'center',
        marginBottom: '2rem',
      }}>
        <a href="/demo" className="btn-primary" style={{ padding: '1rem 2rem', fontSize: '1rem' }}>
          Book demo
        </a>
        <a
          href="https://github.com/passflow-ai/passflow"
          target="_blank"
          rel="noopener noreferrer"
          className="btn-secondary"
          style={{ padding: '1rem 2rem', fontSize: '1rem' }}
        >
          <GitHubIcon />
          Start on GitHub
        </a>
      </div>
      <div style={{
        display: 'flex',
        gap: '2rem',
        justifyContent: 'center',
        color: 'var(--text-secondary)',
        fontSize: '0.85rem',
      }}>
        {trustPoints.map((point) => (
          <span key={point}>✦ {point}</span>
        ))}
      </div>
    </section>
  );
};

export default FinalCTA;
```

- [ ] **Step 2: Commit FinalCTA**

```bash
git add src/components/FinalCTA.tsx
git commit -m "style: rewrite FinalCTA with light-premium design"
```

---

### Task 12: Rewrite Footer Component

**Files:**
- Modify: `src/components/Footer.tsx`

- [ ] **Step 1: Replace Footer with 4-column layout**

```tsx
const GitHubIcon = () => (
  <svg style={{ width: 20, height: 20 }} viewBox="0 0 24 24" fill="currentColor">
    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
  </svg>
);

const TwitterIcon = () => (
  <svg style={{ width: 20, height: 20 }} viewBox="0 0 24 24" fill="currentColor">
    <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
  </svg>
);

const LinkedInIcon = () => (
  <svg style={{ width: 20, height: 20 }} viewBox="0 0 24 24" fill="currentColor">
    <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/>
  </svg>
);

const Footer = () => {
  const columns = [
    {
      title: "Product",
      links: [
        { label: "Features", href: "#how-it-works" },
        { label: "Use Cases", href: "#use-cases" },
        { label: "Architecture", href: "#architecture" },
        { label: "Roadmap", href: "/roadmap" },
      ],
    },
    {
      title: "Developers",
      links: [
        { label: "Documentation", href: "/docs" },
        { label: "GitHub", href: "https://github.com/passflow-ai/passflow" },
        { label: "API Reference", href: "/docs/api" },
        { label: "Templates", href: "/templates" },
      ],
    },
    {
      title: "Company",
      links: [
        { label: "About", href: "/about" },
        { label: "Blog", href: "/blog" },
        { label: "Careers", href: "/careers" },
        { label: "Contact", href: "/contact" },
      ],
    },
    {
      title: "Legal",
      links: [
        { label: "Privacy", href: "/privacy" },
        { label: "Terms", href: "/terms" },
        { label: "Security", href: "/security" },
        { label: "Compliance", href: "/compliance" },
      ],
    },
  ];

  return (
    <footer style={{
      background: 'var(--bg-surface)',
      borderTop: '1px solid var(--border-light)',
      padding: '4rem 2rem 2rem',
    }}>
      <div style={{ maxWidth: '1100px', margin: '0 auto' }}>
        {/* Columns */}
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(4, 1fr)',
          gap: '2rem',
          marginBottom: '3rem',
        }}>
          {columns.map((column) => (
            <div key={column.title}>
              <h4 style={{
                fontSize: '0.85rem',
                fontWeight: 500,
                color: 'var(--text-primary)',
                marginBottom: '1rem',
              }}>
                {column.title}
              </h4>
              <ul style={{
                listStyle: 'none',
                display: 'flex',
                flexDirection: 'column',
                gap: '0.6rem',
              }}>
                {column.links.map((link) => (
                  <li key={link.label}>
                    <a
                      href={link.href}
                      style={{
                        color: 'var(--text-secondary)',
                        fontSize: '0.8rem',
                        textDecoration: 'none',
                      }}
                    >
                      {link.label}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        {/* Copyright row */}
        <div style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          paddingTop: '2rem',
          borderTop: '1px solid var(--border-light)',
        }}>
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: '0.5rem',
          }}>
            <div style={{
              width: '6px',
              height: '6px',
              background: 'var(--accent-primary)',
              borderRadius: '1px',
            }} />
            <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
              © 2026 Passflow
            </span>
          </div>
          <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
            Open source under Apache 2.0
          </span>
          <div style={{ display: 'flex', gap: '1rem' }}>
            <a
              href="https://github.com/passflow-ai/passflow"
              target="_blank"
              rel="noopener noreferrer"
              style={{ color: 'var(--text-secondary)' }}
            >
              <GitHubIcon />
            </a>
            <a
              href="https://twitter.com/passflow_ai"
              target="_blank"
              rel="noopener noreferrer"
              style={{ color: 'var(--text-secondary)' }}
            >
              <TwitterIcon />
            </a>
            <a
              href="https://linkedin.com/company/passflow"
              target="_blank"
              rel="noopener noreferrer"
              style={{ color: 'var(--text-secondary)' }}
            >
              <LinkedInIcon />
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
```

- [ ] **Step 2: Commit Footer**

```bash
git add src/components/Footer.tsx
git commit -m "style: rewrite Footer with 4-column enterprise layout"
```

---

### Task 13: Wire All P1 Sections

**Files:**
- Modify: `src/app/[locale]/page.tsx`

- [ ] **Step 1: Update page with all sections**

```tsx
import Header from "@/components/Header";
import Hero from "@/components/Hero";
import TrustBand from "@/components/TrustBand";
import Problem from "@/components/Problem";
import ValueProp from "@/components/ValueProp";
import GitHub from "@/components/GitHub";
import OpenVsEnterprise from "@/components/OpenVsEnterprise";
import FinalCTA from "@/components/FinalCTA";
import Footer from "@/components/Footer";

export default function Home() {
  return (
    <>
      <Header />
      <main>
        <Hero />
        <TrustBand />
        <Problem />
        <ValueProp />
        <GitHub />
        <OpenVsEnterprise />
        <FinalCTA />
      </main>
      <Footer />
    </>
  );
}
```

- [ ] **Step 2: Verify full page renders**

Run: `npm run dev`
Expected: All 8 sections visible with light-premium design

- [ ] **Step 3: Delete deprecated components**

```bash
rm src/components/Urgency.tsx
rm src/components/HowItWorks.tsx
rm src/components/Differentiation.tsx
rm src/components/Trust.tsx
rm src/components/OpenSource.tsx
rm src/components/UseCases.tsx
rm src/components/Pricing.tsx
```

- [ ] **Step 4: Commit full page**

```bash
git add -A
git commit -m "feat: complete P1 - all sections wired

- Problem, ValueProp, GitHub, OpenVsEnterprise sections
- Remove deprecated components
- Full light-premium landing page
"
```

---

## P2 — Polish

### Task 14: Add Responsive Styles

**Files:**
- Modify: `src/app/globals.css`

- [ ] **Step 1: Add responsive utilities**

Add at end of globals.css:

```css
/* --------------------------------------------------------
   RESPONSIVE
   -------------------------------------------------------- */

@media (max-width: 1024px) {
  .text-hero {
    font-size: 2.5rem;
  }
}

@media (max-width: 768px) {
  .text-hero {
    font-size: 2rem;
  }

  .section {
    padding: 3rem 1.5rem;
  }
}

@media (max-width: 640px) {
  .text-hero {
    font-size: 1.75rem;
  }

  .section {
    padding: 2.5rem 1rem;
  }
}
```

- [ ] **Step 2: Update Hero for mobile**

Add responsive grid to Hero.tsx:

```tsx
// Update grid style
style={{
  display: 'grid',
  gridTemplateColumns: '1fr 1.15fr',
  gap: '4rem',
  // ...
}}
className="lg:grid-cols-[1fr_1.15fr] grid-cols-1"
```

- [ ] **Step 3: Commit responsive**

```bash
git add src/app/globals.css src/components/Hero.tsx
git commit -m "style: add responsive breakpoints for mobile/tablet"
```

---

### Task 15: Update DESIGN-SYSTEM.md

**Files:**
- Modify: `DESIGN-SYSTEM.md`

- [ ] **Step 1: Replace with new design system**

Reference the spec document for full content:
`docs/superpowers/specs/2026-04-09-passflow-web-redesign-design.md`

- [ ] **Step 2: Commit DESIGN-SYSTEM.md**

```bash
git add DESIGN-SYSTEM.md
git commit -m "docs: update DESIGN-SYSTEM.md with light-premium system"
```

---

### Task 16: Final Verification

- [ ] **Step 1: Run full build**

```bash
npm run build
```

Expected: No errors

- [ ] **Step 2: Test in browser**

- Open localhost:3000
- Verify all sections render correctly
- Check mobile responsiveness
- Verify GitHub links work
- Test language switcher

- [ ] **Step 3: Create final commit**

```bash
git add -A
git commit -m "chore: passflow-web light-premium redesign complete

Full redesign from dark terminal to light-premium:
- New color system (warm gray, violet accent)
- 56px dominant hero headline
- Dark product mockups
- GitHub-native narrative
- Enterprise footer
- Responsive breakpoints
"
```

---

## Summary

| Phase | Tasks | Description |
|-------|-------|-------------|
| **P0** | 1-6 | CSS foundation, ProductMockup, TrustBand, Header, Hero, page wiring |
| **P1** | 7-13 | Problem, ValueProp, GitHub, OpenVsEnterprise, FinalCTA, Footer, full page |
| **P2** | 14-16 | Responsive, DESIGN-SYSTEM.md, verification |

Total: 16 tasks with ~80 steps
