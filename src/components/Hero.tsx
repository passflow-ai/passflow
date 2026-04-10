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
