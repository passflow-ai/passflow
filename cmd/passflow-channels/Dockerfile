# =============================================================================
# PASSFLOW CHANNELS SERVICE - Multi-stage Dockerfile (Monorepo)
# =============================================================================
# Build: docker build -f cmd/passflow-channels/Dockerfile -t passflow-channels .
# Run:   docker run -p 8081:8081 passflow-channels
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Build
# -----------------------------------------------------------------------------
FROM golang:1.24-alpine AS builder

ENV GOTOOLCHAIN=auto
ARG VERSION=dev

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy entire monorepo for local module resolution
COPY go.mod go.sum ./
COPY pkg/ pkg/
COPY cmd/passflow-channels/ cmd/passflow-channels/

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=${VERSION}" \
    -o /app/passflow-channels \
    ./cmd/passflow-channels

# -----------------------------------------------------------------------------
# Stage 2: Production
# -----------------------------------------------------------------------------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 -S passflow && \
    adduser -u 1000 -S passflow -G passflow

WORKDIR /app

COPY --from=builder /app/passflow-channels /app/passflow-channels

RUN mkdir -p /app/logs && chown -R passflow:passflow /app

USER passflow

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

ENV ADDR=:8081 \
    PASSFLOW_ENV=production

ENTRYPOINT ["/app/passflow-channels"]
