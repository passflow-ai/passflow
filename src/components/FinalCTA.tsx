"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";

const FinalCTA = () => {
  const t = useTranslations("FinalCTA");

  return (
    <section id="final-cta" className="section" style={{ background: 'var(--gradient-hero)' }}>
      <div className="container">
        {/* Section Header */}
        <div className="text-center mb-16">
          <h2 className="heading-2 mb-6">
            {t("headline")}
          </h2>
          <p className="body-text max-w-2xl mx-auto">
            {t("subheadline")}
          </p>
        </div>

        {/* Steps */}
        <div className="grid-3 max-w-4xl mx-auto mb-16">
          {/* Step 1 - Connect */}
          <div className="card text-center">
            <div className="w-16 h-16 rounded-full mx-auto mb-6 flex items-center justify-center" style={{ background: 'var(--accent-primary-ghost)', border: '1px solid rgba(0, 255, 135, 0.2)' }}>
              <svg
                className="w-8 h-8"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                style={{ color: 'var(--accent-primary)' }}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
                />
              </svg>
            </div>
            <span className="badge badge-green mb-4">{t("steps.connect.time")}</span>
            <h3 className="heading-4 mb-3">{t("steps.connect.title")}</h3>
            <p className="body-text-sm">
              {t("steps.connect.description")}
            </p>
          </div>

          {/* Step 2 - Configure */}
          <div className="card text-center">
            <div className="w-16 h-16 rounded-full mx-auto mb-6 flex items-center justify-center" style={{ background: 'var(--accent-primary-ghost)', border: '1px solid rgba(0, 255, 135, 0.2)' }}>
              <svg
                className="w-8 h-8"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                style={{ color: 'var(--accent-primary)' }}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                />
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                />
              </svg>
            </div>
            <span className="badge badge-green mb-4">{t("steps.configure.time")}</span>
            <h3 className="heading-4 mb-3">{t("steps.configure.title")}</h3>
            <p className="body-text-sm">
              {t("steps.configure.description")}
            </p>
          </div>

          {/* Step 3 - Deploy */}
          <div className="card text-center">
            <div className="w-16 h-16 rounded-full mx-auto mb-6 flex items-center justify-center" style={{ background: 'var(--accent-primary-ghost)', border: '1px solid rgba(0, 255, 135, 0.2)' }}>
              <svg
                className="w-8 h-8"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                style={{ color: 'var(--accent-primary)' }}
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M13 10V3L4 14h7v7l9-11h-7z"
                />
              </svg>
            </div>
            <span className="badge badge-green mb-4">{t("steps.deploy.time")}</span>
            <h3 className="heading-4 mb-3">{t("steps.deploy.title")}</h3>
            <p className="body-text-sm">
              {t("steps.deploy.description")}
            </p>
          </div>
        </div>

        {/* CTAs */}
        <div className="text-center mb-8">
          <div className="flex flex-col sm:flex-row gap-4 justify-center mb-6">
            <a
              href="https://app.passflow.ai/login?mode=register"
              className="btn-primary"
            >
              {t("cta.primary")}
            </a>
            <Link href="/demo" className="btn-secondary">
              {t("cta.secondary")}
            </Link>
          </div>
          <p className="body-text-sm" style={{ color: 'var(--text-tertiary)' }}>
            {t("microcopy")}
          </p>
        </div>

        {/* Trust Badges */}
        <div className="flex flex-wrap items-center justify-center gap-6 mt-12">
          <div className="flex items-center gap-2" style={{ color: 'var(--text-secondary)' }}>
            <svg
              className="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              style={{ color: 'var(--accent-primary)' }}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
              />
            </svg>
            <span className="label">{t("trustBadges.encryption")}</span>
          </div>
          <div className="flex items-center gap-2" style={{ color: 'var(--text-secondary)' }}>
            <svg
              className="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              style={{ color: 'var(--accent-primary)' }}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
              />
            </svg>
            <span className="label">{t("trustBadges.soc2")}</span>
          </div>
          <div className="flex items-center gap-2" style={{ color: 'var(--text-secondary)' }}>
            <svg
              className="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              style={{ color: 'var(--accent-primary)' }}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span className="label">{t("trustBadges.dataResidency")}</span>
          </div>
        </div>
      </div>
    </section>
  );
};

export default FinalCTA;
