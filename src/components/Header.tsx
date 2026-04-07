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
    { label: t("nav.product"), id: "how-it-works" },
    { label: t("nav.useCases"), id: "use-cases" },
    { label: t("nav.pricing"), id: "pricing" },
    { label: t("nav.docs"), href: "/docs" },
  ];

  return (
    <header className="fixed top-0 left-0 right-0 z-[200] bg-[var(--bg-overlay)] backdrop-blur-[12px] backdrop-saturate-[180%] border-b border-[var(--border-primary)]">
      <div className="max-w-[var(--max-width-content)] mx-auto px-4 md:px-6 lg:px-8">
        <div className="flex items-center justify-between h-14">
          <Link
            href="/"
            className="font-[var(--font-mono)] text-lg font-bold text-[var(--text-primary)] hover:text-[var(--accent-primary)] transition-colors duration-[var(--duration-fast)]"
          >
            Passflow<span className="text-[var(--accent-primary)]">.ai</span>
          </Link>

          {/* Desktop nav */}
          <nav className="hidden md:flex items-center gap-8">
            {navLinks.map((link) => (
              link.href ? (
                <Link
                  key={link.href}
                  href={link.href}
                  className="font-[var(--font-mono)] text-sm font-medium text-[var(--text-secondary)] hover:text-[var(--accent-primary)] transition-colors duration-[var(--duration-fast)]"
                >
                  {link.label}
                </Link>
              ) : (
                <button
                  key={link.id}
                  onClick={() => scrollToSection(link.id!)}
                  className="font-[var(--font-mono)] text-sm font-medium text-[var(--text-secondary)] hover:text-[var(--accent-primary)] transition-colors duration-[var(--duration-fast)]"
                >
                  {link.label}
                </button>
              )
            ))}
          </nav>

          {/* Desktop CTA */}
          <div className="hidden md:block">
            <a
              href="https://app.passflow.ai/login?mode=register"
              className="btn-primary text-xs px-5 py-2.5"
            >
              {t("cta")}
            </a>
          </div>

          {/* Language switcher */}
          <div className="hidden md:flex items-center gap-1 ml-4">
            <button
              onClick={() => switchLocale("es")}
              className={`font-[var(--font-mono)] text-xs px-2 py-1 rounded transition-colors ${
                locale === "es"
                  ? "text-[var(--accent-primary)] bg-[var(--accent-primary)]/10"
                  : "text-[var(--text-secondary)] hover:text-[var(--accent-primary)]"
              }`}
            >
              ES
            </button>
            <span className="text-[var(--text-tertiary)]">/</span>
            <button
              onClick={() => switchLocale("en")}
              className={`font-[var(--font-mono)] text-xs px-2 py-1 rounded transition-colors ${
                locale === "en"
                  ? "text-[var(--accent-primary)] bg-[var(--accent-primary)]/10"
                  : "text-[var(--text-secondary)] hover:text-[var(--accent-primary)]"
              }`}
            >
              EN
            </button>
          </div>

          {/* Mobile menu button */}
          <button
            onClick={() => setIsMenuOpen(!isMenuOpen)}
            className="md:hidden text-[var(--text-primary)] p-2 hover:text-[var(--accent-primary)] transition-colors"
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
          <div className="md:hidden py-4 border-t border-[var(--border-primary)]">
            <div className="flex flex-col gap-4">
              {navLinks.map((link) => (
                link.href ? (
                  <Link
                    key={link.href}
                    href={link.href}
                    className="font-[var(--font-mono)] text-sm text-[var(--text-secondary)] hover:text-[var(--accent-primary)] transition-colors"
                    onClick={() => setIsMenuOpen(false)}
                  >
                    {link.label}
                  </Link>
                ) : (
                  <button
                    key={link.id}
                    onClick={() => scrollToSection(link.id!)}
                    className="font-[var(--font-mono)] text-sm text-[var(--text-secondary)] hover:text-[var(--accent-primary)] transition-colors text-left"
                  >
                    {link.label}
                  </button>
                )
              ))}
              <a
                href="https://app.passflow.ai/login?mode=register"
                className="btn-primary text-center mt-2"
              >
                {t("cta")}
              </a>
              {/* Mobile language switcher */}
              <div className="flex items-center justify-center gap-2 mt-4 pt-4 border-t border-[var(--border-primary)]">
                <button
                  onClick={() => switchLocale("es")}
                  className={`font-[var(--font-mono)] text-sm px-3 py-1 rounded transition-colors ${
                    locale === "es"
                      ? "text-[var(--accent-primary)] bg-[var(--accent-primary)]/10"
                      : "text-[var(--text-secondary)] hover:text-[var(--accent-primary)]"
                  }`}
                >
                  ES
                </button>
                <span className="text-[var(--text-tertiary)]">/</span>
                <button
                  onClick={() => switchLocale("en")}
                  className={`font-[var(--font-mono)] text-sm px-3 py-1 rounded transition-colors ${
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
        )}
      </div>
    </header>
  );
};

export default Header;
