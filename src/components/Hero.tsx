"use client";

import { useTranslations } from "next-intl";

const Hero = () => {
  const t = useTranslations("Hero");

  const metrics = [
    {
      value: t("metrics.coldStart.value"),
      label: t("metrics.coldStart.label"),
      tooltip: t("metrics.coldStart.tooltip"),
    },
    {
      value: t("metrics.llmProviders.value"),
      label: t("metrics.llmProviders.label"),
      tooltip: t("metrics.llmProviders.tooltip"),
    },
    {
      value: t("metrics.mcpTools.value"),
      label: t("metrics.mcpTools.label"),
      tooltip: t("metrics.mcpTools.tooltip"),
    },
    {
      value: t("metrics.isolation.value"),
      label: t("metrics.isolation.label"),
      tooltip: t("metrics.isolation.tooltip"),
    },
  ];
  return (
    <section className="relative min-h-screen flex items-center overflow-hidden pt-16" style={{ background: "var(--gradient-hero)" }}>
      {/* Background effects */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="dot-grid" />
        <div className="radial-glow top-1/4 left-1/2 -translate-x-1/2" />
      </div>

      <div className="relative max-w-[var(--max-width-content)] mx-auto px-4 md:px-6 lg:px-8 py-24">
        <div className="max-w-4xl">
          {/* Headline */}
          <h1 className="heading-1 mb-6">
            {t("headline.part1")}{" "}
            <span className="text-[var(--accent-primary)]">{t("headline.part2")}</span>
          </h1>

          {/* Subheadline */}
          <p className="body-text text-xl md:text-2xl mb-10 max-w-2xl">
            {t("subheadline")}
          </p>

          {/* CTAs */}
          <div className="flex flex-col sm:flex-row gap-4 mb-4">
            <a href="/demo" className="btn-primary">
              {t("cta.primary")}
            </a>
            <a href="/docs" className="btn-secondary">
              {t("cta.secondary")}
            </a>
          </div>

          {/* Microcopy */}
          <p className="text-[var(--text-tertiary)] text-sm font-[var(--font-mono)] mb-12">
            {t("microcopy")}
          </p>

          {/* Metrics bar */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-6 pt-8 border-t border-[var(--border-primary)]">
            {metrics.map((metric) => (
              <div key={metric.label} className="group relative">
                <div className="text-2xl md:text-3xl font-bold text-[var(--accent-primary)] font-[var(--font-mono)]">
                  {metric.value}
                </div>
                <div className="text-sm text-[var(--text-secondary)] font-[var(--font-mono)] uppercase tracking-wider">
                  {metric.label}
                </div>
                {/* Tooltip */}
                <div className="absolute bottom-full left-0 mb-2 px-3 py-2 bg-[var(--bg-surface)] border border-[var(--border-primary)] rounded-lg text-xs text-[var(--text-secondary)] opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none w-48 z-10">
                  {metric.tooltip}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
};

export default Hero;
