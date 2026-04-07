"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import Header from "@/components/Header";
import Footer from "@/components/Footer";
import {
  Book,
  Zap,
  Bot,
  GitBranch,
  Bell,
  Wrench,
  Shield,
  Code,
  Terminal,
  ChevronRight,
  ExternalLink,
} from "lucide-react";

type NavItem = {
  id: string;
  icon: React.ReactNode;
};

const navItems: NavItem[] = [
  { id: "introduction", icon: <Book size={18} /> },
  { id: "quickstart", icon: <Zap size={18} /> },
  { id: "agents", icon: <Bot size={18} /> },
  { id: "workflows", icon: <GitBranch size={18} /> },
  { id: "triggers", icon: <Bell size={18} /> },
  { id: "tools", icon: <Wrench size={18} /> },
  { id: "api", icon: <Code size={18} /> },
  { id: "security", icon: <Shield size={18} /> },
];

const CodeBlock = ({ code, filename }: { code: string; filename?: string }) => (
  <div className="terminal" style={{ marginBottom: "var(--space-6)" }}>
    {filename && (
      <div className="terminal-titlebar">
        <span className="terminal-light terminal-light-red" />
        <span className="terminal-light terminal-light-yellow" />
        <span className="terminal-light terminal-light-green" />
        <span className="terminal-filename">{filename}</span>
      </div>
    )}
    <div className="terminal-body">
      <pre className="terminal-code">{code}</pre>
    </div>
  </div>
);

export default function DocsPage() {
  const t = useTranslations("Docs");
  const [activeSection, setActiveSection] = useState("introduction");

  const scrollToSection = (id: string) => {
    setActiveSection(id);
    const element = document.getElementById(id);
    if (element) {
      element.scrollIntoView({ behavior: "smooth", block: "start" });
    }
  };

  return (
    <>
      <Header />
      <main className="pt-20" style={{ background: "var(--bg-primary)", minHeight: "100vh" }}>
        <div className="max-w-[var(--max-width-content)] mx-auto px-4 md:px-6 lg:px-8 py-12">
          <div className="flex gap-8">
            {/* Sidebar */}
            <aside className="hidden lg:block w-64 flex-shrink-0">
              <nav className="sticky top-24">
                <p className="label mb-4" style={{ color: "var(--text-tertiary)" }}>
                  {t("nav.title")}
                </p>
                <ul className="space-y-1">
                  {navItems.map((item) => (
                    <li key={item.id}>
                      <button
                        onClick={() => scrollToSection(item.id)}
                        className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-left transition-colors ${
                          activeSection === item.id
                            ? "bg-[var(--bg-elevated)] text-[var(--accent-primary)]"
                            : "text-[var(--text-secondary)] hover:text-[var(--text-primary)] hover:bg-[var(--bg-surface)]"
                        }`}
                      >
                        {item.icon}
                        <span className="text-sm font-medium">{t(`nav.${item.id}`)}</span>
                      </button>
                    </li>
                  ))}
                </ul>
              </nav>
            </aside>

            {/* Content */}
            <div className="flex-1 max-w-3xl">
              {/* Introduction */}
              <section id="introduction" className="mb-16">
                <h1 className="heading-1 mb-6">{t("introduction.title")}</h1>
                <p className="body-text mb-6">{t("introduction.description")}</p>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
                  {["react", "isolation", "multiProvider", "triggers"].map((feature) => (
                    <div
                      key={feature}
                      className="card"
                      style={{ padding: "var(--space-4)" }}
                    >
                      <h4 className="heading-5 mb-2">{t(`introduction.features.${feature}.title`)}</h4>
                      <p className="body-text-sm" style={{ color: "var(--text-secondary)" }}>
                        {t(`introduction.features.${feature}.description`)}
                      </p>
                    </div>
                  ))}
                </div>
              </section>

              {/* Quick Start */}
              <section id="quickstart" className="mb-16">
                <h2 className="heading-2 mb-6">{t("quickstart.title")}</h2>
                <p className="body-text mb-6">{t("quickstart.description")}</p>

                <div className="space-y-6">
                  <div>
                    <h3 className="heading-4 mb-3">
                      <span className="text-[var(--accent-primary)] mr-2">1.</span>
                      {t("quickstart.step1.title")}
                    </h3>
                    <p className="body-text-sm mb-4" style={{ color: "var(--text-secondary)" }}>
                      {t("quickstart.step1.description")}
                    </p>
                    <a
                      href="https://app.passflow.ai/login?mode=register"
                      className="btn-primary inline-flex items-center gap-2"
                    >
                      {t("quickstart.step1.cta")}
                      <ExternalLink size={16} />
                    </a>
                  </div>

                  <div>
                    <h3 className="heading-4 mb-3">
                      <span className="text-[var(--accent-primary)] mr-2">2.</span>
                      {t("quickstart.step2.title")}
                    </h3>
                    <p className="body-text-sm mb-4" style={{ color: "var(--text-secondary)" }}>
                      {t("quickstart.step2.description")}
                    </p>
                  </div>

                  <div>
                    <h3 className="heading-4 mb-3">
                      <span className="text-[var(--accent-primary)] mr-2">3.</span>
                      {t("quickstart.step3.title")}
                    </h3>
                    <p className="body-text-sm mb-4" style={{ color: "var(--text-secondary)" }}>
                      {t("quickstart.step3.description")}
                    </p>
                    <CodeBlock
                      filename="my-first-agent.yaml"
                      code={`agent:
  name: slack-responder
  persona: |
    You are a helpful assistant that responds
    to questions in Slack channels.
  model: claude-sonnet-4-6
  temperature: 0.7
  max_tokens: 1024

trigger:
  type: slack
  channel: "#support"
  match: "@passflow"

tools:
  - slack_reply
  - search_docs`}
                    />
                  </div>

                  <div>
                    <h3 className="heading-4 mb-3">
                      <span className="text-[var(--accent-primary)] mr-2">4.</span>
                      {t("quickstart.step4.title")}
                    </h3>
                    <p className="body-text-sm" style={{ color: "var(--text-secondary)" }}>
                      {t("quickstart.step4.description")}
                    </p>
                  </div>
                </div>
              </section>

              {/* Agents */}
              <section id="agents" className="mb-16">
                <h2 className="heading-2 mb-6">{t("agents.title")}</h2>
                <p className="body-text mb-6">{t("agents.description")}</p>

                <h3 className="heading-4 mb-4">{t("agents.config.title")}</h3>
                <CodeBlock
                  filename="agent-config.yaml"
                  code={`agent:
  name: k8s-ops-agent
  persona: |
    You are a Kubernetes operations expert.
    You help diagnose and resolve cluster issues.
  instructions: |
    1. Always check pod status first
    2. Look at logs for error patterns
    3. Suggest fixes but ask for approval
  model: claude-sonnet-4-6
  temperature: 0.3
  max_tokens: 2048

react:
  max_iterations: 5
  stop_on_error: false

pod:
  cpu: "500m"
  memory: "512Mi"
  ttl: 300s`}
                />

                <h3 className="heading-4 mb-4 mt-8">{t("agents.models.title")}</h3>
                <div className="overflow-x-auto">
                  <table className="comparison-table">
                    <thead>
                      <tr>
                        <th>{t("agents.models.headers.provider")}</th>
                        <th>{t("agents.models.headers.models")}</th>
                        <th>{t("agents.models.headers.useCase")}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr>
                        <td>Anthropic</td>
                        <td><code>claude-sonnet-4-6</code>, <code>claude-opus-4</code></td>
                        <td>{t("agents.models.anthropic")}</td>
                      </tr>
                      <tr>
                        <td>OpenAI</td>
                        <td><code>gpt-4o</code>, <code>gpt-4-turbo</code></td>
                        <td>{t("agents.models.openai")}</td>
                      </tr>
                      <tr>
                        <td>Google</td>
                        <td><code>gemini-1.5-pro</code>, <code>gemini-1.5-flash</code></td>
                        <td>{t("agents.models.google")}</td>
                      </tr>
                      <tr>
                        <td>AWS Bedrock</td>
                        <td><code>claude-*</code>, <code>titan-*</code></td>
                        <td>{t("agents.models.bedrock")}</td>
                      </tr>
                      <tr>
                        <td>Azure OpenAI</td>
                        <td><code>gpt-4</code>, <code>gpt-35-turbo</code></td>
                        <td>{t("agents.models.azure")}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </section>

              {/* Workflows */}
              <section id="workflows" className="mb-16">
                <h2 className="heading-2 mb-6">{t("workflows.title")}</h2>
                <p className="body-text mb-6">{t("workflows.description")}</p>

                <h3 className="heading-4 mb-4">{t("workflows.nodes.title")}</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
                  {["agent", "condition", "action", "approval"].map((node) => (
                    <div key={node} className="card" style={{ padding: "var(--space-4)" }}>
                      <h4 className="heading-5 mb-2">{t(`workflows.nodes.${node}.title`)}</h4>
                      <p className="body-text-sm" style={{ color: "var(--text-secondary)" }}>
                        {t(`workflows.nodes.${node}.description`)}
                      </p>
                    </div>
                  ))}
                </div>

                <CodeBlock
                  filename="workflow-example.yaml"
                  code={`workflow:
  name: incident-response
  steps:
    - id: detect
      type: agent
      agent: anomaly-detector

    - id: check-severity
      type: condition
      expression: "output.severity >= 'high'"

    - id: auto-fix
      type: agent
      agent: k8s-healer
      when: "steps.check-severity.result == false"

    - id: escalate
      type: approval
      channel: slack
      timeout: 30m
      when: "steps.check-severity.result == true"

    - id: manual-fix
      type: agent
      agent: k8s-healer
      when: "steps.escalate.approved == true"`}
                />
              </section>

              {/* Triggers */}
              <section id="triggers" className="mb-16">
                <h2 className="heading-2 mb-6">{t("triggers.title")}</h2>
                <p className="body-text mb-6">{t("triggers.description")}</p>

                <div className="space-y-6">
                  {["cron", "slack", "email", "webhook"].map((trigger) => (
                    <div key={trigger}>
                      <h3 className="heading-4 mb-3">{t(`triggers.types.${trigger}.title`)}</h3>
                      <p className="body-text-sm mb-4" style={{ color: "var(--text-secondary)" }}>
                        {t(`triggers.types.${trigger}.description`)}
                      </p>
                      <CodeBlock
                        code={
                          trigger === "cron"
                            ? `trigger:
  type: cron
  schedule: "0 9 * * 1-5"  # Weekdays at 9am`
                            : trigger === "slack"
                            ? `trigger:
  type: slack
  channel: "#alerts"
  match: "@oncall"
  cel: "message.text.contains('urgent')"`
                            : trigger === "email"
                            ? `trigger:
  type: email
  address: support@yourcompany.com
  cel: "email.subject.contains('[TICKET]')"`
                            : `trigger:
  type: webhook
  path: /incidents
  method: POST
  cel: "body.priority == 'critical'"`
                        }
                      />
                    </div>
                  ))}
                </div>

                <div className="card mt-8" style={{ padding: "var(--space-6)", borderColor: "var(--accent-primary)" }}>
                  <h4 className="heading-5 mb-2">{t("triggers.cel.title")}</h4>
                  <p className="body-text-sm mb-4" style={{ color: "var(--text-secondary)" }}>
                    {t("triggers.cel.description")}
                  </p>
                  <CodeBlock
                    code={`# Match high-priority Slack messages from #incidents
cel: |
  message.channel == '#incidents' &&
  message.text.contains('P1') &&
  message.user != 'bot'`}
                  />
                </div>
              </section>

              {/* Tools */}
              <section id="tools" className="mb-16">
                <h2 className="heading-2 mb-6">{t("tools.title")}</h2>
                <p className="body-text mb-6">{t("tools.description")}</p>

                <h3 className="heading-4 mb-4">{t("tools.builtin.title")}</h3>
                <div className="overflow-x-auto mb-8">
                  <table className="comparison-table">
                    <thead>
                      <tr>
                        <th>{t("tools.builtin.headers.category")}</th>
                        <th>{t("tools.builtin.headers.tools")}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr>
                        <td>Kubernetes</td>
                        <td><code>kubectl_get</code>, <code>kubectl_describe</code>, <code>kubectl_logs</code>, <code>kubectl_rollout</code></td>
                      </tr>
                      <tr>
                        <td>Slack</td>
                        <td><code>slack_send</code>, <code>slack_reply</code>, <code>slack_react</code></td>
                      </tr>
                      <tr>
                        <td>GitHub</td>
                        <td><code>github_pr_create</code>, <code>github_issue_comment</code>, <code>github_file_read</code></td>
                      </tr>
                      <tr>
                        <td>HTTP</td>
                        <td><code>http_get</code>, <code>http_post</code>, <code>http_request</code></td>
                      </tr>
                      <tr>
                        <td>Database</td>
                        <td><code>sql_query</code>, <code>mongo_find</code>, <code>redis_get</code></td>
                      </tr>
                    </tbody>
                  </table>
                </div>

                <h3 className="heading-4 mb-4">{t("tools.mcp.title")}</h3>
                <p className="body-text-sm mb-4" style={{ color: "var(--text-secondary)" }}>
                  {t("tools.mcp.description")}
                </p>
                <CodeBlock
                  filename="mcp-tool.yaml"
                  code={`tools:
  - type: mcp
    server: my-custom-server
    tools:
      - name: search_documents
        description: Search internal docs
      - name: create_ticket
        description: Create support ticket`}
                />
              </section>

              {/* API */}
              <section id="api" className="mb-16">
                <h2 className="heading-2 mb-6">{t("api.title")}</h2>
                <p className="body-text mb-6">{t("api.description")}</p>

                <h3 className="heading-4 mb-4">{t("api.auth.title")}</h3>
                <p className="body-text-sm mb-4" style={{ color: "var(--text-secondary)" }}>
                  {t("api.auth.description")}
                </p>
                <CodeBlock
                  code={`curl -X GET "https://api.passflow.ai/v1/agents" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "X-Workspace-ID: ws_xxxxx"`}
                />

                <h3 className="heading-4 mb-4 mt-8">{t("api.endpoints.title")}</h3>
                <div className="overflow-x-auto">
                  <table className="comparison-table">
                    <thead>
                      <tr>
                        <th>{t("api.endpoints.headers.method")}</th>
                        <th>{t("api.endpoints.headers.endpoint")}</th>
                        <th>{t("api.endpoints.headers.description")}</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr>
                        <td><span className="badge badge-green">GET</span></td>
                        <td><code>/v1/agents</code></td>
                        <td>{t("api.endpoints.listAgents")}</td>
                      </tr>
                      <tr>
                        <td><span className="badge badge-blue">POST</span></td>
                        <td><code>/v1/agents</code></td>
                        <td>{t("api.endpoints.createAgent")}</td>
                      </tr>
                      <tr>
                        <td><span className="badge badge-blue">POST</span></td>
                        <td><code>/v1/agents/:id/execute</code></td>
                        <td>{t("api.endpoints.executeAgent")}</td>
                      </tr>
                      <tr>
                        <td><span className="badge badge-green">GET</span></td>
                        <td><code>/v1/executions</code></td>
                        <td>{t("api.endpoints.listExecutions")}</td>
                      </tr>
                      <tr>
                        <td><span className="badge badge-green">GET</span></td>
                        <td><code>/v1/executions/:id</code></td>
                        <td>{t("api.endpoints.getExecution")}</td>
                      </tr>
                      <tr>
                        <td><span className="badge badge-green">GET</span></td>
                        <td><code>/v1/workflows</code></td>
                        <td>{t("api.endpoints.listWorkflows")}</td>
                      </tr>
                      <tr>
                        <td><span className="badge badge-blue">POST</span></td>
                        <td><code>/v1/workflows/:id/execute</code></td>
                        <td>{t("api.endpoints.executeWorkflow")}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </section>

              {/* Security */}
              <section id="security" className="mb-16">
                <h2 className="heading-2 mb-6">{t("security.title")}</h2>
                <p className="body-text mb-6">{t("security.description")}</p>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {["encryption", "isolation", "audit", "compliance"].map((feature) => (
                    <div key={feature} className="card" style={{ padding: "var(--space-4)" }}>
                      <h4 className="heading-5 mb-2">{t(`security.features.${feature}.title`)}</h4>
                      <p className="body-text-sm" style={{ color: "var(--text-secondary)" }}>
                        {t(`security.features.${feature}.description`)}
                      </p>
                    </div>
                  ))}
                </div>
              </section>

              {/* Help */}
              <section className="card" style={{ padding: "var(--space-8)", background: "var(--bg-elevated)" }}>
                <h3 className="heading-3 mb-4">{t("help.title")}</h3>
                <p className="body-text mb-6" style={{ color: "var(--text-secondary)" }}>
                  {t("help.description")}
                </p>
                <div className="flex flex-wrap gap-4">
                  <a
                    href="https://app.passflow.ai"
                    className="btn-primary inline-flex items-center gap-2"
                  >
                    {t("help.dashboard")}
                    <ChevronRight size={16} />
                  </a>
                  <a
                    href="mailto:support@passflow.ai"
                    className="btn-secondary inline-flex items-center gap-2"
                  >
                    {t("help.contact")}
                    <ExternalLink size={16} />
                  </a>
                </div>
              </section>
            </div>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
