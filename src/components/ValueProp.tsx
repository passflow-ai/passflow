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
