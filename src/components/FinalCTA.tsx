"use client";

import { useState } from "react";

const FinalCTA = () => {
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    company: "",
    industry: "",
    volume: "",
    timeline: "",
  });
  const [status, setStatus] = useState<"idle" | "loading" | "success" | "error">("idle");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setStatus("loading");

    try {
      const response = await fetch("/api/contact", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(formData),
      });

      if (response.ok) {
        setStatus("success");
        setFormData({ name: "", email: "", company: "", industry: "", volume: "", timeline: "" });
      } else {
        setStatus("error");
      }
    } catch {
      setStatus("error");
    }
  };

  return (
    <section id="contact" className="py-16 md:py-20 bg-gradient-to-br from-[#0f172a] to-[#1e293b]">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-2xl mx-auto text-center mb-12">
          <h2 className="text-3xl md:text-4xl lg:text-5xl font-bold text-white mb-6">
            Let&apos;s grow your revenue — not your fraud.
          </h2>
          <p className="text-white/60">
            Get a personalized demo and pricing quote for your business.
          </p>
        </div>

        <div className="max-w-lg mx-auto">
          {status === "success" ? (
            <div className="bg-green-500/10 border border-green-500/20 rounded-xl p-8 text-center">
              <svg
                className="w-12 h-12 text-green-400 mx-auto mb-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5 13l4 4L19 7"
                />
              </svg>
              <h3 className="text-xl font-semibold text-white mb-2">
                Thanks for reaching out!
              </h3>
              <p className="text-white/70">
                We&apos;ll get back to you within 24 hours with a personalized demo.
              </p>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <input
                    type="text"
                    placeholder="Name"
                    required
                    value={formData.name}
                    onChange={(e) =>
                      setFormData({ ...formData, name: e.target.value })
                    }
                    className="w-full px-4 py-3 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/50 focus:border-[#3b82f6] focus:ring-2 focus:ring-[#3b82f6]/20 outline-none transition-colors"
                  />
                </div>
                <div>
                  <input
                    type="email"
                    placeholder="Work email"
                    required
                    value={formData.email}
                    onChange={(e) =>
                      setFormData({ ...formData, email: e.target.value })
                    }
                    className="w-full px-4 py-3 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/50 focus:border-[#3b82f6] focus:ring-2 focus:ring-[#3b82f6]/20 outline-none transition-colors"
                  />
                </div>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <input
                    type="text"
                    placeholder="Company"
                    required
                    value={formData.company}
                    onChange={(e) =>
                      setFormData({ ...formData, company: e.target.value })
                    }
                    className="w-full px-4 py-3 rounded-lg bg-white/10 border border-white/20 text-white placeholder-white/50 focus:border-[#3b82f6] focus:ring-2 focus:ring-[#3b82f6]/20 outline-none transition-colors"
                  />
                </div>
                <div>
                  <select
                    value={formData.industry}
                    onChange={(e) =>
                      setFormData({ ...formData, industry: e.target.value })
                    }
                    className="w-full px-4 py-3 rounded-lg bg-white/10 border border-white/20 text-white focus:border-[#3b82f6] focus:ring-2 focus:ring-[#3b82f6]/20 outline-none transition-colors appearance-none cursor-pointer"
                  >
                    <option value="" className="bg-[#1e293b]">Industry</option>
                    <option value="fintech" className="bg-[#1e293b]">Fintech</option>
                    <option value="lending" className="bg-[#1e293b]">Lending</option>
                    <option value="marketplace" className="bg-[#1e293b]">Marketplace</option>
                    <option value="crypto" className="bg-[#1e293b]">Crypto / Web3</option>
                    <option value="gaming" className="bg-[#1e293b]">Gaming / iGaming</option>
                    <option value="insurance" className="bg-[#1e293b]">Insurance</option>
                    <option value="other" className="bg-[#1e293b]">Other</option>
                  </select>
                </div>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <select
                    value={formData.volume}
                    onChange={(e) =>
                      setFormData({ ...formData, volume: e.target.value })
                    }
                    className="w-full px-4 py-3 rounded-lg bg-white/10 border border-white/20 text-white focus:border-[#3b82f6] focus:ring-2 focus:ring-[#3b82f6]/20 outline-none transition-colors appearance-none cursor-pointer"
                  >
                    <option value="" className="bg-[#1e293b]">Monthly volume</option>
                    <option value="<10k" className="bg-[#1e293b]">&lt; 10K verifications</option>
                    <option value="10k-50k" className="bg-[#1e293b]">10K - 50K verifications</option>
                    <option value="50k-100k" className="bg-[#1e293b]">50K - 100K verifications</option>
                    <option value="100k+" className="bg-[#1e293b]">100K+ verifications</option>
                  </select>
                </div>
                <div>
                  <select
                    value={formData.timeline}
                    onChange={(e) =>
                      setFormData({ ...formData, timeline: e.target.value })
                    }
                    className="w-full px-4 py-3 rounded-lg bg-white/10 border border-white/20 text-white focus:border-[#3b82f6] focus:ring-2 focus:ring-[#3b82f6]/20 outline-none transition-colors appearance-none cursor-pointer"
                  >
                    <option value="" className="bg-[#1e293b]">Timeline</option>
                    <option value="asap" className="bg-[#1e293b]">ASAP</option>
                    <option value="1-3months" className="bg-[#1e293b]">1-3 months</option>
                    <option value="3-6months" className="bg-[#1e293b]">3-6 months</option>
                    <option value="exploring" className="bg-[#1e293b]">Just exploring</option>
                  </select>
                </div>
              </div>
              <button
                type="submit"
                disabled={status === "loading"}
                className="w-full bg-[#3b82f6] text-white py-3.5 px-6 rounded-lg font-semibold hover:bg-[#2563eb] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {status === "loading" ? "Sending..." : "Get a Demo"}
              </button>
              {status === "error" && (
                <p className="text-red-400 text-sm text-center">
                  Something went wrong. Please try again.
                </p>
              )}
            </form>
          )}

          <p className="text-center text-white/50 text-lg md:text-xl font-medium mt-8">
            Built for growth teams. Not paperwork.
          </p>
        </div>
      </div>
    </section>
  );
};

export default FinalCTA;
