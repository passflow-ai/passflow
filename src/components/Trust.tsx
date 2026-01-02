const Trust = () => {
  const features = [
    "High-volume onboarding",
    "Real-time decisions",
    "Flexible deployment with US & EU data options",
  ];

  return (
    <section className="section-padding bg-[#0f172a]">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-3xl mx-auto text-center">
          <h3 className="text-3xl md:text-4xl lg:text-5xl font-bold text-white mb-10">
            Built for teams that operate at scale
          </h3>

          <ul className="space-y-4 max-w-md mx-auto mb-10">
            {features.map((feature, index) => (
              <li key={index} className="flex items-center gap-3 justify-center">
                <span className="flex-shrink-0 w-6 h-6 bg-[#3b82f6]/20 rounded-full flex items-center justify-center">
                  <svg
                    className="w-4 h-4 text-[#3b82f6]"
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
                <span className="text-white font-medium">{feature}</span>
              </li>
            ))}
          </ul>

          <p className="text-white/60">
            Compliance and security details available on request.
          </p>
        </div>
      </div>
    </section>
  );
};

export default Trust;
