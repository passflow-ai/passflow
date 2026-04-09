# Passflow CLI

Command-line interface for managing AI agents via PAL (Passflow Agent Language).

## Installation

### Quick Install (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/jaak-ai/passflow-cli/main/scripts/install.sh | bash
```

### Install specific version

```bash
VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/jaak-ai/passflow-cli/main/scripts/install.sh | bash
```

### Install to custom directory

```bash
INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/jaak-ai/passflow-cli/main/scripts/install.sh | bash
```

### Build from source

```bash
git clone https://github.com/jaak-ai/passflow-cli.git
cd passflow-cli
make build
make install
```

## Quick Start

```bash
# Configure API endpoint
passflow config set api-url https://api.passflow.ai

# Authenticate
passflow login --token <your-jwt-token>

# Set default workspace
passflow config set workspace ws-abc123

# Validate a PAL file
passflow pal validate agent.yaml

# Apply configuration
passflow pal apply agent.yaml

# Export existing agent
passflow pal export <agent-id> -o agent.yaml
```

## Commands

### Authentication

```bash
# Login with JWT token
passflow login --token eyJhbGciOiJIUzI1NiIs...
passflow login -t <token>

# Logout (remove stored credentials)
passflow logout
```

### Configuration

```bash
# Set configuration values
passflow config set api-url https://api.passflow.ai
passflow config set workspace ws-abc123

# Get configuration values
passflow config get api-url
passflow config get workspace

# List all configuration
passflow config list
```

Configuration is stored in `~/.passflow/config.yaml`.

### PAL Commands

```bash
# Validate syntax
passflow pal validate agent.yaml
passflow pal validate agent.yaml --mode warn

# Apply (create/update agent)
passflow pal apply agent.yaml
passflow pal apply agent.yaml --dry-run
passflow pal apply agent.yaml -w <workspace>

# Export agent to PAL
passflow pal export <agent-id>
passflow pal export <agent-id> -o agent.yaml
passflow pal export <agent-id> --format json

# Compare agent with PAL file
passflow pal diff <agent-id> agent.yaml
```

### Agent Management

```bash
# List agents
passflow agents list
passflow agents list -w <workspace>

# Get agent details
passflow agents get <agent-id>
passflow agents get <agent-id> --format json
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--config` | Config file path (default: `~/.passflow/config.yaml`) |
| `-o, --output` | Output format: table, json, yaml |
| `-w, --workspace` | Workspace ID (overrides default) |

## Development

```bash
# Install dependencies
make deps

# Run with arguments
make dev ARGS="pal validate agent.yaml"

# Run tests
make test

# Lint
make lint
```

## PAL File Format

```yaml
apiVersion: passflow/v1
kind: Agent
metadata:
  name: my-agent
  description: Agent description
spec:
  model: claude-3-5-sonnet-20241022
  systemPrompt: |
    You are a helpful assistant.
  tools:
    - name: web-search
      config:
        maxResults: 5
  triggers:
    - type: slack
      channel: "#general"
```
