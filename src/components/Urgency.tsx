'use client';

import { useTranslations } from 'next-intl';

const columnConfig = [
  { key: 'infrastructure', badgeColor: 'badge-green' },
  { key: 'operations', badgeColor: 'badge-blue' },
  { key: 'sales', badgeColor: 'badge-neutral' },
];

const Urgency = () => {
  const t = useTranslations('Urgency');

  return (
    <section className="section" style={{ background: 'var(--bg-primary)' }}>
      <div className="container">
        {/* Header */}
        <div className="text-center" style={{ marginBottom: 'var(--space-16)' }}>
          <h2 className="heading-2" style={{ marginBottom: 'var(--space-6)' }}>
            {t('headline')}
          </h2>
          <p
            className="body-text"
            style={{ maxWidth: 'var(--max-width-narrow)', margin: '0 auto' }}
          >
            {t('subheadline')}
          </p>
        </div>

        {/* 3 Column Grid */}
        <div className="grid-3">
          {columnConfig.map((column) => (
            <article
              key={column.key}
              className="card"
              style={{
                display: 'flex',
                flexDirection: 'column',
                height: '100%',
              }}
            >
              {/* Badge */}
              <span
                className={`badge ${column.badgeColor}`}
                style={{ alignSelf: 'flex-start', marginBottom: 'var(--space-4)' }}
              >
                {t(`columns.${column.key}.badge`)}
              </span>

              {/* Title */}
              <h3 className="heading-4" style={{ marginBottom: 'var(--space-4)' }}>
                {t(`columns.${column.key}.title`)}
              </h3>

              {/* Narrative */}
              <p
                className="body-text-sm"
                style={{
                  marginBottom: 'var(--space-6)',
                  flex: 1,
                }}
              >
                {t(`columns.${column.key}.narrative`)}
              </p>

              {/* Stat */}
              <div
                style={{
                  padding: 'var(--space-4)',
                  background: 'var(--bg-surface)',
                  borderRadius: 'var(--radius-md)',
                  marginBottom: 'var(--space-3)',
                }}
              >
                <p
                  className="code-text"
                  style={{
                    color: 'var(--color-warning)',
                    marginBottom: 0,
                  }}
                >
                  {t(`columns.${column.key}.stat`)}
                </p>
              </div>

              {/* Label */}
              <p
                className="label"
                style={{
                  color: 'var(--accent-primary)',
                  marginBottom: 0,
                }}
              >
                {t(`columns.${column.key}.label`)}
              </p>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
};

export default Urgency;
