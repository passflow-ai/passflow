const ValueProp = () => {
  const benefits = [
    "Less fraud hitting your funnel",
    "More real users completing onboarding",
    "Fewer manual reviews slowing growth",
  ];

  return (
    <section className="section-padding bg-[#f8fafc]">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-3xl mx-auto text-center">
          <h3 className="text-3xl md:text-4xl lg:text-5xl font-bold text-[#0f172a] mb-10">
            What teams see after deploying Passflow
          </h3>

          <ul className="space-y-4 text-left max-w-md mx-auto mb-10">
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

          <p className="text-[#64748b] text-sm">
            Typical impact: ↓ fraud up to 60%, ↑ onboarding conversion 20–40%, ↓ manual reviews 40–70%
          </p>
        </div>
      </div>
    </section>
  );
};

export default ValueProp;
