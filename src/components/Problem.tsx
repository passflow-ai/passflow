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
