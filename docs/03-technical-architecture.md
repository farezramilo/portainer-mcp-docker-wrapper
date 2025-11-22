# Technical Architecture: MCP Transport Bridge

## Problem Statement

The Portainer MCP server uses stdio transport (stdin/stdout), designed for subprocess execution. Claude Code on a remote Windows machine needs HTTP/SSE transport for network communication.

## Solution Architecture

### High-Level Flow

```
Windows Machine (Claude Code)
    ↓ HTTP/SSE Request
Docker Host (Wrapper Container)
    ↓ Stdio Bridge
Portainer MCP Binary (Subprocess)
    ↓ HTTP API
Portainer Instance
```

## Transport Bridge Implementation

### MCP Go SDK Transport Types

**Client Side (Claude Code)**:
- `StreamableClientTransport`: Connects to HTTP/SSE MCP server
- Handles both SSE streaming and JSON polling modes
- Configurable with custom HTTP client for timeouts/auth

**Server Side (Wrapper)**:
- `NewStreamableHTTPHandler`: Creates HTTP handler for MCP sessions
- Supports multiple concurrent sessions
- Provides both SSE streaming and JSON response modes
- Configurable session timeout and stateless mode

### Bridge Architecture

**Approach**: Dual-Transport Bridge

1. **External Interface**: HTTP/SSE via `StreamableHTTPHandler`
2. **Internal Interface**: Launch Portainer MCP as subprocess, communicate via stdio
3. **Bridge Layer**: Translate between transports

### Key Technical Decisions

**Why Not Fork Portainer MCP?**
- Maintenance burden: must sync with upstream changes
- Portainer MCP receives active updates
- Wrapper approach is modular and maintainable

**Why Go Over Node.js?**
- Portainer MCP is written in Go
- MCP Go SDK has excellent HTTP/SSE support
- Single compiled binary (no runtime dependencies)
- Better performance for stdio bridging
- Easier deployment in Alpine container

**Why StreamableHTTPHandler?**
- Official MCP Go SDK transport
- Supports both SSE streaming and JSON modes
- Built-in session management
- Handles connection lifecycle properly
- Compatible with Claude Code's HTTP transport

## Authentication Strategy

### Two-Layer Security Model

**Layer 1: MCP Wrapper Authentication**
- Custom Bearer token for HTTP endpoint
- Validates before proxying to Portainer MCP
- Prevents unauthorized MCP access

**Layer 2: Portainer API Authentication**
- Portainer API token passed to MCP subprocess
- MCP server validates with Portainer
- Standard Portainer security model

### Token Flow

```
Claude Code Request
    → Authorization: Bearer <MCP_ACCESS_TOKEN>
        → Wrapper validates token
            → If valid: spawn/proxy to Portainer MCP
                → Portainer MCP uses <PORTAINER_API_TOKEN>
                    → Portainer validates API token
                        → Execute operation
```

## Network Architecture

### Port Mapping

- External: `<host-ip>:8080` (HTTP/SSE endpoint)
- Internal: Container port 8080
- Portainer: Configurable via `PORTAINER_URL` env var

### Network Modes

**Development**: Bridge network, expose port
```yaml
ports:
  - "8080:8080"
networks:
  - bridge
```

**Production**: Custom network with Portainer
```yaml
networks:
  portainer_network:
    external: true
```

### Security Considerations

- Bind to `0.0.0.0:8080` in container (accessible externally)
- Host firewall controls external access
- Optional: Reverse proxy with automatic HTTPS (Caddy)
- Optional: VPN-only access for sensitive deployments

## Error Handling

### Subprocess Failure

- Monitor Portainer MCP process health
- Automatic restart on crash
- Return 503 Service Unavailable during downtime
- Log errors to stderr

### Network Failures

- Connection timeout handling
- Graceful degradation
- Proper HTTP status codes
- Client retry logic in Claude Code

### Authentication Failures

- 401 Unauthorized for invalid MCP token
- 403 Forbidden for Portainer API issues
- Clear error messages in response

## Performance Considerations

### Latency

- Additional hop: ~5-10ms overhead
- Acceptable for interactive use
- SSE keeps connection alive (no reconnection overhead)

### Concurrency

- StreamableHTTPHandler supports multiple sessions
- Each session gets dedicated subprocess (optional) OR
- Shared subprocess with request multiplexing

### Resource Usage

- Wrapper: ~10-20MB memory
- Portainer MCP subprocess: ~30-50MB memory
- Total container: ~100MB runtime memory

## Scalability Options

### Single Instance (Recommended)

- One wrapper container
- One or more Portainer MCP subprocesses
- Sufficient for individual/small team use

### Multi-Instance (Advanced)

- Multiple wrapper containers
- Load balancer (HAProxy/nginx)
- For high-availability scenarios

## Transport Compatibility Matrix

### Stdio Transport
- **Use case**: Local subprocess execution
- **Client launches**: Server as child process
- **Communication**: stdin/stdout with newline-delimited JSON
- **Best for**: Desktop apps, CLI tools, single-user scenarios
- **Limitation**: Cannot be used remotely over network

### HTTP/SSE Transport
- **Use case**: Remote network access, web services
- **Server runs**: Independently as HTTP service
- **Communication**: HTTP POST for requests, SSE for server→client messages
- **Best for**: Multi-user, cloud deployments, cross-machine scenarios
- **Advantage**: Works over standard network protocols

**Key Insight**: MCP servers using stdio cannot be directly accessed remotely. A bridge/wrapper is required to translate between transports.

## MCP Go SDK Capabilities

### Built-in Transport Support

The MCP Go SDK provides excellent transport abstractions:

1. `StdioTransport`: For stdin/stdout communication
2. `CommandTransport`: Manages subprocess with stdio
3. `StreamableHTTPHandler`: HTTP/SSE server implementation
4. `StreamableClientTransport`: HTTP/SSE client implementation

### StreamableHTTPHandler Features

```go
mcp.NewStreamableHTTPHandler(getServer, &mcp.StreamableHTTPOptions{
    SessionTimeout: 5 * time.Minute,
    Stateless:      false,
})
```

**Key features**:
- Supports both SSE streaming AND standard JSON HTTP
- Automatic session management
- Configurable timeout
- Multiple concurrent sessions
- Standard Go `http.Handler` interface (easy to wrap with middleware)

### Why This Matters

The wrapper can leverage `StreamableHTTPHandler` to expose MCP over HTTP while using `CommandTransport` internally to manage the Portainer MCP subprocess. This is the most elegant solution.
