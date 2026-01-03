"use client";

const Pricing = () => {
  const scrollToContact = () => {
    const element = document.getElementById("contact");
    if (element) {
      element.scrollIntoView({ behavior: "smooth" });
    }
  };

  const tiers = [
    {
      name: "Starter",
      price: "$0.85 - $1.50",
      unit: "per verification",
      description: "For teams getting started with identity verification",
      features: [
        "Up to 10K verifications/month",
        "Document + Selfie verification",
        "Standard API access",
        "Email support",
        "99.5% uptime SLA",
      ],
      cta: "Get Started",
      highlighted: false,
    },
    {
      name: "Growth",
      price: "$0.65 - $0.95",
      unit: "per verification",
      description: "For scaling teams with higher volumes",
      features: [
        "10K - 100K verifications/month",
        "All Starter features",
        "Advanced liveness detection",
        "Webhooks & custom flows",
        "Priority support",
        "99.9% uptime SLA",
      ],
      cta: "Get a Quote",
      highlighted: true,
    },
    {
      name: "Enterprise",
      price: "$0.45 - $0.75",
      unit: "per verification",
      description: "For high-volume operations with custom needs",
      features: [
        "100K+ verifications/month",
        "All Growth features",
        "On-premise deployment option",
        "Custom integrations",
        "Dedicated success manager",
        "24/7 phone support",
        "99.99% uptime SLA",
      ],
      cta: "Contact Sales",
      highlighted: false,
    },
  ];

  return (
    <section id="pricing" className="section-padding bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-3xl mx-auto text-center mb-12">
          <h3 className="text-3xl md:text-4xl lg:text-5xl font-bold text-[#0f172a] mb-6">
            Simple, transparent pricing
          </h3>
          <p className="text-lg text-[#475569]">
            Pay only for successful verifications. No setup fees, no hidden costs.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl mx-auto mb-12">
          {tiers.map((tier, index) => (
            <div
              key={index}
              className={`rounded-2xl p-8 ${
                tier.highlighted
                  ? "bg-[#0f172a] text-white ring-2 ring-[#3b82f6] scale-105"
                  : "bg-[#f8fafc] border border-[#e2e8f0]"
              }`}
            >
              {tier.highlighted && (
                <span className="inline-block bg-[#3b82f6] text-white text-xs font-semibold px-3 py-1 rounded-full mb-4">
                  Most Popular
                </span>
              )}
              <h4
                className={`text-xl font-bold mb-2 ${
                  tier.highlighted ? "text-white" : "text-[#0f172a]"
                }`}
              >
                {tier.name}
              </h4>
              <div className="mb-4">
                <span
                  className={`text-3xl font-bold ${
                    tier.highlighted ? "text-white" : "text-[#0f172a]"
                  }`}
                >
                  {tier.price}
                </span>
                <span
                  className={`text-sm ml-1 ${
                    tier.highlighted ? "text-white/70" : "text-[#64748b]"
                  }`}
                >
                  {tier.unit}
                </span>
              </div>
              <p
                className={`text-sm mb-6 ${
                  tier.highlighted ? "text-white/70" : "text-[#64748b]"
                }`}
              >
                {tier.description}
              </p>
              <ul className="space-y-3 mb-8">
                {tier.features.map((feature, i) => (
                  <li key={i} className="flex items-start gap-2">
                    <svg
                      className={`w-5 h-5 flex-shrink-0 mt-0.5 ${
                        tier.highlighted ? "text-[#3b82f6]" : "text-green-500"
                      }`}
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M5 13l4 4L19 7"
                      />
                    </svg>
                    <span
                      className={`text-sm ${
                        tier.highlighted ? "text-white/80" : "text-[#475569]"
                      }`}
                    >
                      {feature}
                    </span>
                  </li>
                ))}
              </ul>
              <button
                onClick={scrollToContact}
                className={`w-full py-3 px-6 rounded-lg font-semibold transition-colors ${
                  tier.highlighted
                    ? "bg-[#3b82f6] text-white hover:bg-[#2563eb]"
                    : "bg-[#0f172a] text-white hover:bg-[#1e293b]"
                }`}
              >
                {tier.cta}
              </button>
            </div>
          ))}
        </div>

        <p className="text-center text-[#64748b] text-sm">
          All plans include document verification for 190+ countries. Volume discounts available.
        </p>
      </div>
    </section>
  );
};

export default Pricing;
