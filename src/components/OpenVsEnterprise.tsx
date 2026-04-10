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
