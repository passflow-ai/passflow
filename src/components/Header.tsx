"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";

const Header = () => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const pathname = usePathname();
  const isHomePage = pathname === "/";

  const scrollToSection = (id: string) => {
    if (isHomePage) {
      const element = document.getElementById(id);
      if (element) {
        element.scrollIntoView({ behavior: "smooth" });
      }
    } else {
      window.location.href = `/#${id}`;
    }
    setIsMenuOpen(false);
  };

  const navLinks = [
    { label: "Product", id: "how-it-works" },
    { label: "Pricing", id: "pricing" },
    { label: "Security", id: "trust" },
  ];

  return (
    <header className="fixed top-0 left-0 right-0 z-50 bg-[#0f172a]/95 backdrop-blur-sm border-b border-white/10">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          <Link href="/" className="text-2xl font-bold text-white">
            Passflow<span className="text-[#3b82f6]">.ai</span>
          </Link>

          {/* Desktop nav */}
          <nav className="hidden md:flex items-center gap-8">
            {navLinks.map((link) => (
              <button
                key={link.id}
                onClick={() => scrollToSection(link.id)}
                className="text-white/70 hover:text-white transition-colors text-sm"
              >
                {link.label}
              </button>
            ))}
          </nav>

          {/* Desktop CTA */}
          <div className="hidden md:block">
            <a
              href="https://app.passflow.ai/login?mode=register"
              className="bg-[#3b82f6] text-white px-6 py-2.5 rounded-lg font-semibold hover:bg-[#2563eb] transition-colors"
            >
              Start Free
            </a>
          </div>

          {/* Mobile menu button */}
          <button
            onClick={() => setIsMenuOpen(!isMenuOpen)}
            className="md:hidden text-white p-2"
            aria-label="Toggle menu"
          >
            <svg
              className="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              {isMenuOpen ? (
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              ) : (
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 6h16M4 12h16M4 18h16"
                />
              )}
            </svg>
          </button>
        </div>

        {/* Mobile menu */}
        {isMenuOpen && (
          <div className="md:hidden py-4 border-t border-white/10">
            <div className="flex flex-col gap-4">
              {navLinks.map((link) => (
                <button
                  key={link.id}
                  onClick={() => scrollToSection(link.id)}
                  className="text-white/70 hover:text-white transition-colors text-left"
                >
                  {link.label}
                </button>
              ))}
              <a
                href="https://app.passflow.ai/login?mode=register"
                className="bg-[#3b82f6] text-white px-6 py-2.5 rounded-lg font-semibold hover:bg-[#2563eb] transition-colors w-full mt-2 text-center block"
              >
                Start Free
              </a>
            </div>
          </div>
        )}
      </div>
    </header>
  );
};

export default Header;
