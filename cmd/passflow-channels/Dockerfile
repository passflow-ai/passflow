# =============================================================================
# PASSFLOW CHANNELS SERVICE - Multi-stage Dockerfile
# =============================================================================
# Build: docker build -t passflow-channels .
# Run:   docker run -p 8081:8081 passflow-channels
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Build
# -----------------------------------------------------------------------------
FROM golang:1.25-alpine AS builder

ENV GOTOOLCHAIN=auto

# Build arguments
ARG VERSION=dev
ARG GIT_COMMIT=unknown

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build the application with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT}" \
    -o /app/passflow-channels \
    .

# -----------------------------------------------------------------------------
# Stage 2: Production
# -----------------------------------------------------------------------------
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 -S passflow && \
    adduser -u 1000 -S passflow -G passflow

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/passflow-channels /app/passflow-channels

# Create necessary directories
RUN mkdir -p /app/logs && \
    chown -R passflow:passflow /app

# Switch to non-root user
USER passflow

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Set environment defaults
ENV ADDR=:8081 \
    PASSFLOW_ENV=production

# Run the application
ENTRYPOINT ["/app/passflow-channels"]
