const HowItWorks = () => {
  const steps = [
    {
      number: "1",
      title: "Design your workflow visually",
    },
    {
      number: "2",
      title: "Connect your tools and LLMs",
    },
    {
      number: "3",
      title: "Deploy and let agents work 24/7",
    },
  ];

  return (
    <section id="how-it-works" className="section-padding bg-[#f8fafc]">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-3xl mx-auto text-center mb-16">
          <h2 className="text-3xl md:text-4xl lg:text-5xl font-bold text-[#0f172a]">
            How Passflow works
          </h2>
        </div>

        {/* Desktop stepper - horizontal */}
        <div className="hidden md:block max-w-4xl mx-auto">
          <div className="flex items-center justify-between relative">
            {/* Connector line */}
            <div className="absolute top-5 left-[10%] right-[10%] h-0.5 bg-[#e2e8f0]" />
            <div className="absolute top-5 left-[10%] w-[40%] h-0.5 bg-[#3b82f6]" />

            {steps.map((step, index) => (
              <div key={index} className="relative flex flex-col items-center w-1/3">
                <div className="w-10 h-10 bg-[#3b82f6] rounded-full flex items-center justify-center text-white text-lg font-bold z-10">
                  {step.number}
                </div>
                <p className="text-center text-[#0f172a] font-medium mt-4 px-4">
                  {step.title}
                </p>
              </div>
            ))}
          </div>
        </div>

        {/* Mobile stepper - vertical */}
        <div className="md:hidden max-w-sm mx-auto">
          <div className="relative">
            {/* Vertical connector line */}
            <div className="absolute left-5 top-5 bottom-5 w-0.5 bg-[#e2e8f0]" />

            <div className="space-y-8">
              {steps.map((step, index) => (
                <div key={index} className="relative flex items-start gap-4">
                  <div className="w-10 h-10 bg-[#3b82f6] rounded-full flex items-center justify-center text-white text-lg font-bold z-10 flex-shrink-0">
                    {step.number}
                  </div>
                  <p className="text-[#0f172a] font-medium pt-2">
                    {step.title}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </div>

        <p className="text-center text-lg text-[#475569] mt-12">
          No code. No complex integrations. No DevOps required.
        </p>
      </div>
    </section>
  );
};

export default HowItWorks;
