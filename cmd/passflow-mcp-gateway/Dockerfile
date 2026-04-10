# =============================================================================
# PASSFLOW MCP GATEWAY - Multi-stage Dockerfile (Monorepo)
# =============================================================================
FROM golang:1.24-alpine AS builder

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy entire monorepo for local module resolution
COPY go.mod go.sum ./
COPY pkg/ pkg/
COPY cmd/passflow-mcp-gateway/ cmd/passflow-mcp-gateway/

# Download dependencies
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/passflow-mcp-gateway \
    ./cmd/passflow-mcp-gateway

# Runtime
FROM alpine:3.20

RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 -S passflow && \
    adduser -u 1000 -S passflow -G passflow

WORKDIR /app

COPY --from=builder /app/passflow-mcp-gateway /app/passflow-mcp-gateway

USER passflow

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/passflow-mcp-gateway"]
