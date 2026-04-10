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
