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

      {/* Main content - workflow canvas + side panel */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 240px', fontSize: '0.72rem', color: 'var(--text-on-dark)' }}>
        {/* Canvas with workflow nodes */}
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
                <span style={{ color: '#6B7280' }}>Version</span><span>v1.4.2</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: '#6B7280' }}>GitHub</span><span style={{ color: 'var(--accent-primary)' }}>9f3a2b1</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: '#6B7280' }}>Env</span><span style={{ color: 'var(--color-success)' }}>prod</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: '#6B7280' }}>Risk</span><span style={{ color: 'var(--color-warning)' }}>High</span>
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
