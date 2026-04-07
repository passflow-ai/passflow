"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Check, Info } from "lucide-react";

const Pricing = () => {
  const t = useTranslations("Pricing");
  const [isAnnual, setIsAnnual] = useState(false);

  const scrollToContact = () => {
    const element = document.getElementById("contact");
    if (element) {
      element.scrollIntoView({ behavior: "smooth" });
    }
  };

  const handleSignUp = () => {
    window.location.href = "https://app.passflow.ai/login?mode=register";
  };

  const planKeys = ["free", "starter", "growth", "business"] as const;

  const planConfigs: Record<
    (typeof planKeys)[number],
    {
      price: number;
      highlighted: boolean;
      action: () => void;
      featuresWithTooltips?: number[];
    }
  > = {
    free: {
      price: 0,
      highlighted: false,
      action: handleSignUp,
    },
    starter: {
      price: 99,
      highlighted: false,
      action: handleSignUp,
      featuresWithTooltips: [4], // BYOK feature index
    },
    growth: {
      price: 299,
      highlighted: true,
      action: handleSignUp,
    },
    business: {
      price: 799,
      highlighted: false,
      action: scrollToContact,
    },
  };

  const formatPrice = (price: number) => {
    if (price === 0) return "$0";
    const finalPrice = isAnnual ? Math.round((price * 10) / 12) : price;
    return `$${finalPrice}`;
  };

  const Tooltip = ({ content }: { content: string }) => (
    <span className="group/tooltip relative inline-flex ml-1 cursor-help">
      <Info className="w-3.5 h-3.5 text-[var(--text-tertiary)] hover:text-[var(--accent-primary)] transition-colors" />
      <span className="invisible group-hover/tooltip:visible absolute left-1/2 -translate-x-1/2 bottom-full mb-2 w-64 p-3 bg-[var(--bg-surface)] border border-[var(--border-primary)] rounded-lg text-xs text-[var(--text-secondary)] z-10 shadow-lg">
        {content}
      </span>
    </span>
  );

  return (
    <section id="pricing" className="section bg-[var(--bg-primary)]">
      <div className="container">
        {/* Header */}
        <div className="max-w-4xl mx-auto text-center mb-12">
          <h2 className="heading-2 mb-6">{t("headline")}</h2>
          <p className="body-text max-w-3xl mx-auto mb-8">{t("subheadline")}</p>

          {/* Billing Toggle */}
          <div className="flex items-center justify-center gap-4">
            <span
              className={`label transition-colors ${
                !isAnnual
                  ? "text-[var(--text-primary)]"
                  : "text-[var(--text-tertiary)]"
              }`}
            >
              {t("billing.monthly")}
            </span>
            <button
              onClick={() => setIsAnnual(!isAnnual)}
              className={`relative w-14 h-7 rounded-full transition-colors duration-300 ${
                isAnnual
                  ? "bg-[var(--accent-primary)]"
                  : "bg-[var(--bg-surface)] border border-[var(--border-primary)]"
              }`}
              aria-label="Cambiar entre facturacion mensual y anual"
            >
              <span
                className={`absolute top-1 w-5 h-5 rounded-full transition-all duration-300 ${
                  isAnnual
                    ? "left-8 bg-[var(--text-inverse)]"
                    : "left-1 bg-[var(--text-secondary)]"
                }`}
              />
            </button>
            <span
              className={`label transition-colors ${
                isAnnual
                  ? "text-[var(--text-primary)]"
                  : "text-[var(--text-tertiary)]"
              }`}
            >
              {t("billing.annual")}
            </span>
            {isAnnual && (
              <span className="badge badge-green">{t("billing.annualBadge")}</span>
            )}
          </div>
        </div>

        {/* Pricing Cards */}
        <div className="grid-4 max-w-6xl mx-auto mb-12">
          {planKeys.map((planKey, index) => {
            const config = planConfigs[planKey];
            const features = t.raw(`plans.${planKey}.features`) as string[];
            const ctaSubtext = t.has(`plans.${planKey}.ctaSubtext`)
              ? t(`plans.${planKey}.ctaSubtext`)
              : null;
            const badge = t.has(`plans.${planKey}.badge`)
              ? t(`plans.${planKey}.badge`)
              : null;

            return (
              <div
                key={index}
                className={`card relative ${
                  config.highlighted ? "card-recommended" : ""
                }`}
              >
                {/* Recommended Badge */}
                {badge && (
                  <span className="absolute -top-3 left-1/2 -translate-x-1/2 badge bg-[var(--accent-primary)] text-[var(--text-inverse)] border-0 px-4 py-1">
                    {badge}
                  </span>
                )}

                {/* Plan Name */}
                <h4 className="heading-4 mb-2">{t(`plans.${planKey}.name`)}</h4>

                {/* Price */}
                <div className="flex items-baseline gap-1 mb-4">
                  <span
                    className="font-[family-name:var(--font-display)] text-[var(--text-5xl)] font-bold text-[var(--text-primary)]"
                    style={{ fontSize: "var(--text-5xl)" }}
                  >
                    {formatPrice(config.price)}
                  </span>
                  <span className="body-text-sm">{t("perMonth")}</span>
                  {isAnnual && config.price > 0 && (
                    <Tooltip content={t("tooltips.tokens")} />
                  )}
                </div>

                {/* Description */}
                <p className="body-text-sm mb-6">
                  {t(`plans.${planKey}.description`)}
                </p>

                {/* Features */}
                <ul className="space-y-3 mb-8">
                  {features.map((feature, i) => {
                    const hasTooltip =
                      config.featuresWithTooltips?.includes(i) ?? false;

                    return (
                      <li key={i} className="flex items-start gap-3">
                        <Check
                          className="w-5 h-5 flex-shrink-0 mt-0.5 text-[var(--accent-primary)]"
                          strokeWidth={2}
                        />
                        <span className="body-text-sm text-[var(--text-primary)]">
                          {feature}
                          {hasTooltip && (
                            <Tooltip content={t("tooltips.byok")} />
                          )}
                        </span>
                      </li>
                    );
                  })}
                </ul>

                {/* CTA Button */}
                <button
                  onClick={config.action}
                  className={`w-full ${
                    config.highlighted ? "btn-primary" : "btn-secondary"
                  }`}
                >
                  {t(`plans.${planKey}.cta`)}
                </button>

                {/* CTA Subtext */}
                {ctaSubtext && (
                  <p className="mt-3 text-center text-xs text-[var(--text-tertiary)]">
                    {ctaSubtext}
                  </p>
                )}
              </div>
            );
          })}
        </div>

        {/* Footer Note */}
        <p className="text-center label text-[var(--text-tertiary)]">
          {t("footnote")}
        </p>
      </div>
    </section>
  );
};

export default Pricing;
