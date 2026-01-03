import Header from "@/components/Header";
import Footer from "@/components/Footer";

export default function Security() {
  const certifications = [
    {
      title: "ISO 27001",
      description: "Information Security Management System certification demonstrating our commitment to protecting sensitive data.",
    },
    {
      title: "ISO 9001",
      description: "Quality Management System certification ensuring consistent, high-quality service delivery.",
    },
    {
      title: "iBeta Level 2",
      description: "Biometric Presentation Attack Detection (PAD) testing certification, validating our liveness detection capabilities.",
    },
    {
      title: "NIST Compliant",
      description: "Adherence to National Institute of Standards and Technology guidelines for identity verification.",
    },
  ];

  const securityFeatures = [
    {
      title: "End-to-End Encryption",
      description: "All data is encrypted in transit using TLS 1.3 and at rest using AES-256 encryption.",
    },
    {
      title: "US & EU Data Residency",
      description: "Choose where your data is processed and stored to meet regional compliance requirements.",
    },
    {
      title: "Zero Data Retention Option",
      description: "Process verifications without storing biometric data beyond the session.",
    },
    {
      title: "Regular Penetration Testing",
      description: "Third-party security assessments conducted quarterly to identify and address vulnerabilities.",
    },
    {
      title: "SOC 2 Type II Ready",
      description: "Security controls designed to meet SOC 2 requirements for enterprise customers.",
    },
    {
      title: "GDPR & CCPA Compliant",
      description: "Full compliance with major privacy regulations including GDPR and CCPA.",
    },
  ];

  return (
    <>
      <Header />
      <main className="min-h-screen bg-[#0f172a] pt-24">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
          <h1 className="text-4xl font-bold text-white mb-4">Security</h1>
          <p className="text-xl text-white/70 mb-12">
            Enterprise-grade security built into every layer of our platform.
          </p>

          <section className="mb-16">
            <h2 className="text-2xl font-semibold text-white mb-6">Certifications</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {certifications.map((cert, index) => (
                <div
                  key={index}
                  className="bg-white/5 border border-white/10 rounded-xl p-6"
                >
                  <div className="flex items-center gap-3 mb-3">
                    <div className="w-10 h-10 bg-[#3b82f6]/20 rounded-lg flex items-center justify-center">
                      <svg
                        className="w-5 h-5 text-[#3b82f6]"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
                        />
                      </svg>
                    </div>
                    <h3 className="text-lg font-semibold text-white">{cert.title}</h3>
                  </div>
                  <p className="text-white/60">{cert.description}</p>
                </div>
              ))}
            </div>
          </section>

          <section className="mb-16">
            <h2 className="text-2xl font-semibold text-white mb-6">Security Features</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {securityFeatures.map((feature, index) => (
                <div
                  key={index}
                  className="bg-white/5 border border-white/10 rounded-xl p-6"
                >
                  <h3 className="text-lg font-semibold text-white mb-2">{feature.title}</h3>
                  <p className="text-white/60">{feature.description}</p>
                </div>
              ))}
            </div>
          </section>

          <section className="bg-white/5 border border-white/10 rounded-xl p-8">
            <h2 className="text-2xl font-semibold text-white mb-4">Request Security Documentation</h2>
            <p className="text-white/70 mb-6">
              Need detailed security documentation for your compliance team? We provide comprehensive security assessments, penetration test reports, and compliance documentation upon request.
            </p>
            <a
              href="mailto:security@passflow.ai"
              className="inline-block bg-[#3b82f6] text-white px-6 py-3 rounded-lg font-semibold hover:bg-[#2563eb] transition-colors"
            >
              Contact Security Team
            </a>
          </section>
        </div>
      </main>
      <Footer />
    </>
  );
}
