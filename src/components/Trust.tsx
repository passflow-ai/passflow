"use client";

import { useTranslations } from "next-intl";
import { Globe, Lock, FileText, Shield } from "lucide-react";

const Trust = () => {
  const t = useTranslations("Trust");

  const features = [
    {
      icon: Globe,
      key: "dataResidency",
    },
    {
      icon: Lock,
      key: "encryptedCredentials",
    },
    {
      icon: FileText,
      key: "auditTrail",
    },
    {
      icon: Shield,
      key: "rbacApprovals",
    },
  ];

  return (
    <section id="trust" className="section bg-[var(--bg-primary)]">
      <div className="container">
        {/* Header */}
        <div className="max-w-4xl mx-auto text-center mb-16">
          <h2 className="heading-2 mb-6">{t("headline")}</h2>
          <p className="body-text max-w-3xl mx-auto">{t("subheadline")}</p>
        </div>

        {/* Features Grid - 4 columns */}
        <div className="grid-4">
          {features.map((feature, index) => {
            const Icon = feature.icon;
            return (
              <div key={index} className="card group">
                {/* Icon */}
                <div className="w-12 h-12 rounded-lg bg-[var(--accent-primary-ghost)] border border-[rgba(0,255,135,0.2)] flex items-center justify-center mb-6 transition-all duration-300 group-hover:bg-[var(--accent-primary-glow)]">
                  <Icon
                    className="w-6 h-6 text-[var(--accent-primary)]"
                    strokeWidth={1.5}
                  />
                </div>

                {/* Title */}
                <h4 className="heading-4 mb-3">
                  {t(`features.${feature.key}.title`)}
                </h4>

                {/* Description */}
                <p className="body-text-sm mb-4">
                  {t(`features.${feature.key}.description`)}
                </p>

                {/* Technical Detail */}
                <p className="code-text text-[var(--text-tertiary)] text-xs leading-relaxed">
                  {t(`features.${feature.key}.technical`)}
                </p>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
};

export default Trust;
