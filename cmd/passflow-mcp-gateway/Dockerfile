FROM golang:1.25-alpine AS builder

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags="-w -s" -o passflow-mcp-gateway .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/passflow-mcp-gateway /usr/local/bin/
EXPOSE 8080
CMD ["passflow-mcp-gateway"]
