FROM golang:1.23-alpine AS builder

ENV GOTOOLCHAIN=auto
ENV GOPRIVATE=github.com/jaak-ai/*

# Install build dependencies
RUN apk add --no-cache git ca-certificates openssh-client

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Configure SSH for private repos (requires GITHUB_TOKEN secret)
RUN --mount=type=secret,id=github_token \
    git config --global url."https://oauth2:$(cat /run/secrets/github_token)@github.com".insteadOf "https://github.com"

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=secret,id=github_token \
    go mod download

# Copy source code
COPY . .

# Build with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=secret,id=github_token \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o passflow-agent-executor .

# Runtime image
FROM alpine:3.19

ARG K8SGPT_VERSION=v0.4.26

# Install required tools
RUN apk --no-cache add ca-certificates git openssh-client curl && \
    curl -LO "https://github.com/k8sgpt-ai/k8sgpt/releases/download/${K8SGPT_VERSION}/k8sgpt_amd64.apk" && \
    apk add --allow-untrusted k8sgpt_amd64.apk && \
    rm -f k8sgpt_amd64.apk && \
    curl -LO "https://dl.k8s.io/release/v1.31.0/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

WORKDIR /app

COPY --from=builder /app/passflow-agent-executor .

RUN mkdir -p /tmp/passflow && chmod 777 /tmp/passflow

ENTRYPOINT ["./passflow-agent-executor"]
