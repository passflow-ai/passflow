"use client";

import { useTranslations } from "next-intl";
import Header from "@/components/Header";
import Footer from "@/components/Footer";
import { useState, useEffect } from "react";

const TerminalDemo = () => {
  const [currentLine, setCurrentLine] = useState(0);
  const [showCursor, setShowCursor] = useState(true);

  const lines = [
    { type: "prompt", text: "$ passflow pal validate agent.yaml" },
    { type: "output", text: "✓ agent.yaml is valid" },
    { type: "prompt", text: "$ passflow pal apply agent.yaml --dry-run" },
    { type: "output", text: "→ Dry run - no changes applied" },
    { type: "output", text: "✓ Agent 'data-pipeline' would be created" },
    { type: "output", text: "" },
    { type: "output", text: "Changes:" },
    { type: "output", text: "  name: data-pipeline" },
    { type: "output", text: "  model: claude-3-5-sonnet-20241022" },
    { type: "output", text: "  tools: [web-search, calculator]" },
    { type: "prompt", text: "$ passflow pal apply agent.yaml" },
    { type: "output", text: "✓ Agent 'data-pipeline' created" },
    { type: "output", text: "" },
    { type: "output", text: "Agent ID: agent-7f3a9b2c" },
  ];

  useEffect(() => {
    const cursorInterval = setInterval(() => {
      setShowCursor((prev) => !prev);
    }, 500);

    const lineInterval = setInterval(() => {
      setCurrentLine((prev) => (prev < lines.length ? prev + 1 : prev));
    }, 800);

    return () => {
      clearInterval(cursorInterval);
      clearInterval(lineInterval);
    };
  }, [lines.length]);

  return (
    <div className="terminal-window">
      <div className="terminal-header">
        <div className="terminal-dots">
          <span className="dot red" />
          <span className="dot yellow" />
          <span className="dot green" />
        </div>
        <span className="terminal-title">passflow — zsh</span>
      </div>
      <div className="terminal-body">
        {lines.slice(0, currentLine).map((line, i) => (
          <div key={i} className={`terminal-line ${line.type}`}>
            {line.type === "prompt" ? (
              <span className="text-[var(--accent-primary)]">{line.text}</span>
            ) : (
              <span className="text-[var(--text-primary)]">{line.text}</span>
            )}
          </div>
        ))}
        {currentLine < lines.length && showCursor && (
          <span className="terminal-cursor">▋</span>
        )}
      </div>
    </div>
  );
};

const InstallCommand = () => {
  const [copied, setCopied] = useState(false);
  const command = "curl -fsSL https://raw.githubusercontent.com/jaak-ai/passflow-cli/main/scripts/install.sh | bash";

  const handleCopy = async () => {
    await navigator.clipboard.writeText(command);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="install-command" onClick={handleCopy}>
      <code>{command}</code>
      <button className="copy-btn" aria-label="Copy to clipboard">
        {copied ? (
          <svg className="w-5 h-5 text-[var(--accent-primary)]" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
          </svg>
        ) : (
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
        )}
      </button>
    </div>
  );
};

const CommandCard = ({ command, description, example }: { command: string; description: string; example: string }) => (
  <div className="command-card">
    <code className="command-name">{command}</code>
    <p className="command-description">{description}</p>
    <div className="command-example">
      <code>{example}</code>
    </div>
  </div>
);

export default function CLIPage() {
  const t = useTranslations("CLI");

  const commands = [
    {
      command: "passflow pal validate",
      description: t("commands.validate.description"),
      example: "passflow pal validate agent.yaml",
    },
    {
      command: "passflow pal apply",
      description: t("commands.apply.description"),
      example: "passflow pal apply agent.yaml --dry-run",
    },
    {
      command: "passflow pal export",
      description: t("commands.export.description"),
      example: "passflow pal export agent-123 -o agent.yaml",
    },
    {
      command: "passflow pal diff",
      description: t("commands.diff.description"),
      example: "passflow pal diff agent-123 agent.yaml",
    },
    {
      command: "passflow agents list",
      description: t("commands.list.description"),
      example: "passflow agents list -w workspace-id",
    },
    {
      command: "passflow agents get",
      description: t("commands.get.description"),
      example: "passflow agents get agent-123 --format json",
    },
  ];

  const features = [
    {
      icon: "⚡",
      title: t("features.fast.title"),
      description: t("features.fast.description"),
    },
    {
      icon: "🔐",
      title: t("features.secure.title"),
      description: t("features.secure.description"),
    },
    {
      icon: "📦",
      title: t("features.portable.title"),
      description: t("features.portable.description"),
    },
    {
      icon: "🔄",
      title: t("features.gitops.title"),
      description: t("features.gitops.description"),
    },
  ];

  return (
    <>
      <Header />
      <main>
        {/* Hero Section */}
        <section className="cli-hero">
          <div className="cli-hero-content">
            <div className="cli-hero-text">
              <div className="badge-pill">{t("badge")}</div>
              <h1 className="heading-1">
                {t("headline.part1")}{" "}
                <span className="text-[var(--accent-primary)]">{t("headline.part2")}</span>
              </h1>
              <p className="body-text text-xl md:text-2xl mb-8">
                {t("subheadline")}
              </p>
              <InstallCommand />
              <p className="text-sm text-[var(--text-tertiary)] mt-4 font-[family-name:var(--font-jetbrains)]">
                {t("platforms")}
              </p>
            </div>
            <div className="cli-hero-terminal">
              <TerminalDemo />
            </div>
          </div>
        </section>

        {/* Features Section */}
        <section className="section">
          <div className="container">
            <h2 className="heading-2 text-center mb-12">{t("features.title")}</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              {features.map((feature) => (
                <div key={feature.title} className="feature-card">
                  <span className="feature-icon">{feature.icon}</span>
                  <h3 className="feature-title">{feature.title}</h3>
                  <p className="feature-description">{feature.description}</p>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Commands Section */}
        <section className="section bg-[var(--bg-elevated)]">
          <div className="container">
            <h2 className="heading-2 text-center mb-4">{t("commands.title")}</h2>
            <p className="body-text text-center mb-12 max-w-2xl mx-auto">
              {t("commands.subtitle")}
            </p>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {commands.map((cmd) => (
                <CommandCard key={cmd.command} {...cmd} />
              ))}
            </div>
          </div>
        </section>

        {/* PAL Format Section */}
        <section className="section">
          <div className="container">
            <div className="pal-section">
              <div className="pal-text">
                <h2 className="heading-2 mb-4">{t("pal.title")}</h2>
                <p className="body-text mb-6">{t("pal.description")}</p>
                <ul className="pal-features">
                  <li>{t("pal.features.declarative")}</li>
                  <li>{t("pal.features.versionable")}</li>
                  <li>{t("pal.features.validatable")}</li>
                  <li>{t("pal.features.portable")}</li>
                </ul>
              </div>
              <div className="pal-code">
                <div className="terminal-window">
                  <div className="terminal-header">
                    <div className="terminal-dots">
                      <span className="dot red" />
                      <span className="dot yellow" />
                      <span className="dot green" />
                    </div>
                    <span className="terminal-title">agent.yaml</span>
                  </div>
                  <div className="terminal-body">
                    <pre className="text-sm">
{`apiVersion: passflow/v1
kind: Agent
metadata:
  name: data-pipeline
  description: Processes daily reports
spec:
  model: claude-3-5-sonnet-20241022
  systemPrompt: |
    You analyze data and generate reports.
  tools:
    - name: web-search
    - name: calculator
  triggers:
    - type: cron
      schedule: "0 9 * * *"`}
                    </pre>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* CTA Section */}
        <section className="section bg-[var(--bg-surface)]">
          <div className="container text-center">
            <h2 className="heading-2 mb-4">{t("cta.title")}</h2>
            <p className="body-text mb-8 max-w-xl mx-auto">{t("cta.description")}</p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <a href="https://github.com/jaak-ai/passflow-cli" target="_blank" rel="noopener noreferrer" className="btn-primary">
                {t("cta.github")}
              </a>
              <a href="/docs" className="btn-secondary">
                {t("cta.docs")}
              </a>
            </div>
          </div>
        </section>
      </main>
      <Footer />

      <style jsx>{`
        .cli-hero {
          min-height: 100vh;
          display: flex;
          align-items: center;
          padding: 120px 0 80px;
          background: var(--gradient-hero);
          position: relative;
          overflow: hidden;
        }

        .cli-hero::before {
          content: '';
          position: absolute;
          inset: 0;
          background-image: radial-gradient(circle, var(--border-primary) 1px, transparent 1px);
          background-size: 32px 32px;
          opacity: 0.4;
          mask-image: radial-gradient(ellipse at center, black 30%, transparent 70%);
        }

        .cli-hero-content {
          max-width: var(--max-width-content);
          margin: 0 auto;
          padding: 0 var(--space-4);
          display: grid;
          grid-template-columns: 1fr;
          gap: var(--space-12);
          position: relative;
          z-index: 1;
        }

        @media (min-width: 1024px) {
          .cli-hero-content {
            grid-template-columns: 1fr 1fr;
            align-items: center;
          }
        }

        .cli-hero-text {
          max-width: 600px;
        }

        .badge-pill {
          display: inline-flex;
          align-items: center;
          font-family: var(--font-mono);
          font-size: var(--text-xs);
          font-weight: 500;
          text-transform: uppercase;
          letter-spacing: 0.05em;
          padding: var(--space-1) var(--space-3);
          border-radius: var(--radius-full);
          background: var(--accent-primary-ghost);
          color: var(--accent-primary);
          border: 1px solid rgba(0, 255, 135, 0.2);
          margin-bottom: var(--space-4);
        }

        .install-command {
          display: flex;
          align-items: center;
          gap: var(--space-3);
          background: var(--bg-surface);
          border: 1px solid var(--border-primary);
          border-radius: var(--radius-lg);
          padding: var(--space-4);
          cursor: pointer;
          transition: border-color var(--duration-normal);
          overflow-x: auto;
        }

        .install-command:hover {
          border-color: var(--accent-primary-muted);
        }

        .install-command code {
          font-family: var(--font-mono);
          font-size: var(--text-sm);
          color: var(--accent-primary);
          white-space: nowrap;
          flex: 1;
        }

        .copy-btn {
          background: transparent;
          border: none;
          color: var(--text-secondary);
          cursor: pointer;
          padding: var(--space-1);
          flex-shrink: 0;
        }

        .terminal-window {
          background: var(--bg-primary);
          border: 1px solid var(--border-primary);
          border-radius: var(--radius-lg);
          overflow: hidden;
          box-shadow: var(--shadow-lg);
        }

        .terminal-header {
          background: var(--bg-surface);
          padding: var(--space-3) var(--space-4);
          display: flex;
          align-items: center;
          gap: var(--space-3);
          border-bottom: 1px solid var(--border-primary);
        }

        .terminal-dots {
          display: flex;
          gap: var(--space-2);
        }

        .dot {
          width: 12px;
          height: 12px;
          border-radius: 50%;
        }

        .dot.red { background: #FF5F57; }
        .dot.yellow { background: #FEBC2E; }
        .dot.green { background: #28C840; }

        .terminal-title {
          font-family: var(--font-mono);
          font-size: var(--text-xs);
          color: var(--text-tertiary);
        }

        .terminal-body {
          padding: var(--space-6);
          font-family: var(--font-mono);
          font-size: var(--text-sm);
          min-height: 300px;
          line-height: 1.65;
        }

        .terminal-line {
          margin-bottom: var(--space-1);
        }

        .terminal-cursor {
          color: var(--accent-primary);
          animation: blink 1s step-end infinite;
        }

        @keyframes blink {
          50% { opacity: 0; }
        }

        .section {
          padding: var(--space-20) 0;
        }

        @media (min-width: 1024px) {
          .section {
            padding: var(--space-24) 0;
          }
        }

        .container {
          max-width: var(--max-width-content);
          margin: 0 auto;
          padding: 0 var(--space-4);
        }

        @media (min-width: 768px) {
          .container {
            padding: 0 var(--space-6);
          }
        }

        .feature-card {
          background: var(--bg-elevated);
          border: 1px solid var(--border-primary);
          border-radius: var(--radius-lg);
          padding: var(--space-6);
          transition: border-color var(--duration-normal), box-shadow var(--duration-normal);
        }

        .feature-card:hover {
          border-color: var(--accent-primary-muted);
          box-shadow: var(--shadow-glow-green);
        }

        .feature-icon {
          font-size: 2rem;
          margin-bottom: var(--space-3);
          display: block;
        }

        .feature-title {
          font-family: var(--font-body);
          font-size: var(--text-lg);
          font-weight: 600;
          color: var(--text-primary);
          margin-bottom: var(--space-2);
        }

        .feature-description {
          font-size: var(--text-sm);
          color: var(--text-secondary);
          line-height: 1.5;
        }

        .command-card {
          background: var(--bg-surface);
          border: 1px solid var(--border-primary);
          border-radius: var(--radius-lg);
          padding: var(--space-6);
          transition: border-color var(--duration-normal);
        }

        .command-card:hover {
          border-color: var(--accent-primary-muted);
        }

        .command-name {
          font-family: var(--font-mono);
          font-size: var(--text-base);
          font-weight: 600;
          color: var(--accent-primary);
          display: block;
          margin-bottom: var(--space-3);
        }

        .command-description {
          font-size: var(--text-sm);
          color: var(--text-secondary);
          margin-bottom: var(--space-4);
          line-height: 1.5;
        }

        .command-example {
          background: var(--bg-primary);
          border-radius: var(--radius-md);
          padding: var(--space-3);
        }

        .command-example code {
          font-family: var(--font-mono);
          font-size: var(--text-xs);
          color: var(--text-primary);
        }

        .pal-section {
          display: grid;
          grid-template-columns: 1fr;
          gap: var(--space-12);
          align-items: center;
        }

        @media (min-width: 1024px) {
          .pal-section {
            grid-template-columns: 1fr 1fr;
          }
        }

        .pal-features {
          list-style: none;
          padding: 0;
        }

        .pal-features li {
          position: relative;
          padding-left: var(--space-6);
          margin-bottom: var(--space-3);
          color: var(--text-secondary);
        }

        .pal-features li::before {
          content: '✓';
          position: absolute;
          left: 0;
          color: var(--accent-primary);
          font-weight: bold;
        }

        .pal-code pre {
          color: var(--text-primary);
          white-space: pre-wrap;
          word-break: break-word;
        }
      `}</style>
    </>
  );
}
