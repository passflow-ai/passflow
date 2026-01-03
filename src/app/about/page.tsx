import Header from "@/components/Header";
import Footer from "@/components/Footer";

export default function About() {
  return (
    <>
      <Header />
      <main className="min-h-screen bg-[#0f172a] pt-24">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
          <h1 className="text-4xl font-bold text-white mb-4">About Passflow</h1>
          <p className="text-xl text-white/70 mb-12">
            Building the future of identity verification for growth-focused teams.
          </p>

          <section className="mb-12">
            <h2 className="text-2xl font-semibold text-white mb-4">Our Mission</h2>
            <p className="text-white/70 text-lg">
              We believe identity verification should accelerate growth, not slow it down. Passflow was built to help fintechs, marketplaces, and digital platforms verify users in real-time without sacrificing conversion rates or creating friction in the onboarding experience.
            </p>
          </section>

          <section className="mb-12">
            <h2 className="text-2xl font-semibold text-white mb-4">What We Do</h2>
            <p className="text-white/70 text-lg mb-4">
              Passflow provides AI-powered identity verification that combines document verification, biometric matching, and liveness detection into a seamless experience. Our technology processes verifications in under 10 seconds while maintaining the highest accuracy standards.
            </p>
            <ul className="list-disc list-inside text-white/70 space-y-2">
              <li>Real-time document verification across 190+ countries</li>
              <li>iBeta Level 2 certified liveness detection</li>
              <li>Biometric matching with 99.9% accuracy</li>
              <li>Flexible deployment options (cloud, on-premise, hybrid)</li>
            </ul>
          </section>

          <section className="mb-12">
            <h2 className="text-2xl font-semibold text-white mb-4">Why Teams Choose Us</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="bg-white/5 border border-white/10 rounded-xl p-6">
                <h3 className="text-lg font-semibold text-white mb-2">Speed to Market</h3>
                <p className="text-white/60">
                  Get up and running in days, not months. Our APIs are designed for quick integration with minimal engineering overhead.
                </p>
              </div>
              <div className="bg-white/5 border border-white/10 rounded-xl p-6">
                <h3 className="text-lg font-semibold text-white mb-2">Conversion Focused</h3>
                <p className="text-white/60">
                  We optimize for user experience without compromising security. Higher conversion rates mean more revenue.
                </p>
              </div>
              <div className="bg-white/5 border border-white/10 rounded-xl p-6">
                <h3 className="text-lg font-semibold text-white mb-2">Enterprise Ready</h3>
                <p className="text-white/60">
                  ISO 27001, ISO 9001, iBeta, and NIST compliant. Built for teams with serious security requirements.
                </p>
              </div>
              <div className="bg-white/5 border border-white/10 rounded-xl p-6">
                <h3 className="text-lg font-semibold text-white mb-2">Global Coverage</h3>
                <p className="text-white/60">
                  Verify identity documents from over 190 countries with data residency options in the US and EU.
                </p>
              </div>
            </div>
          </section>

          <section className="bg-white/5 border border-white/10 rounded-xl p-8">
            <h2 className="text-2xl font-semibold text-white mb-4">Get in Touch</h2>
            <p className="text-white/70 mb-6">
              Ready to learn more about how Passflow can help your team grow? We&apos;d love to hear from you.
            </p>
            <a
              href="mailto:hello@passflow.ai"
              className="inline-block bg-[#3b82f6] text-white px-6 py-3 rounded-lg font-semibold hover:bg-[#2563eb] transition-colors"
            >
              Contact Us
            </a>
          </section>
        </div>
      </main>
      <Footer />
    </>
  );
}
