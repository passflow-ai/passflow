FROM golang:1.24-alpine AS builder

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy entire monorepo for local module resolution
COPY go.mod go.sum ./
COPY pkg/ pkg/
COPY cmd/passflow-executor/ cmd/passflow-executor/

# Download dependencies
RUN go mod download

# Build the executor
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o passflow-executor \
    ./cmd/passflow-executor

# Runtime image
FROM alpine:3.20

ARG K8SGPT_VERSION=v0.4.26

RUN apk --no-cache add ca-certificates git curl && \
    curl -LO "https://github.com/k8sgpt-ai/k8sgpt/releases/download/${K8SGPT_VERSION}/k8sgpt_amd64.apk" && \
    apk add --allow-untrusted k8sgpt_amd64.apk && \
    rm -f k8sgpt_amd64.apk && \
    curl -LO "https://dl.k8s.io/release/v1.31.0/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

# Create non-root user
RUN addgroup -g 1000 passflow && \
    adduser -u 1000 -G passflow -s /bin/sh -D passflow

WORKDIR /app

COPY --from=builder /app/passflow-executor .

RUN mkdir -p /tmp/passflow && chown passflow:passflow /tmp/passflow

USER passflow

ENTRYPOINT ["./passflow-executor"]
