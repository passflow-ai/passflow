'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';

// Flow step config - only structural data, text comes from translations
const flowStepConfig = [
  { id: 'trigger', tooltipKey: null },
  { id: 'channels', tooltipKey: 'celRules' },
  { id: 'redis', tooltipKey: null },
  { id: 'k8s', tooltipKey: 'ephemeralPod' },
  { id: 'react', tooltipKey: 'reactLoop' },
  { id: 'result', tooltipKey: null },
  { id: 'cleanup', tooltipKey: null },
];

// Tab IDs for iteration
const tabIds = ['k8s', 'outbound', 'gitops'] as const;

// YAML configs are not translatable - they are code examples
const yamlConfigs: Record<string, string> = {
  k8s: `agent:
  name: k8s-self-healing
  trigger:
    type: webhook
    source: prometheus-alertmanager
  react:
    max_iterations: 5
    tools:
      - kubectl_get_pods
      - kubectl_describe
      - kubectl_rollout_restart
      - slack_notify
  pod:
    cpu: "500m"
    memory: "512Mi"
    ttl: 300s`,
  outbound: `agent:
  name: outbound-enrichment
  trigger:
    type: cron
    schedule: "0 8 * * 1-5"
  react:
    max_iterations: 8
    tools:
      - crm_get_prospects
      - enrichment_api
      - llm_classify_icp
      - email_send_sequence
      - slack_notify_sales
  pod:
    cpu: "250m"
    memory: "256Mi"
    ttl: 600s`,
  gitops: `agent:
  name: gitops-reconciler
  trigger:
    type: cron
    schedule: "*/15 * * * *"
  react:
    max_iterations: 6
    tools:
      - git_diff_manifests
      - kubectl_get_current_state
      - diff_generator
      - human_approval
      - kubectl_apply
  approval:
    required: true
    channel: slack
    timeout: 30m
  pod:
    cpu: "500m"
    memory: "512Mi"
    ttl: 180s`,
};

const Tooltip = ({ tooltipText, children }: { tooltipText: string; children: React.ReactNode }) => {
  return (
    <span className="tooltip-wrapper">
      <span className="tooltip-trigger">{children}</span>
      <span className="tooltip-content">{tooltipText}</span>
    </span>
  );
};

const FlowDiagram = ({ t }: { t: ReturnType<typeof useTranslations<'HowItWorks'>> }) => {
  return (
    <div className="flow-diagram">
      {flowStepConfig.map((step, index) => (
        <div key={step.id} className="flow-step">
          <div
            className={`flow-step-box ${index === 0 || index === flowStepConfig.length - 1 ? 'flow-step-accent' : ''}`}
          >
            <span className="flow-step-label">
              {step.tooltipKey ? (
                <Tooltip tooltipText={t(`tooltips.${step.tooltipKey}`)}>
                  {t(`flowSteps.${step.id}.label`)}
                </Tooltip>
              ) : (
                t(`flowSteps.${step.id}.label`)
              )}
            </span>
          </div>
          {index < flowStepConfig.length - 1 && (
            <span className="flow-arrow">→</span>
          )}
        </div>
      ))}
    </div>
  );
};

const Terminal = ({ yaml, filename }: { yaml: string; filename: string }) => {
  return (
    <div className="terminal">
      {/* Title bar */}
      <div className="terminal-titlebar">
        {/* Traffic lights */}
        <span className="terminal-light terminal-light-red" />
        <span className="terminal-light terminal-light-yellow" />
        <span className="terminal-light terminal-light-green" />
        <span className="terminal-filename">{filename}</span>
      </div>

      {/* Terminal body */}
      <div className="terminal-body">
        <pre className="terminal-code">{yaml}</pre>
      </div>
    </div>
  );
};

const HowItWorks = () => {
  const t = useTranslations('HowItWorks');
  const [activeTab, setActiveTab] = useState<(typeof tabIds)[number]>(tabIds[0]);

  return (
    <section id="how-it-works" className="section how-it-works-section">
      <div className="container">
        {/* Header */}
        <div className="how-it-works-header">
          <h2 className="heading-2">
            {t('headline')}
          </h2>
          <p className="body-text how-it-works-subheadline">
            {t('subheadline')}
          </p>
        </div>

        {/* Flow Diagram */}
        <div className="flow-diagram-wrapper">
          <FlowDiagram t={t} />
        </div>

        {/* Flow Steps Description */}
        <div className="flow-steps-grid">
          {flowStepConfig.map((step, index) => (
            <div key={step.id} className="flow-step-description">
              <span className="flow-step-number">{index + 1}</span>
              <p className="flow-step-text">{t(`flowSteps.${step.id}.description`)}</p>
            </div>
          ))}
        </div>

        {/* Tabs Section */}
        <div className="tabs-container">
          {/* Tab Bar */}
          <div className="tabs-bar">
            {tabIds.map((tabId) => (
              <button
                key={tabId}
                onClick={() => setActiveTab(tabId)}
                className={`tab-button ${activeTab === tabId ? 'tab-button-active' : ''}`}
              >
                {t(`tabs.${tabId}.shortTitle`)}
              </button>
            ))}
          </div>

          {/* Tab Content */}
          <div className="tabs-content">
            <div className="tabs-content-grid">
              {/* Narrative */}
              <div>
                <h3 className="heading-4 tab-title">{t(`tabs.${activeTab}.title`)}</h3>
                <p className="body-text-sm">{t(`tabs.${activeTab}.narrative`)}</p>
              </div>

              {/* Terminal */}
              <Terminal
                yaml={yamlConfigs[activeTab]}
                filename={`${activeTab}-agent.yaml`}
              />
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default HowItWorks;
