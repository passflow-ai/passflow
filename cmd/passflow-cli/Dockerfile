# =============================================================================
# PASSFLOW CLI - Multi-stage Dockerfile (Monorepo)
# =============================================================================
# Build: docker build -f cmd/passflow-cli/Dockerfile -t passflow-cli .
# Run:   docker run passflow-cli --help
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Build
# -----------------------------------------------------------------------------
FROM golang:1.24-alpine AS builder

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy entire monorepo for local module resolution
COPY go.mod go.sum ./
COPY pkg/ pkg/
COPY cmd/passflow-cli/ cmd/passflow-cli/

# Download dependencies
RUN go mod download

# Build the CLI
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /passflow \
    ./cmd/passflow-cli

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 passflow && \
    adduser -u 1000 -G passflow -s /bin/sh -D passflow

COPY --from=builder /passflow /usr/local/bin/passflow

USER passflow

ENTRYPOINT ["passflow"]
