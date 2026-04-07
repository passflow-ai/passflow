"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";

interface UseCaseConfig {
  id: string;
  key: "k8sOps" | "outboundB2B" | "gitops";
  badgeVariant: "green" | "blue" | "neutral";
}

const useCaseConfigs: UseCaseConfig[] = [
  {
    id: "k8s-ops",
    key: "k8sOps",
    badgeVariant: "green",
  },
  {
    id: "outbound-b2b",
    key: "outboundB2B",
    badgeVariant: "blue",
  },
  {
    id: "gitops",
    key: "gitops",
    badgeVariant: "neutral",
  },
];

const ChevronDownIcon = () => (
  <svg
    className="w-5 h-5"
    fill="none"
    stroke="currentColor"
    viewBox="0 0 24 24"
    strokeWidth={2}
  >
    <path strokeLinecap="round" strokeLinejoin="round" d="M19 9l-7 7-7-7" />
  </svg>
);

const UseCases = () => {
  const t = useTranslations("UseCases");
  const [expandedCard, setExpandedCard] = useState<string | null>(null);

  const toggleCard = (id: string) => {
    setExpandedCard(expandedCard === id ? null : id);
  };

  const getBadgeClass = (variant: "green" | "blue" | "neutral") => {
    const variants = {
      green: "badge-green",
      blue: "badge-blue",
      neutral: "badge-neutral",
    };
    return `badge ${variants[variant]}`;
  };

  // Get tools array from translation
  const getTools = (key: string): string[] => {
    const toolsRaw = t.raw(`cards.${key}.tools`);
    return Array.isArray(toolsRaw) ? toolsRaw : [];
  };

  return (
    <section id="use-cases" className="section">
      <div className="container">
        {/* Header */}
        <div className="text-center mb-12 md:mb-16">
          <h2 className="heading-2 mb-6">
            {t("headline")}
          </h2>
          <p className="body-text max-w-3xl mx-auto">
            {t("subheadline")}
          </p>
        </div>

        {/* Use Case Cards */}
        <div className="grid-3 max-w-5xl mx-auto">
          {useCaseConfigs.map((config) => {
            const isExpanded = expandedCard === config.id;
            const tools = getTools(config.key);

            return (
              <div
                key={config.id}
                className={`card-expandable ${isExpanded ? "expanded" : ""}`}
                onClick={() => toggleCard(config.id)}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    toggleCard(config.id);
                  }
                }}
                aria-expanded={isExpanded}
              >
                {/* Card Header */}
                <div className="flex items-start justify-between gap-4 mb-4">
                  <span className={getBadgeClass(config.badgeVariant)}>
                    {t(`cards.${config.key}.badge`)}
                  </span>
                  <span
                    className={`expand-icon ${isExpanded ? "rotated" : ""}`}
                    style={{ color: "var(--text-tertiary)" }}
                  >
                    <ChevronDownIcon />
                  </span>
                </div>

                {/* Title */}
                <h3 className="heading-4 mb-3">{t(`cards.${config.key}.title`)}</h3>

                {/* Summary (always visible) */}
                <p className="body-text-sm">{t(`cards.${config.key}.summary`)}</p>

                {/* Expandable Content */}
                <div className="card-expandable-content">
                  {/* Detail */}
                  <div
                    className="pt-4 border-t"
                    style={{ borderColor: "var(--border-primary)" }}
                  >
                    <p className="body-text-sm mb-4">{t(`cards.${config.key}.detail`)}</p>

                    {/* Tools */}
                    <div className="mb-4">
                      <span
                        className="label-uppercase block mb-2"
                        style={{ fontSize: "var(--text-xs)" }}
                      >
                        {t("toolsLabel")}
                      </span>
                      <div className="flex flex-wrap gap-2">
                        {tools.map((tool, index) => (
                          <span
                            key={index}
                            className="badge badge-neutral"
                            style={{
                              fontSize: "var(--text-xs)",
                              padding: "var(--space-0-5) var(--space-2)",
                            }}
                          >
                            {tool}
                          </span>
                        ))}
                      </div>
                    </div>

                    {/* Result */}
                    <div
                      className="p-3 rounded-md"
                      style={{
                        background: "var(--accent-primary-ghost)",
                        border: "1px solid rgba(0, 255, 135, 0.15)",
                      }}
                    >
                      <span
                        className="label-uppercase block mb-1"
                        style={{
                          fontSize: "var(--text-xs)",
                          color: "var(--accent-primary)",
                        }}
                      >
                        {t("resultLabel")}
                      </span>
                      <span
                        className="block"
                        style={{
                          fontFamily: "var(--font-mono)",
                          fontSize: "var(--text-sm)",
                          color: "var(--text-primary)",
                          fontWeight: "var(--weight-medium)",
                        }}
                      >
                        {t(`cards.${config.key}.result`)}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
};

export default UseCases;
