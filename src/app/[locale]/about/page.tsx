import { Metadata } from "next";
import Header from "@/components/Header";
import Footer from "@/components/Footer";

export const metadata: Metadata = {
  title: "About Us",
  description: "Learn about PassFlow, Inc - the company behind Passflow.ai. We build AI agent orchestration for teams that want to automate workflows without writing code.",
  openGraph: {
    title: "About PassFlow, Inc",
    description: "Building the future of workflow automation with AI agents.",
  },
  alternates: {
    canonical: "https://passflow.ai/about",
  },
};

export default function About() {
  return (
    <>
      <Header />
      <main className="min-h-screen bg-[#0f172a] pt-24">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
          <h1 className="text-4xl font-bold text-white mb-4">About Passflow</h1>
          <p className="text-xl text-white/70 mb-12">
            Building the future of workflow automation with AI agents.
          </p>

          <section className="mb-12">
            <h2 className="text-2xl font-semibold text-white mb-4">Our Mission</h2>
            <p className="text-white/70 text-lg">
              We believe AI should handle the repetitive work so teams can focus on what matters. Passflow was built to help companies of any size deploy autonomous AI agents without writing code.
            </p>
          </section>

          <section className="mb-12">
            <h2 className="text-2xl font-semibold text-white mb-4">What We Do</h2>
            <p className="text-white/70 text-lg mb-4">
              Passflow provides a visual platform for building AI agents that connect to your tools, process data, and execute workflows autonomously. Our platform supports multiple LLM providers and integrates with 25+ services.
            </p>
            <ul className="list-disc list-inside text-white/70 space-y-2">
              <li>Visual agent builder with drag-and-drop workflows</li>
              <li>Multi-LLM support (OpenAI, Anthropic, Google, Azure, and more)</li>
              <li>25+ integrations with popular tools and services</li>
              <li>Flexible deployment options (cloud, self-host, hybrid)</li>
            </ul>
          </section>

          <section className="mb-12">
            <h2 className="text-2xl font-semibold text-white mb-4">Why Teams Choose Us</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="bg-white/5 border border-white/10 rounded-xl p-6">
                <h3 className="text-lg font-semibold text-white mb-2">Speed to Value</h3>
                <p className="text-white/60">
                  Deploy your first AI agent in minutes. Our visual builder and pre-built templates eliminate the need for custom development.
                </p>
              </div>
              <div className="bg-white/5 border border-white/10 rounded-xl p-6">
                <h3 className="text-lg font-semibold text-white mb-2">No Code Required</h3>
                <p className="text-white/60">
                  Build powerful automations without writing a single line of code. Designed for ops teams, founders, and non-technical users.
                </p>
              </div>
              <div className="bg-white/5 border border-white/10 rounded-xl p-6">
                <h3 className="text-lg font-semibold text-white mb-2">Enterprise Ready</h3>
                <p className="text-white/60">
                  SOC 2 ready, end-to-end encryption, role-based access control, and audit logging. Built for teams with serious security requirements.
                </p>
              </div>
              <div className="bg-white/5 border border-white/10 rounded-xl p-6">
                <h3 className="text-lg font-semibold text-white mb-2">Multi-LLM Flexibility</h3>
                <p className="text-white/60">
                  Choose the best model for each task. Bring your own keys for 30% off, or use our managed tokens with data residency options in the US and EU.
                </p>
              </div>
            </div>
          </section>

          <section className="bg-white/5 border border-white/10 rounded-xl p-8">
            <h2 className="text-2xl font-semibold text-white mb-4">Get in Touch</h2>
            <p className="text-white/70 mb-6">
              Ready to learn more about how Passflow can help your team automate? We&apos;d love to hear from you.
            </p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
              <div>
                <h3 className="text-white font-semibold mb-2">PassFlow, Inc</h3>
                <p className="text-white/60 text-sm">
                  3723 Greenville Ave STE 57075<br />
                  Dallas, TX 75206
                </p>
              </div>
              <div>
                <h3 className="text-white font-semibold mb-2">Contact</h3>
                <p className="text-white/60 text-sm">
                  <a href="tel:+14692493858" className="hover:text-white transition-colors">
                    (469) 249-3858
                  </a>
                  <br />
                  <a href="mailto:hello@passflow.ai" className="hover:text-white transition-colors">
                    hello@passflow.ai
                  </a>
                </p>
              </div>
            </div>
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
