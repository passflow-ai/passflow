import { Metadata } from "next";
import Header from "@/components/Header";
import Footer from "@/components/Footer";

export const metadata: Metadata = {
  title: "Privacy Policy",
  description: "Passflow.ai Privacy Policy - Learn how we collect, use, and protect your data. ISO 27001 certified identity verification.",
  openGraph: {
    title: "Privacy Policy - Passflow.ai",
    description: "Learn how Passflow.ai protects your data and privacy.",
  },
  alternates: {
    canonical: "https://passflow.ai/privacy",
  },
};

export default function Privacy() {
  return (
    <>
      <Header />
      <main className="min-h-screen bg-[#0f172a] pt-24">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
          <h1 className="text-4xl font-bold text-white mb-8">Privacy Policy</h1>

          <div className="prose prose-invert prose-lg max-w-none">
            <p className="text-white/70 mb-6">
              Last updated: January 2025
            </p>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-white mb-4">1. Introduction</h2>
              <p className="text-white/70">
                Passflow.ai (&quot;we&quot;, &quot;our&quot;, or &quot;us&quot;) is committed to protecting your privacy. This Privacy Policy explains how we collect, use, disclose, and safeguard your information when you use our identity verification services.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-white mb-4">2. Information We Collect</h2>
              <p className="text-white/70 mb-4">
                We collect information that you provide directly to us, including:
              </p>
              <ul className="list-disc list-inside text-white/70 space-y-2">
                <li>Identity documents and biometric data for verification purposes</li>
                <li>Contact information (name, email, company)</li>
                <li>Usage data and analytics</li>
              </ul>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-white mb-4">3. How We Use Your Information</h2>
              <p className="text-white/70 mb-4">
                We use the information we collect to:
              </p>
              <ul className="list-disc list-inside text-white/70 space-y-2">
                <li>Provide and maintain our identity verification services</li>
                <li>Process and complete verification requests</li>
                <li>Communicate with you about our services</li>
                <li>Comply with legal obligations</li>
              </ul>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-white mb-4">4. Data Retention</h2>
              <p className="text-white/70">
                We retain personal data only for as long as necessary to fulfill the purposes for which it was collected, or as required by law. Biometric data is processed in real-time and not stored beyond the verification session unless required by our clients for compliance purposes.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-white mb-4">5. Data Security</h2>
              <p className="text-white/70">
                We implement industry-standard security measures including end-to-end encryption, secure data centers, and regular security audits. We are ISO 27001 and ISO 9001 certified.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-white mb-4">6. Your Rights</h2>
              <p className="text-white/70">
                Depending on your location, you may have rights regarding your personal data, including the right to access, correct, delete, or port your data. Contact us at privacy@passflow.ai to exercise these rights.
              </p>
            </section>

            <section className="mb-8">
              <h2 className="text-2xl font-semibold text-white mb-4">7. Contact Us</h2>
              <p className="text-white/70">
                If you have questions about this Privacy Policy, please contact us at:{" "}
                <a href="mailto:privacy@passflow.ai" className="text-[#3b82f6] hover:underline">
                  privacy@passflow.ai
                </a>
              </p>
            </section>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
