# Passflow

Open-source AI agent orchestration platform. Define agents declaratively with [PAL](https://github.com/passflow-ai/pal), execute them with the ReAct pattern, and integrate with any tool via MCP.

## Features

- **Declarative Agents** - Define agents as code using PAL (YAML-based DSL)
- **ReAct Execution** - Reason + Act loop with tool calling
- **Multi-Provider LLM** - Anthropic, OpenAI, Bedrock, Azure, Gemini
- **MCP Integration** - Use any MCP-compatible tool
- **Event-Driven** - Triggers via cron, webhooks, Slack, email
- **Self-Hosted** - Run on your infrastructure

## Quick Start

### Install CLI

```bash
# macOS
brew install passflow-ai/tap/passflow

# Linux
curl -sSL https://passflow.ai/install.sh | bash

# From source
go install github.com/passflow-ai/passflow/cmd/passflow-cli@latest
```

### Define an Agent

Create `my-agent.pal.yaml`:

```yaml
apiVersion: pal/v1
kind: Agent

metadata:
  name: code-reviewer
  description: Reviews pull requests

spec:
  persona: You are a senior engineer who reviews code constructively.

  model:
    provider: anthropic
    id: claude-sonnet-4-20250514

  tools:
    - github.get_pull_request
    - github.create_review
```

### Validate and Apply

```bash
# Validate the agent definition
passflow pal validate my-agent.pal.yaml

# Apply to your workspace
passflow pal apply my-agent.pal.yaml

# List agents
passflow agents list
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Passflow                                │
├─────────────┬─────────────┬─────────────┬─────────────┬────────┤
│   CLI       │  Executor   │  Channels   │ MCP Gateway │ Common │
│             │             │             │             │        │
│ PAL parse   │ ReAct loop  │ Triggers:   │ Tool proxy  │ Shared │
│ Agent mgmt  │ LLM calls   │ - Cron      │ MCP <-> HTTP│ libs   │
│ Validation  │ Tool exec   │ - Webhook   │             │        │
│             │             │ - Slack     │             │        │
│             │             │ - Email     │             │        │
└─────────────┴─────────────┴─────────────┴─────────────┴────────┘
                              │
                              ▼
                        ┌───────────┐
                        │   Redis   │  Event bus + Job queue
                        └───────────┘
```

## Components

| Component | Description |
|-----------|-------------|
| `passflow-cli` | Command-line interface for managing agents |
| `passflow-executor` | ReAct loop agent executor |
| `passflow-channels` | Event triggers (cron, webhook, Slack, email) |
| `passflow-mcp-gateway` | MCP protocol gateway |
| `pkg/common` | Shared libraries |
| `pkg/pal` | PAL parser and validator |

## Documentation

- [PAL Specification](https://github.com/passflow-ai/pal) - Agent definition language
- [Architecture](docs/architecture.md) - System design
- [Development](docs/development.md) - Contributing guide

## Self-Hosting

Requires:
- Redis (event bus + job queue)
- MongoDB (optional, for persistence)

```bash
# Docker Compose
docker-compose up -d

# Kubernetes
kubectl apply -f k8s/
```

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `REDIS_URL` | Redis connection URL | `redis://localhost:6379` |
| `PASSFLOW_API_URL` | API URL | `http://localhost:8080` |
| `ANTHROPIC_API_KEY` | Anthropic API key | - |
| `OPENAI_API_KEY` | OpenAI API key | - |

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
