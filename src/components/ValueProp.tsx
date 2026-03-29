const ValueProp = () => {
  const benefits = [
    "Less time on repetitive manual tasks",
    "Faster operations across your team",
    "Always-on automation that never sleeps",
  ];

  const metrics = [
    { value: "↓ 80%", label: "Manual work" },
    { value: "10x", label: "Faster execution" },
    { value: "24/7", label: "Uptime" },
  ];

  return (
    <section className="section-padding bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-3xl mx-auto text-center mb-12">
          <h2 className="text-3xl md:text-4xl lg:text-5xl font-bold text-[#0f172a] mb-10">
            What teams see after deploying Passflow
          </h2>

          <ul className="space-y-4 text-left max-w-md mx-auto mb-12">
            {benefits.map((benefit, index) => (
              <li key={index} className="flex items-center gap-3">
                <span className="flex-shrink-0 w-6 h-6 bg-green-100 rounded-full flex items-center justify-center">
                  <svg
                    className="w-4 h-4 text-green-600"
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
                </span>
                <span className="text-[#0f172a] text-lg font-medium">{benefit}</span>
              </li>
            ))}
          </ul>
        </div>

        {/* Metric cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-4xl mx-auto">
          {metrics.map((metric, index) => (
            <div
              key={index}
              className="bg-white rounded-xl p-8 text-center shadow-md border border-[#e2e8f0] hover:shadow-lg transition-shadow"
            >
              <div className="text-3xl md:text-4xl font-bold text-[#3b82f6] mb-2">
                {metric.value}
              </div>
              <div className="text-[#64748b] text-sm font-medium uppercase tracking-wide">
                {metric.label}
              </div>
            </div>
          ))}
        </div>

        <p className="text-center text-[#64748b] text-sm mt-8">
          Typical impact based on customer results
        </p>
      </div>
    </section>
  );
};

export default ValueProp;
