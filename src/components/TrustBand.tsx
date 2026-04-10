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
