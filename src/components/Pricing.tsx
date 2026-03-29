"use client";

const Pricing = () => {
  const scrollToContact = () => {
    const element = document.getElementById("contact");
    if (element) {
      element.scrollIntoView({ behavior: "smooth" });
    }
  };

  const handleSignUp = () => {
    window.location.href = 'https://app.passflow.ai/login?mode=register';
  };

  const tiers = [
    {
      name: "Free",
      price: "$0",
      unit: "/month",
      description: "For individuals getting started with AI agents",
      features: [
        "Up to 3 agents",
        "500K tokens/month",
        "1 workspace",
        "Basic templates",
        "Community support",
      ],
      cta: "Start Free",
      highlighted: false,
      action: handleSignUp,
    },
    {
      name: "Starter",
      price: "$99",
      unit: "/month",
      description: "For small teams automating key workflows",
      features: [
        "Up to 5 agents",
        "5M tokens/month",
        "1 workspace",
        "All templates",
        "BYOK (30% off)",
        "Email support",
        "Audit log",
      ],
      cta: "Get Started",
      highlighted: false,
      action: handleSignUp,
    },
    {
      name: "Growth",
      price: "$299",
      unit: "/month",
      description: "For scaling teams with multiple workflows",
      features: [
        "Up to 15 agents",
        "15M tokens/month",
        "3 workspaces",
        "All Starter features",
        "Advanced analytics",
        "Priority support",
        "Webhooks & custom flows",
      ],
      cta: "Get Started",
      highlighted: true,
      action: handleSignUp,
    },
    {
      name: "Business",
      price: "$799",
      unit: "/month",
      description: "For organizations with high-volume automation needs",
      features: [
        "Unlimited agents",
        "40M tokens/month",
        "10 workspaces",
        "All Growth features",
        "SSO/SAML",
        "SLA 99.99%",
        "Dedicated support",
        "Custom integrations",
      ],
      cta: "Contact Sales",
      highlighted: false,
      action: scrollToContact,
    },
  ];

  return (
    <section id="pricing" className="section-padding bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-3xl mx-auto text-center mb-12">
          <h2 className="text-3xl md:text-4xl lg:text-5xl font-bold text-[#0f172a] mb-6">
            Simple, transparent pricing
          </h2>
          <p className="text-lg text-[#475569]">
            Start free. Scale as you grow. No hidden costs.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8 max-w-6xl mx-auto mb-12">
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
                onClick={tier.action}
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
          All plans include multi-LLM support. BYOK available on paid plans for 30% off.
        </p>
      </div>
    </section>
  );
};

export default Pricing;
