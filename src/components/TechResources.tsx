"use client";

const TechResources = () => {
  const scrollToContact = () => {
    const element = document.getElementById("contact");
    if (element) {
      element.scrollIntoView({ behavior: "smooth" });
    }
  };

  const resources = [
    {
      icon: (
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
        </svg>
      ),
      title: "API Documentation",
      description: "Comprehensive guides and endpoint references",
    },
    {
      icon: (
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
        </svg>
      ),
      title: "Sandbox Environment",
      description: "Test the full verification flow before going live",
    },
    {
      icon: (
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
      ),
      title: "SDKs & Code Samples",
      description: "Ready-to-use examples in Node, Python, and Go",
    },
    {
      icon: (
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192l-3.536 3.536M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
      ),
      title: "Integration Support",
      description: "Dedicated help from our solutions engineers",
    },
  ];

  return (
    <section className="section-padding bg-[#f8fafc]">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-3xl mx-auto text-center mb-12">
          <h3 className="text-3xl md:text-4xl lg:text-5xl font-bold text-[#0f172a] mb-6">
            Technical Resources & Demo
          </h3>
          <p className="text-lg text-[#475569]">
            Get everything you need to evaluate and integrate Passflow.
          </p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 max-w-5xl mx-auto mb-12">
          {resources.map((resource, index) => (
            <div
              key={index}
              className="bg-white rounded-xl p-6 shadow-sm border border-[#e2e8f0]"
            >
              <div className="w-12 h-12 bg-[#3b82f6]/10 rounded-lg flex items-center justify-center text-[#3b82f6] mb-4">
                {resource.icon}
              </div>
              <h4 className="text-lg font-semibold text-[#0f172a] mb-2">
                {resource.title}
              </h4>
              <p className="text-sm text-[#64748b]">{resource.description}</p>
            </div>
          ))}
        </div>

        <div className="max-w-2xl mx-auto text-center bg-white rounded-2xl p-8 shadow-sm border border-[#e2e8f0]">
          <p className="text-[#475569] mb-6">
            After an initial conversation, you&apos;ll receive access to our complete
            technical documentation, sandbox environment, and integration support.
          </p>
          <button
            onClick={scrollToContact}
            className="bg-[#3b82f6] text-white px-8 py-3 rounded-lg font-semibold hover:bg-[#2563eb] transition-colors"
          >
            Request Demo Access
          </button>
        </div>
      </div>
    </section>
  );
};

export default TechResources;
