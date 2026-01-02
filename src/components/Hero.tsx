"use client";

const Hero = () => {
  const scrollToContact = () => {
    const element = document.getElementById("contact");
    if (element) {
      element.scrollIntoView({ behavior: "smooth" });
    }
  };

  const scrollToHowItWorks = () => {
    const element = document.getElementById("how-it-works");
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
            Turn onboarding into revenue. Not fraud.
          </h1>
          <p className="text-xl md:text-2xl text-white/70 mb-10 max-w-2xl">
            Real-time identity verification built for growth teams.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 mb-4 justify-center md:justify-start">
            <button onClick={scrollToHowItWorks} className="btn-primary text-lg">
              See how it works
            </button>
            <button onClick={scrollToContact} className="btn-secondary text-lg">
              Talk to our team
            </button>
          </div>
          <p className="text-white/50 text-sm mb-8">
            No long sales cycles. No heavy integrations.
          </p>

          {/* Social proof */}
          <div className="pt-8 border-t border-white/10">
            <p className="text-white/40 text-sm mb-4">
              Trusted by growth teams processing $50M+ monthly
            </p>
            <div className="flex items-center gap-6 justify-center md:justify-start opacity-60">
              {/* Placeholder logos - grayscale */}
              <div className="h-6 w-20 bg-white/20 rounded" />
              <div className="h-6 w-16 bg-white/20 rounded" />
              <div className="h-6 w-24 bg-white/20 rounded" />
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default Hero;
