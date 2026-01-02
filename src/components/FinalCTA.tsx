"use client";

import { useState } from "react";

const FinalCTA = () => {
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    company: "",
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
        setFormData({ name: "", email: "", company: "" });
      } else {
        setStatus("error");
      }
    } catch {
      setStatus("error");
    }
  };

  return (
    <section id="contact" className="section-padding bg-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="max-w-2xl mx-auto text-center mb-12">
          <h2 className="text-3xl md:text-4xl lg:text-5xl font-bold text-[#0f172a] mb-6">
            Stop fraud. Increase conversion. Move faster.
          </h2>
        </div>

        <div className="max-w-md mx-auto">
          {status === "success" ? (
            <div className="bg-green-50 border border-green-200 rounded-xl p-8 text-center">
              <svg
                className="w-12 h-12 text-green-500 mx-auto mb-4"
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
              <h3 className="text-xl font-semibold text-[#0f172a] mb-2">
                Thanks for reaching out!
              </h3>
              <p className="text-[#475569]">
                We&apos;ll get back to you within 24 hours.
              </p>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <input
                  type="text"
                  placeholder="Name"
                  required
                  value={formData.name}
                  onChange={(e) =>
                    setFormData({ ...formData, name: e.target.value })
                  }
                  className="w-full px-4 py-3 rounded-lg border border-[#e2e8f0] focus:border-[#3b82f6] focus:ring-2 focus:ring-[#3b82f6]/20 outline-none transition-colors"
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
                  className="w-full px-4 py-3 rounded-lg border border-[#e2e8f0] focus:border-[#3b82f6] focus:ring-2 focus:ring-[#3b82f6]/20 outline-none transition-colors"
                />
              </div>
              <div>
                <input
                  type="text"
                  placeholder="Company"
                  required
                  value={formData.company}
                  onChange={(e) =>
                    setFormData({ ...formData, company: e.target.value })
                  }
                  className="w-full px-4 py-3 rounded-lg border border-[#e2e8f0] focus:border-[#3b82f6] focus:ring-2 focus:ring-[#3b82f6]/20 outline-none transition-colors"
                />
              </div>
              <button
                type="submit"
                disabled={status === "loading"}
                className="w-full btn-primary text-lg disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {status === "loading" ? "Sending..." : "Talk to Sales"}
              </button>
              {status === "error" && (
                <p className="text-red-500 text-sm text-center">
                  Something went wrong. Please try again.
                </p>
              )}
            </form>
          )}
        </div>
      </div>
    </section>
  );
};

export default FinalCTA;
