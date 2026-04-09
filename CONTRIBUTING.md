# Contributing to Passflow

Thank you for your interest in contributing to Passflow! This document provides guidelines and information for contributors.

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

### Reporting Issues

- Check existing issues before creating a new one
- Use issue templates when available
- Include reproduction steps, expected behavior, and actual behavior
- Include environment details (OS, Go version, etc.)

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Write/update tests
5. Run tests (`make test`)
6. Run linter (`make lint`)
7. Commit with a clear message
8. Push and create a Pull Request

### Commit Messages

Follow conventional commits:

```
type(scope): description

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:
- `feat(cli): add export command for PAL files`
- `fix(executor): handle nil pointer in tool response`
- `docs: update installation instructions`

### Code Style

- Follow Go conventions (`gofmt`, `go vet`)
- Write tests for new functionality
- Keep functions small and focused
- Document exported functions and types

## Development Setup

```bash
# Clone
git clone https://github.com/passflow-ai/passflow.git
cd passflow

# Install dependencies
go mod download

# Build all binaries
make build

# Run tests
make test

# Run linter
make lint
```

## Project Structure

```
passflow/
├── cmd/                    # Entry points
│   ├── passflow-cli/       # CLI tool
│   ├── passflow-executor/  # Agent executor
│   ├── passflow-channels/  # Event triggers
│   └── passflow-mcp-gateway/ # MCP gateway
├── pkg/                    # Shared packages
│   ├── common/             # Common utilities
│   └── pal/                # PAL parser
├── examples/               # Example agents
└── docs/                   # Documentation
```

## Questions?

- Open a [Discussion](https://github.com/passflow-ai/passflow/discussions)
- Join our community chat

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
