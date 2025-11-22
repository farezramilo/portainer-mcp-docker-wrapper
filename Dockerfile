# syntax=docker/dockerfile:1

# Stage 1: Build the Go wrapper
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the wrapper binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o wrapper \
    ./cmd/wrapper

# Stage 2: Download Portainer MCP binary
FROM alpine:latest AS downloader

ARG PORTAINER_MCP_VERSION=v0.6.0
ARG TARGETARCH=amd64

WORKDIR /download

# Install download tools
RUN apk add --no-cache curl tar

# Download Portainer MCP for the target architecture
RUN curl -fsSL \
    "https://github.com/portainer/portainer-mcp/releases/download/${PORTAINER_MCP_VERSION}/portainer-mcp-${PORTAINER_MCP_VERSION}-linux-${TARGETARCH}.tar.gz" \
    -o portainer-mcp.tar.gz && \
    tar -xzf portainer-mcp.tar.gz && \
    chmod +x portainer-mcp

# Stage 3: Final runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    wget \
    && addgroup -g 1000 mcpuser \
    && adduser -D -u 1000 -G mcpuser mcpuser

WORKDIR /app

# Copy binaries from previous stages
COPY --from=builder /build/wrapper /app/wrapper
COPY --from=downloader /download/portainer-mcp /app/portainer-mcp

# Set ownership
RUN chown -R mcpuser:mcpuser /app

# Create config directory
RUN mkdir -p /config && chown mcpuser:mcpuser /config

# Switch to non-root user
USER mcpuser

# Expose MCP HTTP port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --spider -q http://localhost:${MCP_PORT:-8080}/health || exit 1

# Set environment defaults
ENV MCP_PORT=8080 \
    MCP_BINARY_PATH=/app/portainer-mcp

# Run the wrapper
ENTRYPOINT ["/app/wrapper"]
