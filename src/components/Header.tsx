"use client";

import { useState } from "react";
import { useTranslations, useLocale } from "next-intl";
import { Link, usePathname, useRouter } from "@/i18n/routing";

const Header = () => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const pathname = usePathname();
  const router = useRouter();
  const locale = useLocale();
  const t = useTranslations("Header");
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

  const switchLocale = (newLocale: "es" | "en") => {
    router.replace(pathname, { locale: newLocale });
  };

  const navLinks = [
    { label: "Product", id: "how-it-works" },
    { label: "Use Cases", id: "use-cases" },
    { label: "Architecture", id: "architecture" },
    { label: "Security", id: "security" },
    { label: "Docs", href: "/docs" },
  ];

  return (
    <header className="fixed top-0 left-0 right-0 z-[200] h-[52px]" style={{
      background: "rgba(255, 255, 255, 0.85)",
      backdropFilter: "blur(12px)",
      borderBottom: "1px solid var(--border-light)"
    }}>
      <div className="max-w-[var(--max-width-content)] mx-auto px-4 md:px-6 lg:px-8 h-full">
        <div className="flex items-center justify-between h-full">
          {/* Logo */}
          <Link
            href="/"
            className="flex items-center gap-2 hover:opacity-80 transition-opacity duration-[var(--duration-fast)]"
          >
            <div
              className="flex-shrink-0 rounded"
              style={{
                width: "8px",
                height: "8px",
                backgroundColor: "var(--accent-primary)"
              }}
            />
            <span
              className="font-semibold text-[var(--text-primary)]"
              style={{ fontSize: "0.95rem" }}
            >
              Passflow
            </span>
          </Link>

          {/* Desktop nav */}
          <nav className="hidden lg:flex items-center gap-8">
            {navLinks.map((link) => (
              link.href ? (
                <Link
                  key={link.href}
                  href={link.href}
                  className="text-[0.8rem] text-[var(--text-secondary)] hover:text-[var(--accent-primary)] transition-colors duration-[var(--duration-fast)]"
                >
                  {link.label}
                </Link>
              ) : (
                <button
                  key={link.id}
                  onClick={() => scrollToSection(link.id!)}
                  className="text-[0.8rem] text-[var(--text-secondary)] hover:text-[var(--accent-primary)] transition-colors duration-[var(--duration-fast)] cursor-pointer"
                >
                  {link.label}
                </button>
              )
            ))}
          </nav>

          {/* Desktop Actions */}
          <div className="hidden lg:flex items-center gap-4">
            {/* GitHub button */}
            <a
              href="https://github.com/passflow-ai/passflow"
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-2 px-3 py-2 text-[0.8rem] border border-[var(--border-light)] rounded-md text-[var(--text-primary)] hover:border-[var(--text-secondary)] hover:bg-[var(--bg-surface)] transition-all duration-[var(--duration-fast)]"
            >
              <svg
                className="w-4 h-4"
                fill="currentColor"
                viewBox="0 0 24 24"
              >
                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v-3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
              </svg>
              <span>GitHub</span>
            </a>

            {/* Language switcher */}
            <div className="flex items-center gap-1 pl-4 border-l border-[var(--border-light)]">
              <button
                onClick={() => switchLocale("es")}
                className={`text-xs px-2 py-1 rounded transition-all duration-[var(--duration-fast)] ${
                  locale === "es"
                    ? "text-[var(--accent-primary)] bg-[var(--accent-primary)]/10"
                    : "text-[var(--text-secondary)] hover:text-[var(--accent-primary)]"
                }`}
              >
                ES
              </button>
              <span className="text-[var(--text-muted)] text-xs">/</span>
              <button
                onClick={() => switchLocale("en")}
                className={`text-xs px-2 py-1 rounded transition-all duration-[var(--duration-fast)] ${
                  locale === "en"
                    ? "text-[var(--accent-primary)] bg-[var(--accent-primary)]/10"
                    : "text-[var(--text-secondary)] hover:text-[var(--accent-primary)]"
                }`}
              >
                EN
              </button>
            </div>

            {/* CTA Button */}
            <a
              href="https://app.passflow.ai/login?mode=register"
              className="btn-primary text-xs px-4 py-2"
            >
              Book demo
            </a>
          </div>

          {/* Mobile menu button */}
          <button
            onClick={() => setIsMenuOpen(!isMenuOpen)}
            className="lg:hidden text-[var(--text-primary)] p-2 hover:text-[var(--accent-primary)] transition-colors"
            aria-label={isMenuOpen ? t("mobileMenu.close") : t("mobileMenu.open")}
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
          <div
            className="lg:hidden absolute top-[52px] left-0 right-0 py-4 border-t border-[var(--border-light)]"
            style={{
              background: "rgba(255, 255, 255, 0.95)",
              backdropFilter: "blur(12px)"
            }}
          >
            <div className="max-w-[var(--max-width-content)] mx-auto px-4 md:px-6">
              <div className="flex flex-col gap-4">
                {navLinks.map((link) => (
                  link.href ? (
                    <Link
                      key={link.href}
                      href={link.href}
                      className="text-sm text-[var(--text-secondary)] hover:text-[var(--accent-primary)] transition-colors"
                      onClick={() => setIsMenuOpen(false)}
                    >
                      {link.label}
                    </Link>
                  ) : (
                    <button
                      key={link.id}
                      onClick={() => scrollToSection(link.id!)}
                      className="text-sm text-[var(--text-secondary)] hover:text-[var(--accent-primary)] transition-colors text-left cursor-pointer"
                    >
                      {link.label}
                    </button>
                  )
                ))}

                <a
                  href="https://github.com/passflow-ai/passflow"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-2 px-3 py-2 text-sm border border-[var(--border-light)] rounded-md text-[var(--text-primary)] hover:border-[var(--text-secondary)] hover:bg-[var(--bg-surface)] transition-all mt-2"
                >
                  <svg
                    className="w-4 h-4"
                    fill="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v-3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
                  </svg>
                  <span>GitHub</span>
                </a>

                <a
                  href="https://app.passflow.ai/login?mode=register"
                  className="btn-primary text-center mt-2"
                >
                  Book demo
                </a>

                {/* Mobile language switcher */}
                <div className="flex items-center justify-center gap-2 mt-4 pt-4 border-t border-[var(--border-light)]">
                  <button
                    onClick={() => switchLocale("es")}
                    className={`text-sm px-3 py-1 rounded transition-all duration-[var(--duration-fast)] ${
                      locale === "es"
                        ? "text-[var(--accent-primary)] bg-[var(--accent-primary)]/10"
                        : "text-[var(--text-secondary)] hover:text-[var(--accent-primary)]"
                    }`}
                  >
                    ES
                  </button>
                  <span className="text-[var(--text-muted)]">/</span>
                  <button
                    onClick={() => switchLocale("en")}
                    className={`text-sm px-3 py-1 rounded transition-all duration-[var(--duration-fast)] ${
                      locale === "en"
                        ? "text-[var(--accent-primary)] bg-[var(--accent-primary)]/10"
                        : "text-[var(--text-secondary)] hover:text-[var(--accent-primary)]"
                    }`}
                  >
                    EN
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </header>
  );
};

export default Header;
