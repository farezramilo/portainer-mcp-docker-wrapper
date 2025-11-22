# Implementation Plan: Portainer MCP Docker Wrapper

## Phase 1: Architecture & Design

### 1.1 Transport Bridge Design

**Objective**: Design HTTP/SSE wrapper for stdio-based Portainer MCP

**Approach Options**:

#### Option A: Go HTTP Wrapper (Recommended)
- Create Go service using `mcp.NewStreamableHTTPHandler`
- Launch Portainer MCP binary as subprocess
- Bridge stdio ↔ HTTP/SSE using MCP Go SDK
- **Pros**: Native Go SDK support, efficient, maintainable
- **Cons**: Requires Go development

#### Option B: Node.js Bridge
- Use `@modelcontextprotocol/sdk` TypeScript/Node
- Spawn Portainer MCP as child process
- Bridge using MCP SDK transports
- **Pros**: Familiar ecosystem, good MCP SDK support
- **Cons**: Additional runtime dependency

#### Option C: Direct Fork & Modify
- Fork Portainer MCP repository
- Add HTTP/SSE transport support directly
- **Pros**: Single binary, optimal performance
- **Cons**: Maintenance burden, upstream sync issues

**Selected**: Option A (Go HTTP Wrapper)

### 1.2 Security Architecture

**Authentication Layer**:
- API key-based authentication for MCP HTTP endpoint
- TLS/HTTPS for encrypted transport
- Environment variable for MCP access token
- Separate from Portainer API token

**Network Security**:
- Expose only necessary port (e.g., 8080)
- Optional: IP whitelist/firewall rules
- Optional: Reverse proxy with authentication (Caddy/Traefik)

### 1.3 Configuration Strategy

**Environment Variables**:
```
PORTAINER_URL=http://portainer:9000
PORTAINER_API_TOKEN=ptr_xxx...
MCP_ACCESS_TOKEN=mcp_access_secret
MCP_PORT=8080
MCP_TOOLS_FILE=/config/tools.yaml (optional)
DISABLE_VERSION_CHECK=false
READ_ONLY_MODE=false
```

**Docker Volumes**:
- `/config`: Optional tools.yaml configuration
- Secrets: Use Docker secrets for tokens in production

---

## Phase 2: Development

### 2.1 Create Go HTTP Wrapper Service

**File Structure**:
```
portainer-mcp-docker-wrapper/
├── cmd/
│   └── wrapper/
│       └── main.go          # HTTP wrapper entrypoint
├── internal/
│   ├── auth/
│   │   └── auth.go          # API key authentication
│   ├── bridge/
│   │   └── bridge.go        # Stdio ↔ HTTP bridge
│   └── config/
│       └── config.go        # Environment config
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

**Key Components**:

**2.1.1 Main Wrapper** (`cmd/wrapper/main.go`):
- Load configuration from environment
- Download/verify Portainer MCP binary
- Create MCP server with StreamableHTTPHandler
- Implement authentication middleware
- Start HTTP server on configured port
- Health check endpoint

**2.1.2 Bridge Logic** (`internal/bridge/bridge.go`):
- Spawn Portainer MCP as subprocess with args
- Connect to subprocess stdin/stdout
- Bridge MCP messages between HTTP/SSE and stdio
- Handle process lifecycle (restart on crash)
- Logging and error handling

**2.1.3 Authentication** (`internal/auth/auth.go`):
- Middleware to check Authorization header
- Bearer token validation against MCP_ACCESS_TOKEN
- 401 Unauthorized for invalid tokens
- Optional: Rate limiting

### 2.2 Dockerfile Creation

**Multi-stage Dockerfile**:
- Stage 1: Build wrapper from Go source
- Stage 2: Download Portainer MCP binary for target architecture
- Stage 3: Create minimal runtime image with Alpine

**Key Features**:
- Multi-stage build for minimal image size
- Support for multiple architectures (amd64, arm64)
- Non-root user execution
- Health check support

### 2.3 Docker Compose Configuration

**docker-compose.yml**:
- Service definition with environment variables
- Port mapping (8080:8080)
- Volume mounts for config
- Network configuration
- Health checks
- Resource limits

**Environment File** (`.env.example`):
- Template for configuration
- Comments explaining each variable
- Security warnings for tokens

---

## Phase 3: Client Configuration

### 3.1 Claude Code Configuration (Windows)

**MCP Settings Location**: `%APPDATA%\Claude\claude_desktop_config.json`

**Configuration**:
```json
{
  "mcpServers": {
    "portainer": {
      "url": "http://<docker-host-ip>:8080",
      "transport": "sse",
      "headers": {
        "Authorization": "Bearer your-secure-mcp-access-token"
      }
    }
  }
}
```

### 3.2 Network Configuration

**Docker Host Requirements**:
- Port 8080 accessible from Windows machine
- Firewall rules allow incoming on 8080
- Optional: Use nginx/Caddy reverse proxy with HTTPS

**Network Scenarios**:

**Scenario 1: Same Local Network**
- Direct IP connection: `http://192.168.1.x:8080`
- No additional configuration needed

**Scenario 2: Different Networks**
- VPN tunnel between machines
- Port forwarding on router
- SSH tunnel: `ssh -L 8080:localhost:8080 user@docker-host`

**Scenario 3: Internet Exposure (Production)**
- Domain name + HTTPS via Let's Encrypt
- Reverse proxy (Caddy recommended)
- Strong authentication tokens
- Rate limiting and monitoring

---

## Phase 4: Security Hardening

### 4.1 Authentication Enhancements
- Generate strong random MCP access tokens (32+ characters)
- Rotate tokens periodically
- Log authentication failures
- Optional: Implement JWT with expiration

### 4.2 TLS/HTTPS Setup

**Option A: Caddy Reverse Proxy** (Recommended)
- Automatic HTTPS with Let's Encrypt
- Simple configuration via Caddyfile
- Built-in security headers

### 4.3 Docker Security
- Run as non-root user
- Read-only filesystem where possible
- Drop unnecessary capabilities
- Use Docker secrets for sensitive data
- Scan images for vulnerabilities

### 4.4 Network Isolation
- Use dedicated Docker network
- Firewall rules limiting access
- Optional: VPN-only access

---

## Phase 5: Testing & Validation

### 5.1 Local Testing
1. Build Docker image
2. Start container with test configuration
3. Verify HTTP endpoint responds
4. Test MCP protocol handshake
5. Execute sample Portainer commands

### 5.2 Integration Testing
1. Configure Claude Code on Windows
2. Attempt connection to remote MCP server
3. List Portainer environments
4. Create/update test stack
5. Verify read-only mode (if enabled)

### 5.3 Security Testing
1. Test without authentication (should fail)
2. Test with invalid tokens (should return 401)
3. Verify TLS certificate (if using HTTPS)
4. Check for exposed secrets in logs
5. Network scanning for open ports

### 5.4 Performance Testing
1. Multiple concurrent requests
2. Large stack deployments
3. Memory and CPU usage monitoring
4. Connection stability over time

---

## Phase 6: Documentation

### 6.1 README.md
- Project overview and architecture
- Prerequisites
- Quick start guide
- Configuration reference
- Troubleshooting section
- Security best practices

### 6.2 Setup Guide
- Step-by-step Docker deployment
- Windows Claude Code configuration
- Network setup for various scenarios
- Token generation and management

### 6.3 API Documentation
- Available MCP tools/endpoints
- Authentication format
- Error codes and handling
- Example requests/responses

---

## Phase 7: Deployment & Maintenance

### 7.1 Initial Deployment
1. Clone repository on Docker host
2. Copy `.env.example` to `.env` and configure
3. Generate secure tokens
4. Build and start: `docker-compose up -d`
5. Verify health check passes
6. Test connectivity from Windows machine

### 7.2 Monitoring
- Container logs: `docker-compose logs -f`
- Health check status
- Resource usage metrics
- Failed authentication attempts

### 7.3 Updates
- Watch Portainer MCP releases
- Update PORTAINER_MCP_VERSION in Dockerfile
- Rebuild image: `docker-compose build`
- Rolling restart: `docker-compose up -d`

### 7.4 Backup & Recovery
- Backup `.env` file (securely!)
- Backup custom tools.yaml
- Document configuration settings
- Container recreation procedure

---

## Timeline Estimate

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Phase 1: Architecture & Design | 2-4 hours | - |
| Phase 2: Development | 8-12 hours | Phase 1 |
| Phase 3: Client Configuration | 1-2 hours | Phase 2 |
| Phase 4: Security Hardening | 2-4 hours | Phase 2 |
| Phase 5: Testing & Validation | 3-5 hours | Phase 2, 3, 4 |
| Phase 6: Documentation | 2-3 hours | All phases |
| Phase 7: Deployment | 1-2 hours | Phase 5, 6 |
| **Total** | **19-32 hours** | - |

---

## Risk Assessment

### High Priority Risks
1. **MCP Protocol Changes**: Portainer MCP updates may break wrapper
   - **Mitigation**: Pin versions, test updates in staging

2. **Security Vulnerabilities**: Exposed MCP server could be compromised
   - **Mitigation**: Strong auth, HTTPS, network isolation, monitoring

3. **Network Connectivity**: Firewall/routing issues between machines
   - **Mitigation**: Multiple transport options, clear network docs

### Medium Priority Risks
1. **Performance Bottlenecks**: Wrapper adds latency
   - **Mitigation**: Efficient bridging, connection pooling, monitoring

2. **Portainer Version Compatibility**: Version check may fail
   - **Mitigation**: Disable version check option, document compatibility

### Low Priority Risks
1. **Docker Host Downtime**: Container unavailable
   - **Mitigation**: Restart policies, monitoring, documentation

---

## Success Metrics

✅ Docker container builds successfully
✅ HTTP/SSE endpoint accessible remotely
✅ Claude Code connects and authenticates
✅ All Portainer MCP tools functional
✅ Secure token-based authentication working
✅ Documentation complete and clear
✅ Zero exposed secrets or credentials
✅ Health checks passing consistently

---

## Next Steps

1. **Review & Approve Plan**: Confirm architecture decisions
2. **Setup Development Environment**: Install Go, Docker, dependencies
3. **Create Repository**: Initialize Git repo with structure
4. **Begin Phase 2**: Start wrapper development
5. **Iterative Testing**: Test each component as built
