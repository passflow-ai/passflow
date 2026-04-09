"use client";

import { useTranslations } from "next-intl";
import { Check, X } from "lucide-react";

const OpenSource = () => {
  const t = useTranslations("OpenSource");

  const openSourceFeatures = [
    "engineCore",
    "palCompiler",
    "basicTools",
    "cli",
    "selfHost",
    "communitySupport",
  ] as const;

  const cloudFeatures = [
    "managedHosting",
    "advancedUI",
    "ssoRbac",
    "auditCompliance",
    "prioritySupport",
    "premiumIntegrations",
    "sla",
    "dedicatedSuccess",
  ] as const;

  return (
    <section id="open-source" className="section">
      <div className="container">
        {/* Header */}
        <div className="text-center mb-12 md:mb-16">
          <div className="inline-flex items-center gap-2 mb-4">
            <span className="badge badge-green">{t("badge")}</span>
          </div>
          <h2 className="heading-2 mb-6">{t("headline")}</h2>
          <p className="body-text max-w-4xl mx-auto">{t("subheadline")}</p>
        </div>

        {/* Comparison Cards */}
        <div className="grid md:grid-cols-2 gap-8 max-w-5xl mx-auto mb-12">
          {/* Open Source Card */}
          <div className="card relative overflow-hidden">
            <div
              className="absolute inset-0 opacity-5"
              style={{
                background:
                  "linear-gradient(135deg, var(--accent-primary) 0%, transparent 50%)",
              }}
            />
            <div className="relative">
              <div className="flex items-center gap-3 mb-4">
                <svg
                  className="w-8 h-8 text-[var(--accent-primary)]"
                  fill="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
                </svg>
                <h3 className="heading-3 text-[var(--accent-primary)]">
                  {t("openSource.title")}
                </h3>
              </div>
              <p className="body-text-sm mb-6">{t("openSource.description")}</p>
              <p
                className="text-3xl font-bold mb-6"
                style={{ fontFamily: "var(--font-display)" }}
              >
                {t("openSource.price")}
              </p>

              <ul className="space-y-3 mb-8">
                {openSourceFeatures.map((feature) => (
                  <li key={feature} className="flex items-start gap-3">
                    <Check
                      className="w-5 h-5 flex-shrink-0 mt-0.5 text-[var(--accent-primary)]"
                      strokeWidth={2}
                    />
                    <span className="body-text-sm">
                      {t(`openSource.features.${feature}`)}
                    </span>
                  </li>
                ))}
              </ul>

              <a
                href="https://github.com/jaak-ai/passflow"
                target="_blank"
                rel="noopener noreferrer"
                className="btn-secondary w-full flex items-center justify-center gap-2"
              >
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
                </svg>
                {t("openSource.cta")}
              </a>
            </div>
          </div>

          {/* Cloud/Enterprise Card */}
          <div className="card card-recommended relative overflow-hidden">
            <div
              className="absolute inset-0 opacity-5"
              style={{
                background:
                  "linear-gradient(135deg, var(--accent-secondary) 0%, transparent 50%)",
              }}
            />
            <div className="relative">
              <div className="flex items-center gap-3 mb-4">
                <svg
                  className="w-8 h-8 text-[var(--accent-secondary)]"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 10-9.78 2.096A4.001 4.001 0 003 15z"
                  />
                </svg>
                <h3 className="heading-3 text-[var(--accent-secondary)]">
                  {t("cloud.title")}
                </h3>
              </div>
              <p className="body-text-sm mb-6">{t("cloud.description")}</p>
              <p
                className="text-3xl font-bold mb-6"
                style={{ fontFamily: "var(--font-display)" }}
              >
                {t("cloud.price")}
              </p>

              <ul className="space-y-3 mb-8">
                {cloudFeatures.map((feature) => (
                  <li key={feature} className="flex items-start gap-3">
                    <Check
                      className="w-5 h-5 flex-shrink-0 mt-0.5 text-[var(--accent-secondary)]"
                      strokeWidth={2}
                    />
                    <span className="body-text-sm">
                      {t(`cloud.features.${feature}`)}
                    </span>
                  </li>
                ))}
              </ul>

              <a
                href="https://app.passflow.ai/login?mode=register"
                className="btn-primary w-full"
              >
                {t("cloud.cta")}
              </a>
            </div>
          </div>
        </div>

        {/* Bottom note */}
        <div className="text-center">
          <p className="body-text-sm text-[var(--text-secondary)] max-w-2xl mx-auto">
            {t("note")}
          </p>
        </div>
      </div>
    </section>
  );
};

export default OpenSource;
