# Code Examples and Implementation Reference

## Go Wrapper Implementation

### Main Wrapper Structure

**File**: `cmd/wrapper/main.go`

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/modelcontextprotocol/go-sdk/mcp"
    "portainer-mcp-wrapper/internal/auth"
    "portainer-mcp-wrapper/internal/bridge"
    "portainer-mcp-wrapper/internal/config"
)

func main() {
    // Load configuration from environment
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Create MCP server instance for each request
    getServer := func(req *http.Request) *mcp.Server {
        mcpBridge, err := bridge.NewPortainerMCPBridge(cfg)
        if err != nil {
            log.Printf("Failed to create MCP bridge: %v", err)
            return nil
        }

        server := mcp.NewServer(&mcp.Implementation{
            Name:    "portainer-mcp-wrapper",
            Version: "1.0.0",
        }, nil)

        mcpBridge.AttachToServer(server)
        return server
    }

    // Create HTTP handler with SSE support
    handler := mcp.NewStreamableHTTPHandler(getServer, &mcp.StreamableHTTPOptions{
        SessionTimeout: 5 * time.Minute,
        Stateless:      false,
    })

    // Wrap with authentication middleware
    authHandler := auth.NewAuthMiddleware(cfg.MCPAccessToken)(handler)

    // Create HTTP mux with health check
    mux := http.NewServeMux()
    mux.Handle("/", authHandler)
    mux.HandleFunc("/health", healthCheckHandler)

    // Create HTTP server
    srv := &http.Server{
        Addr:         fmt.Sprintf(":%d", cfg.MCPPort),
        Handler:      mux,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Graceful shutdown handling
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
        <-sigChan

        log.Println("Shutdown signal received, stopping server...")
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        if err := srv.Shutdown(ctx); err != nil {
            log.Printf("Server shutdown error: %v", err)
        }
    }()

    // Start server
    log.Printf("Starting Portainer MCP wrapper on :%d", cfg.MCPPort)
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("Server failed: %v", err)
    }

    log.Println("Server stopped")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, `{"status":"healthy","service":"portainer-mcp-wrapper"}`)
}
```

## Configuration Package

**File**: `internal/config/config.go`

```go
package config

import (
    "fmt"
    "os"
    "strconv"
)

type Config struct {
    PortainerURL        string
    PortainerAPIToken   string
    MCPAccessToken      string
    MCPPort             int
    MCPToolsFile        string
    DisableVersionCheck bool
    ReadOnlyMode        bool
    MCPBinaryPath       string
}

func LoadConfig() (*Config, error) {
    cfg := &Config{
        PortainerURL:        getEnvOrDefault("PORTAINER_URL", "http://portainer:9000"),
        PortainerAPIToken:   os.Getenv("PORTAINER_API_TOKEN"),
        MCPAccessToken:      os.Getenv("MCP_ACCESS_TOKEN"),
        MCPPort:             getEnvAsIntOrDefault("MCP_PORT", 8080),
        MCPToolsFile:        os.Getenv("MCP_TOOLS_FILE"),
        DisableVersionCheck: getEnvAsBoolOrDefault("DISABLE_VERSION_CHECK", false),
        ReadOnlyMode:        getEnvAsBoolOrDefault("READ_ONLY_MODE", false),
        MCPBinaryPath:       getEnvOrDefault("MCP_BINARY_PATH", "/app/portainer-mcp"),
    }

    // Validate required fields
    if cfg.PortainerAPIToken == "" {
        return nil, fmt.Errorf("PORTAINER_API_TOKEN is required")
    }
    if cfg.MCPAccessToken == "" {
        return nil, fmt.Errorf("MCP_ACCESS_TOKEN is required")
    }

    return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultVal
}

func getEnvAsIntOrDefault(key string, defaultVal int) int {
    if val := os.Getenv(key); val != "" {
        if intVal, err := strconv.Atoi(val); err == nil {
            return intVal
        }
    }
    return defaultVal
}

func getEnvAsBoolOrDefault(key string, defaultVal bool) bool {
    if val := os.Getenv(key); val != "" {
        if boolVal, err := strconv.ParseBool(val); err == nil {
            return boolVal
        }
    }
    return defaultVal
}
```

## Authentication Middleware

**File**: `internal/auth/auth.go`

```go
package auth

import (
    "net/http"
    "strings"
)

// NewAuthMiddleware creates middleware that validates Bearer token
func NewAuthMiddleware(expectedToken string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract Authorization header
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
                return
            }

            // Check Bearer token format
            parts := strings.SplitN(authHeader, " ", 2)
            if len(parts) != 2 || parts[0] != "Bearer" {
                http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
                return
            }

            // Validate token
            if parts[1] != expectedToken {
                http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
                return
            }

            // Token valid, proceed
            next.ServeHTTP(w, r)
        })
    }
}
```

## Bridge Implementation

**File**: `internal/bridge/bridge.go`

```go
package bridge

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "os/exec"
    "sync"

    "github.com/modelcontextprotocol/go-sdk/mcp"
    "portainer-mcp-wrapper/internal/config"
)

type PortainerMCPBridge struct {
    cmd    *exec.Cmd
    stdin  io.WriteCloser
    stdout io.ReadCloser
    stderr io.ReadCloser
    mu     sync.Mutex
    cfg    *config.Config
}

func NewPortainerMCPBridge(cfg *config.Config) (*PortainerMCPBridge, error) {
    // Build command arguments
    args := []string{
        "--server", cfg.PortainerURL,
        "--api-token", cfg.PortainerAPIToken,
    }

    if cfg.MCPToolsFile != "" {
        args = append(args, "--tools-file", cfg.MCPToolsFile)
    }
    if cfg.DisableVersionCheck {
        args = append(args, "--disable-version-check")
    }
    if cfg.ReadOnlyMode {
        args = append(args, "--read-only")
    }

    // Create command
    cmd := exec.Command(cfg.MCPBinaryPath, args...)

    // Get pipes
    stdin, err := cmd.StdinPipe()
    if err != nil {
        return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
    }

    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
    }

    stderr, err := cmd.StderrPipe()
    if err != nil {
        return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
    }

    // Start the process
    if err := cmd.Start(); err != nil {
        return nil, fmt.Errorf("failed to start portainer-mcp: %w", err)
    }

    bridge := &PortainerMCPBridge{
        cmd:    cmd,
        stdin:  stdin,
        stdout: stdout,
        stderr: stderr,
        cfg:    cfg,
    }

    // Start logging stderr
    go bridge.logStderr()

    return bridge, nil
}

func (b *PortainerMCPBridge) logStderr() {
    scanner := bufio.NewScanner(b.stderr)
    for scanner.Scan() {
        log.Printf("[portainer-mcp] %s", scanner.Text())
    }
}

// SendMessage sends a JSON-RPC message to the subprocess
func (b *PortainerMCPBridge) SendMessage(msg interface{}) error {
    b.mu.Lock()
    defer b.mu.Unlock()

    data, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }

    // Write message with newline delimiter
    if _, err := b.stdin.Write(append(data, '\n')); err != nil {
        return fmt.Errorf("failed to write message: %w", err)
    }

    return nil
}

// ReadMessage reads a JSON-RPC message from the subprocess
func (b *PortainerMCPBridge) ReadMessage() (map[string]interface{}, error) {
    scanner := bufio.NewScanner(b.stdout)
    if !scanner.Scan() {
        if err := scanner.Err(); err != nil {
            return nil, fmt.Errorf("failed to read message: %w", err)
        }
        return nil, io.EOF
    }

    var msg map[string]interface{}
    if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
        return nil, fmt.Errorf("failed to unmarshal message: %w", err)
    }

    return msg, nil
}

// AttachToServer configures the MCP server to proxy through this bridge
func (b *PortainerMCPBridge) AttachToServer(server *mcp.Server) {
    // Implementation depends on MCP Go SDK's internal APIs
    // May require custom transport implementation
}

// Close terminates the subprocess
func (b *PortainerMCPBridge) Close() error {
    b.stdin.Close()
    return b.cmd.Wait()
}
```

**Note**: The bridge implementation above is simplified. Consider using `mcp.NewCommandTransport` directly:

```go
cmd := exec.Command(cfg.MCPBinaryPath, args...)
transport, err := mcp.NewCommandTransport(cmd)
if err != nil {
    return nil, err
}
// Use transport with client/server
```

## Docker Configuration Files

### Dockerfile (Multi-stage Build)

```dockerfile
# syntax=docker/dockerfile:1

# Stage 1: Build the Go wrapper
FROM golang:1.23-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o wrapper \
    ./cmd/wrapper

# Stage 2: Download Portainer MCP binary
FROM alpine:latest AS downloader

ARG PORTAINER_MCP_VERSION=v0.6.0
ARG TARGETARCH

WORKDIR /download

RUN apk add --no-cache curl tar

RUN curl -fsSL \
    "https://github.com/portainer/portainer-mcp/releases/download/${PORTAINER_MCP_VERSION}/portainer-mcp-linux-${TARGETARCH}.tar.gz" \
    -o portainer-mcp.tar.gz && \
    tar -xzf portainer-mcp.tar.gz && \
    chmod +x portainer-mcp

# Stage 3: Final runtime image
FROM alpine:latest

RUN apk add --no-cache \
    ca-certificates \
    wget \
    && addgroup -g 1000 mcpuser \
    && adduser -D -u 1000 -G mcpuser mcpuser

WORKDIR /app

COPY --from=builder /build/wrapper /app/wrapper
COPY --from=downloader /download/portainer-mcp /app/portainer-mcp

RUN chown -R mcpuser:mcpuser /app

RUN mkdir -p /config && chown mcpuser:mcpuser /config

USER mcpuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --spider -q http://localhost:8080/health || exit 1

ENV MCP_PORT=8080 \
    MCP_BINARY_PATH=/app/portainer-mcp

ENTRYPOINT ["/app/wrapper"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  portainer-mcp:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        PORTAINER_MCP_VERSION: v0.6.0
    image: portainer-mcp-wrapper:latest
    container_name: portainer-mcp-wrapper
    restart: unless-stopped

    ports:
      - "${MCP_PORT:-8080}:8080"

    environment:
      PORTAINER_URL: ${PORTAINER_URL:-http://portainer:9000}
      PORTAINER_API_TOKEN: ${PORTAINER_API_TOKEN:?PORTAINER_API_TOKEN is required}
      MCP_ACCESS_TOKEN: ${MCP_ACCESS_TOKEN:?MCP_ACCESS_TOKEN is required}
      MCP_PORT: ${MCP_PORT:-8080}
      MCP_TOOLS_FILE: ${MCP_TOOLS_FILE:-}
      DISABLE_VERSION_CHECK: ${DISABLE_VERSION_CHECK:-false}
      READ_ONLY_MODE: ${READ_ONLY_MODE:-false}

    volumes:
      - ./config:/config:ro

    networks:
      - portainer_network

    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s

    security_opt:
      - no-new-privileges:true

    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
        reservations:
          memory: 128M

networks:
  portainer_network:
    external: true
```

### .env.example

```bash
# Portainer Configuration
PORTAINER_URL=http://192.168.1.100:9000
PORTAINER_API_TOKEN=ptr_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# MCP Wrapper Configuration
MCP_ACCESS_TOKEN=your-secure-random-token-here-min-32-chars
MCP_PORT=8080

# Optional: Custom tools configuration file
# MCP_TOOLS_FILE=/config/tools.yaml

# Optional: Disable Portainer version check
# DISABLE_VERSION_CHECK=false

# Optional: Enable read-only mode
# READ_ONLY_MODE=false
```

### go.mod

```go
module portainer-mcp-wrapper

go 1.23

require (
    github.com/modelcontextprotocol/go-sdk v1.0.0
)
```

## Build and Run Instructions

```bash
# 1. Clone/create the project
mkdir portainer-mcp-docker-wrapper
cd portainer-mcp-docker-wrapper

# 2. Copy .env.example to .env and configure
cp .env.example .env
# Edit .env with your values

# 3. Generate secure MCP access token
openssl rand -base64 32

# 4. Build the image
docker-compose build

# 5. Start the service
docker-compose up -d

# 6. Check logs
docker-compose logs -f

# 7. Test health endpoint
curl http://localhost:8080/health

# 8. Test with authentication
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/health
```
