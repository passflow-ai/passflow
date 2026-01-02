const HowItWorks = () => {
  const steps = [
    {
      number: "1",
      title: "User starts onboarding",
    },
    {
      number: "2",
      title: "Identity is verified in real time",
    },
    {
      number: "3",
      title: "Real users pass. Fraud doesn't.",
    },
  ];

  return (
    <section className="section-padding bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-3xl mx-auto text-center mb-12">
          <h2 className="text-3xl md:text-4xl lg:text-5xl font-bold text-[#0f172a] mb-6">
            How Passflow works
          </h2>
        </div>

        <div className="max-w-2xl mx-auto">
          <div className="space-y-6">
            {steps.map((step, index) => (
              <div
                key={index}
                className="flex items-center gap-6 p-6 bg-[#f8fafc] rounded-xl"
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

          <p className="text-center text-2xl font-semibold text-[#0f172a] mt-10">
            That&apos;s it.
          </p>
        </div>
      </div>
    </section>
  );
};

export default HowItWorks;
