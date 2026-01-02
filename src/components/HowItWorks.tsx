const HowItWorks = () => {
  const steps = [
    {
      number: "1",
      title: "User starts onboarding",
    },
    {
      number: "2",
      title: "Identity verified in real time",
    },
    {
      number: "3",
      title: "Real users pass. Fraud doesn't.",
    },
  ];

  return (
    <section id="how-it-works" className="section-padding bg-[#f8fafc]">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-3xl mx-auto text-center mb-12">
          <h3 className="text-3xl md:text-4xl lg:text-5xl font-bold text-[#0f172a]">
            How Passflow works
          </h3>
        </div>

        <div className="max-w-2xl mx-auto">
          <div className="space-y-6">
            {steps.map((step, index) => (
              <div
                key={index}
                className="flex items-center gap-6 p-6 bg-white rounded-xl shadow-sm border border-[#e2e8f0]"
              >
                <div className="flex-shrink-0 w-12 h-12 bg-[#3b82f6] rounded-full flex items-center justify-center text-white text-xl font-bold">
                  {step.number}
                </div>
                <p className="text-lg md:text-xl text-[#0f172a] font-medium">
                  {step.title}
                </p>
              </div>
            ))}
          </div>

          <p className="text-center text-lg text-[#475569] mt-10">
            No friction. No delays. No manual bottlenecks.
          </p>
        </div>
      </div>
    </section>
  );
};

export default HowItWorks;
