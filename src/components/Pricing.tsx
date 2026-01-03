"use client";

const Pricing = () => {
  const scrollToContact = () => {
    const element = document.getElementById("contact");
    if (element) {
      element.scrollIntoView({ behavior: "smooth" });
    }
  };

  const factors = [
    {
      icon: (
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
        </svg>
      ),
      title: "Monthly volume",
      description: "Based on your verification needs",
    },
    {
      icon: (
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
      ),
      title: "Deployment",
      description: "Cloud, on-premise, or hybrid",
    },
    {
      icon: (
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192l-3.536 3.536M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
      ),
      title: "Support level",
      description: "Standard to dedicated 24/7",
    },
  ];

  return (
    <section id="pricing" className="section-padding bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-3xl mx-auto text-center mb-12">
          <h3 className="text-3xl md:text-4xl lg:text-5xl font-bold text-[#0f172a] mb-6">
            Pricing
          </h3>
          <p className="text-lg text-[#475569] mb-4">
            Passflow uses a usage-based pricing model designed to scale with your onboarding volume and operational needs.
          </p>
          <p className="text-[#64748b]">
            Our team will help you define the right plan based on your use case and growth stage.
          </p>
        </div>

        <div className="max-w-3xl mx-auto mb-12">
          <p className="text-center text-[#475569] mb-8">
            Pricing depends on:
          </p>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {factors.map((factor, index) => (
              <div
                key={index}
                className="bg-[#f8fafc] rounded-xl p-6 text-center border border-[#e2e8f0]"
              >
                <div className="w-12 h-12 bg-[#3b82f6]/10 rounded-lg flex items-center justify-center text-[#3b82f6] mx-auto mb-4">
                  {factor.icon}
                </div>
                <h4 className="text-[#0f172a] font-semibold mb-1">{factor.title}</h4>
                <p className="text-sm text-[#64748b]">{factor.description}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <button
            onClick={scrollToContact}
            className="bg-[#3b82f6] text-white px-8 py-3 rounded-lg font-semibold hover:bg-[#2563eb] transition-colors"
          >
            Talk to our team
          </button>
          <button
            onClick={scrollToContact}
            className="border-2 border-[#0f172a] text-[#0f172a] px-8 py-3 rounded-lg font-semibold hover:bg-[#0f172a] hover:text-white transition-colors"
          >
            Request pricing
          </button>
        </div>
      </div>
    </section>
  );
};

export default Pricing;
