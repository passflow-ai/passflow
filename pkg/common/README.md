# passflow-common

Shared Go library for Passflow microservices.

## Packages

| Package | Description |
|---------|-------------|
| `database` | MongoDB and Redis clients |
| `eventbus` | Redis pub/sub event system |
| `queue` | Redis Stream job queue |
| `guardian` | JWT authentication |
| `logger` | Structured logging (Zap) |
| `msg` | Standardized error responses |
| `setup` | Service configuration |
| `crypto` | Cryptographic utilities |
| `cel` | CEL expression evaluation |

## Installation

```bash
go get github.com/jaak-ai/passflow-common
```

## Usage

```go
import (
    "github.com/jaak-ai/passflow-common/database"
    "github.com/jaak-ai/passflow-common/logger"
)
```

## Private Module Configuration

```bash
go env -w GOPRIVATE=github.com/jaak-ai/*
```
