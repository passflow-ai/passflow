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
