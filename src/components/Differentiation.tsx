"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";

interface TooltipProps {
  term: string;
  content: string;
}

const Tooltip = ({ term, content }: TooltipProps) => {
  return (
    <span className="tooltip-wrapper">
      {term}
      <span className="tooltip-trigger">?</span>
      <span className="tooltip-content">{content}</span>
    </span>
  );
};

// Row keys that map to the translation file structure
const rowKeys = [
  "isolation",
  "llmRouting",
  "reactLoop",
  "humanInTheLoop",
  "mcpNative",
  "eventDriven",
  "multiTenancy",
  "selfHost",
  "agentLanguage",
  "coldStart",
] as const;

type RowKey = (typeof rowKeys)[number];

// Glossary term keys
const glossaryKeys = ["circuitBreaker", "cel", "rbac"] as const;

const Differentiation = () => {
  const t = useTranslations("Differentiation");
  const [hoveredRow, setHoveredRow] = useState<number | null>(null);

  const renderCellContent = (value: string) => {
    if (value === "checkmark") {
      return <span className="checkmark">&#10003;</span>;
    }
    if (value === "No" || value === "No nativo") {
      return <span className="cross">{value}</span>;
    }
    return value;
  };

  // Check if a row has a tooltip
  const hasTooltip = (rowKey: RowKey): boolean => {
    try {
      const tooltip = t.raw(`table.rows.${rowKey}.tooltip`);
      return typeof tooltip === "string" && tooltip.length > 0;
    } catch {
      return false;
    }
  };

  return (
    <section id="differentiation" className="section">
      <div className="container">
        {/* Header */}
        <div className="text-center mb-12 md:mb-16">
          <h2 className="heading-2 mb-6">
            {t("headline")}
          </h2>
          <p className="body-text max-w-4xl mx-auto">
            {t("subheadline")}
          </p>
        </div>

        {/* Comparison Table */}
        <div className="table-wrapper">
          <table className="comparison-table">
            <thead>
              <tr>
                <th style={{ minWidth: "200px" }}>{t("table.headers.capability")}</th>
                <th className="highlight" style={{ minWidth: "180px" }}>{t("table.headers.passflow")}</th>
                <th style={{ minWidth: "140px" }}>{t("table.headers.n8n")}</th>
                <th style={{ minWidth: "140px" }}>{t("table.headers.make")}</th>
                <th style={{ minWidth: "140px" }}>{t("table.headers.zapier")}</th>
              </tr>
            </thead>
            <tbody>
              {rowKeys.map((rowKey, index) => (
                <tr
                  key={rowKey}
                  onMouseEnter={() => setHoveredRow(index)}
                  onMouseLeave={() => setHoveredRow(null)}
                >
                  <td className="capability-cell">
                    {hasTooltip(rowKey) ? (
                      <Tooltip
                        term={t(`table.rows.${rowKey}.capability`)}
                        content={t(`table.rows.${rowKey}.tooltip`)}
                      />
                    ) : (
                      t(`table.rows.${rowKey}.capability`)
                    )}
                  </td>
                  <td className="highlight">
                    {renderCellContent(t(`table.rows.${rowKey}.passflow`))}
                  </td>
                  <td>{renderCellContent(t(`table.rows.${rowKey}.n8n`))}</td>
                  <td>{renderCellContent(t(`table.rows.${rowKey}.make`))}</td>
                  <td>{renderCellContent(t(`table.rows.${rowKey}.zapier`))}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Footnote */}
        <p
          className="mt-6 text-center"
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: "var(--text-xs)",
            color: "var(--text-tertiary)",
          }}
        >
          {t("footnote")}
        </p>

        {/* Additional tooltips legend */}
        <div
          className="mt-8 p-4 rounded-lg"
          style={{
            background: "var(--bg-elevated)",
            border: "1px solid var(--border-primary)",
          }}
        >
          <h4
            className="mb-3"
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: "var(--text-xs)",
              textTransform: "uppercase",
              letterSpacing: "var(--tracking-widest)",
              color: "var(--text-secondary)",
            }}
          >
            {t("glossary.title")}
          </h4>
          <div className="grid gap-3 md:grid-cols-3">
            {glossaryKeys.map((key) => (
              <div key={key}>
                <span
                  style={{
                    fontFamily: "var(--font-mono)",
                    fontSize: "var(--text-sm)",
                    color: "var(--accent-primary)",
                    fontWeight: "var(--weight-medium)",
                  }}
                >
                  {t(`glossary.${key}.term`)}
                </span>
                <p
                  style={{
                    fontSize: "var(--text-sm)",
                    color: "var(--text-secondary)",
                    marginTop: "var(--space-1)",
                  }}
                >
                  {t(`glossary.${key}.definition`)}
                </p>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
};

export default Differentiation;
