const ValueProp = () => {
  const metrics = [
    { value: "60%", label: "reduction in fraud" },
    { value: "20-40%", label: "increase in onboarding conversion" },
    { value: "40-70%", label: "fewer manual reviews" },
  ];

  return (
    <section className="section-padding bg-[#f8fafc]">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-3xl mx-auto text-center mb-12">
          <h2 className="text-3xl md:text-4xl lg:text-5xl font-bold text-[#0f172a] mb-6">
            Built for revenue teams, not compliance theater
          </h2>
          <p className="text-lg text-[#475569]">
            Passflow verifies real users in real time while stopping fraud at the door.
          </p>
        </div>

        <div className="max-w-4xl mx-auto">
          <p className="text-center text-[#475569] mb-8">Our customers typically see:</p>
          <div className="grid md:grid-cols-3 gap-8 mb-12">
            {metrics.map((metric, index) => (
              <div
                key={index}
                className="bg-white rounded-xl p-8 text-center shadow-sm border border-[#e2e8f0]"
              >
                <div className="text-4xl md:text-5xl font-bold text-[#3b82f6] mb-2">
                  {metric.value}
                </div>
                <div className="text-[#475569]">{metric.label}</div>
              </div>
            ))}
          </div>

          <div className="text-center space-y-2">
            <p className="text-[#0f172a] font-medium">No extra steps.</p>
            <p className="text-[#0f172a] font-medium">No broken UX.</p>
            <p className="text-[#0f172a] font-medium">No delays.</p>
          </div>
        </div>
      </div>
    </section>
  );
};

export default ValueProp;
