"use client";

import Link from "next/link";
import { useTranslations } from "next-intl";

const Footer = () => {
  const t = useTranslations("Footer");

  return (
    <footer style={{ background: 'var(--bg-primary)', borderTop: '1px solid var(--border-primary)' }}>
      <div className="container py-16">
        {/* Main Footer Content */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-6 gap-12 mb-12">
          {/* Logo and Tagline */}
          <div className="lg:col-span-2">
            <Link href="/" className="inline-block mb-4">
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xl)', fontWeight: 'var(--weight-bold)', color: 'var(--text-primary)' }}>
                Passflow<span style={{ color: 'var(--accent-primary)' }}>.ai</span>
              </span>
            </Link>
            <p className="body-text-sm mb-6" style={{ maxWidth: '280px' }}>
              {t("tagline")}
            </p>

            {/* Newsletter */}
            <div className="mb-6">
              <p className="label mb-3" style={{ color: 'var(--text-tertiary)' }}>
                {t("newsletter.title")}
              </p>
              <form className="flex gap-2" action="/api/newsletter" method="POST">
                <input
                  type="email"
                  name="email"
                  placeholder={t("newsletter.placeholder")}
                  className="flex-1 px-4 py-2 rounded-md"
                  style={{
                    background: 'var(--bg-surface)',
                    border: '1px solid var(--border-primary)',
                    color: 'var(--text-primary)',
                    fontFamily: 'var(--font-mono)',
                    fontSize: 'var(--text-sm)'
                  }}
                />
                <button
                  type="submit"
                  className="btn-primary"
                  style={{ padding: '8px 16px', fontSize: 'var(--text-xs)' }}
                >
                  {t("newsletter.button")}
                </button>
              </form>
            </div>

            {/* Social Links */}
            <div className="flex items-center gap-4">
              <a
                href="https://github.com/passflow-ai"
                target="_blank"
                rel="noopener noreferrer"
                aria-label="GitHub"
                className="social-link"
              >
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                  <path fillRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clipRule="evenodd" />
                </svg>
              </a>
              <a
                href="https://linkedin.com/company/passflow-ai"
                target="_blank"
                rel="noopener noreferrer"
                aria-label="LinkedIn"
                className="social-link"
              >
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/>
                </svg>
              </a>
              <a
                href="https://x.com/passflow_ai"
                target="_blank"
                rel="noopener noreferrer"
                aria-label="X (Twitter)"
                className="social-link"
              >
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
                </svg>
              </a>
            </div>
          </div>

          {/* Producto */}
          <div>
            <h4 className="label-uppercase mb-4" style={{ color: 'var(--text-primary)' }}>{t("columns.product.title")}</h4>
            <ul className="space-y-3">
              <li>
                <Link href="/pricing" className="body-text-sm transition-colors hover:text-[var(--accent-primary)]">
                  {t("columns.product.links.pricing")}
                </Link>
              </li>
              <li>
                <Link href="/changelog" className="body-text-sm transition-colors hover:text-[var(--accent-primary)]">
                  {t("columns.product.links.changelog")}
                </Link>
              </li>
              <li>
                <a href="https://status.passflow.ai" target="_blank" rel="noopener noreferrer" className="body-text-sm transition-colors hover:text-[var(--accent-primary)]">
                  {t("columns.product.links.status")}
                </a>
              </li>
            </ul>
          </div>

          {/* Recursos */}
          <div>
            <h4 className="label-uppercase mb-4" style={{ color: 'var(--text-primary)' }}>{t("columns.resources.title")}</h4>
            <ul className="space-y-3">
              <li>
                <Link href="/docs" className="body-text-sm transition-colors hover:text-[var(--accent-primary)]">
                  {t("columns.resources.links.docs")}
                </Link>
              </li>
              <li>
                <Link href="/blog" className="body-text-sm transition-colors hover:text-[var(--accent-primary)]">
                  {t("columns.resources.links.blog")}
                </Link>
              </li>
              <li>
                <Link href="/docs/api" className="body-text-sm transition-colors hover:text-[var(--accent-primary)]">
                  {t("columns.resources.links.apiReference")}
                </Link>
              </li>
            </ul>
          </div>

          {/* Legal */}
          <div>
            <h4 className="label-uppercase mb-4" style={{ color: 'var(--text-primary)' }}>{t("columns.legal.title")}</h4>
            <ul className="space-y-3">
              <li>
                <Link href="/privacy" className="body-text-sm transition-colors hover:text-[var(--accent-primary)]">
                  {t("columns.legal.links.privacy")}
                </Link>
              </li>
              <li>
                <Link href="/terms" className="body-text-sm transition-colors hover:text-[var(--accent-primary)]">
                  {t("columns.legal.links.terms")}
                </Link>
              </li>
              <li>
                <Link href="/security" className="body-text-sm transition-colors hover:text-[var(--accent-primary)]">
                  {t("columns.legal.links.security")}
                </Link>
              </li>
            </ul>
          </div>

          {/* Empresa */}
          <div>
            <h4 className="label-uppercase mb-4" style={{ color: 'var(--text-primary)' }}>{t("columns.company.title")}</h4>
            <ul className="space-y-3">
              <li>
                <Link href="/about" className="body-text-sm transition-colors hover:text-[var(--accent-primary)]">
                  {t("columns.company.links.about")}
                </Link>
              </li>
              <li>
                <Link href="/contact" className="body-text-sm transition-colors hover:text-[var(--accent-primary)]">
                  {t("columns.company.links.contact")}
                </Link>
              </li>
              <li>
                <Link href="/careers" className="body-text-sm transition-colors hover:text-[var(--accent-primary)]">
                  {t("columns.company.links.careers")}
                </Link>
              </li>
            </ul>
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="pt-8" style={{ borderTop: '1px solid var(--border-primary)' }}>
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <p className="label" style={{ color: 'var(--text-tertiary)' }}>
              &copy; {t("copyright")}
            </p>

            {/* Trust Badges */}
            <div className="flex flex-wrap items-center justify-center gap-4">
              <div className="flex items-center gap-2" style={{ color: 'var(--text-tertiary)' }}>
                <svg className="w-4 h-4" style={{ color: 'var(--accent-primary)' }} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
                <span className="label-uppercase" style={{ fontSize: 'var(--text-xs)' }}>{t("trustBadges.aes256")}</span>
              </div>
              <div className="flex items-center gap-2" style={{ color: 'var(--text-tertiary)' }}>
                <svg className="w-4 h-4" style={{ color: 'var(--accent-primary)' }} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </svg>
                <span className="label-uppercase" style={{ fontSize: 'var(--text-xs)' }}>{t("trustBadges.soc2")}</span>
              </div>
              <div className="flex items-center gap-2" style={{ color: 'var(--text-tertiary)' }}>
                <svg className="w-4 h-4" style={{ color: 'var(--accent-primary)' }} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span className="label-uppercase" style={{ fontSize: 'var(--text-xs)' }}>{t("trustBadges.dataResidency")}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
