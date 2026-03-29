"use client";

const Hero = () => {
  const scrollToContact = () => {
    const element = document.getElementById("contact");
    if (element) {
      element.scrollIntoView({ behavior: "smooth" });
    }
  };

  return (
    <section className="relative min-h-screen flex items-center gradient-bg overflow-hidden pt-16">
      {/* Background decoration */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute -top-40 -right-40 w-96 h-96 bg-[#3b82f6]/10 rounded-full blur-3xl" />
        <div className="absolute -bottom-40 -left-40 w-96 h-96 bg-[#3b82f6]/10 rounded-full blur-3xl" />
      </div>

      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24">
        <div className="max-w-3xl mx-auto text-center md:text-left md:mx-0">
          <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-bold text-white mb-6 leading-tight">
            Automate any workflow. Deploy AI agents in minutes.
          </h1>
          <p className="text-xl md:text-2xl text-white/70 mb-10 max-w-2xl">
            Build, deploy, and manage AI agents that work 24/7. No code required.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 mb-4 justify-center md:justify-start">
            <a href="https://app.passflow.ai/login?mode=register" className="btn-primary text-lg">
              Start Free
            </a>
            <button onClick={scrollToContact} className="btn-secondary text-lg">
              Book a Demo
            </button>
          </div>
          <p className="text-white/50 text-sm mb-10">
            Free plan available. No credit card required.
          </p>

          {/* Social proof */}
          <div className="pt-8 border-t border-white/10">
            <p className="text-white/60 text-lg md:text-xl font-medium">
              Trusted by ops teams at fast-growing companies
            </p>
          </div>
        </div>
      </div>
    </section>
  );
};

export default Hero;
