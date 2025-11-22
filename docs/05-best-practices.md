# Best Practices and Learning

## Security Architecture Insights

### Two-Token Model Rationale

**Why separate MCP access token from Portainer API token?**

1. **Defense in Depth**: Even if MCP endpoint is compromised, attacker still needs valid Portainer credentials
2. **Access Control**: Can revoke MCP access without changing Portainer tokens
3. **Audit Trail**: Separate authentication layers enable better logging
4. **Principle of Least Privilege**: MCP wrapper doesn't need to know about Portainer auth logic

### Authentication Placement

Place authentication at the HTTP layer (middleware) rather than inside MCP protocol:
- Simpler to implement
- Standard HTTP patterns (Bearer tokens, headers)
- Compatible with reverse proxies
- Easier to test and debug

## Docker Optimization Techniques

### Multi-Stage Build Benefits

```dockerfile
FROM golang:1.23-alpine AS builder  # ~300MB
FROM alpine:latest                    # ~5MB runtime
```

**Results**:
- Builder stage: ~300MB (discarded)
- Final image: ~20-30MB
- Build time: ~2-3 min (with cache: ~10s)

### Security Hardening Checklist

✅ Non-root user (`USER mcpuser`)
✅ Read-only config volumes (`:ro`)
✅ No new privileges (`no-new-privileges:true`)
✅ Minimal base image (Alpine)
✅ Health checks
✅ Resource limits
✅ Secrets via environment (not in image)

## Network Architecture Decisions

### Why Not Direct Portainer Connection?

You might wonder: "Why not have Claude Code talk directly to Portainer's API?"

**Answer**: MCP provides valuable abstraction:
- **Natural language interface**: MCP tools have descriptions for LLM
- **Structured operations**: Tools define clear inputs/outputs
- **Safety**: MCP can implement read-only mode, validation
- **Convenience**: MCP handles Portainer API complexity
- **Consistency**: MCP provides standard protocol for all tools

### Reverse Proxy Consideration

For production deployments, Caddy is recommended over nginx because:
- Automatic HTTPS with Let's Encrypt
- Simple configuration syntax
- Built-in security headers
- No need for certbot or manual certificate management

## Development Workflow Recommendations

### Iteration Strategy

1. **Start simple**: Basic HTTP wrapper without auth
2. **Add security**: Authentication middleware
3. **Add production features**: HTTPS, monitoring, health checks
4. **Optimize**: Multi-stage build, resource limits

### Testing Approach

```bash
# Test 1: Health check (no auth)
curl http://localhost:8080/health

# Test 2: MCP endpoint (with auth)
curl -H "Authorization: Bearer TOKEN" http://localhost:8080/

# Test 3: Integration test with Claude Code
# Configure Claude Code and attempt tool listing
```

## Common Pitfalls to Avoid

### ❌ Don't: Embed secrets in Docker image

```dockerfile
# BAD - hardcoded token
ENV PORTAINER_API_TOKEN=ptr_secret123
```

### ✅ Do: Use environment variables at runtime

```yaml
# GOOD - from environment/secrets
environment:
  PORTAINER_API_TOKEN: ${PORTAINER_API_TOKEN}
```

### ❌ Don't: Run as root user

```dockerfile
# BAD - runs as root
ENTRYPOINT ["/app/wrapper"]
```

### ✅ Do: Use non-root user

```dockerfile
# GOOD - runs as mcpuser
USER mcpuser
ENTRYPOINT ["/app/wrapper"]
```

### ❌ Don't: Expose ports unnecessarily

```yaml
# BAD - exposing internal port
ports:
  - "8080:8080"  # Direct exposure
```

### ✅ Do: Use reverse proxy for production

```yaml
# GOOD - behind Caddy with HTTPS
services:
  caddy:
    ports:
      - "443:443"
  portainer-mcp:
    expose:
      - "8080"  # Internal only
```

## Performance Expectations

### Latency Budget

- Network RTT (LAN): 1-5ms
- Wrapper overhead: 1-2ms
- Portainer MCP processing: 10-50ms
- Portainer API call: 20-100ms
- **Total**: 30-160ms per operation

This is acceptable for interactive use with Claude Code.

### Memory Usage

- Wrapper process: 10-20MB
- Portainer MCP subprocess: 30-50MB
- Alpine container overhead: 5MB
- **Total**: ~50-75MB per container

### Concurrency

The wrapper can handle multiple concurrent Claude Code sessions:
- Each session gets independent MCP protocol handling
- Can share single Portainer MCP subprocess OR spawn multiple
- Recommend: Start with single subprocess, scale if needed

## Alternative Approaches Considered

### Approach 1: SSH Tunnel

```bash
ssh -L 8080:localhost:8080 docker-host
```

**Pros**: Simple, secure
**Cons**: Requires SSH access, manual tunnel management, not always available on Windows

### Approach 2: VPN

**Pros**: Full network access, transparent
**Cons**: Complex setup, overkill for single service

### Approach 3: Direct Modification of Portainer MCP

Fork the repo and add HTTP transport directly.

**Pros**: Single binary, optimal performance
**Cons**: Maintenance burden, must sync with upstream, requires Go expertise

### Selected: HTTP Wrapper (Best Balance)

**Pros**: Clean separation, maintainable, no forking, leverages SDK
**Cons**: Slightly more complex deployment, minimal overhead

## Future Enhancement Ideas

### Nice-to-Have Features

1. **Metrics endpoint**: Prometheus metrics for monitoring
2. **Multiple Portainer instances**: Load balance across multiple Portainers
3. **Caching layer**: Cache frequently-accessed data
4. **WebSocket support**: Alternative to SSE for some clients
5. **Admin UI**: Web interface for configuration and monitoring

### Advanced Security

1. **mTLS**: Mutual TLS for client authentication
2. **OAuth2**: Integration with identity providers
3. **Rate limiting**: Prevent abuse
4. **Audit logging**: Comprehensive operation logging

## MCP Transport Architecture Learnings

### Transport Compatibility Matrix

**Stdio Transport**:
- Use case: Local subprocess execution
- Client launches server as child process
- Communication: stdin/stdout with newline-delimited JSON
- Best for: Desktop apps, CLI tools, single-user scenarios
- **Limitation**: Cannot be used remotely over network

**HTTP/SSE Transport**:
- Use case: Remote network access, web services
- Server runs independently as HTTP service
- Communication: HTTP POST for requests, SSE for server→client messages
- Best for: Multi-user, cloud deployments, cross-machine scenarios
- **Advantage**: Works over standard network protocols

**Key Insight**: MCP servers using stdio cannot be directly accessed remotely. A bridge/wrapper is required to translate between transports.

## MCP Go SDK Capabilities

### Built-in Transport Support

The MCP Go SDK provides excellent transport abstractions:

1. **`StdioTransport`**: For stdin/stdout communication
2. **`CommandTransport`**: Manages subprocess with stdio
3. **`StreamableHTTPHandler`**: HTTP/SSE server implementation
4. **`StreamableClientTransport`**: HTTP/SSE client implementation

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

## Production Deployment Recommendations

### Using Caddy for HTTPS

**docker-compose.override.yml**:

```yaml
version: '3.8'

services:
  caddy:
    image: caddy:2-alpine
    container_name: portainer-mcp-caddy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
    networks:
      - portainer_network
    depends_on:
      - portainer-mcp

  portainer-mcp:
    ports: []
    expose:
      - "8080"

volumes:
  caddy_data:
  caddy_config:
```

**Caddyfile**:

```
mcp.yourdomain.com {
    reverse_proxy portainer-mcp:8080
}
```

### Security Checklist for Production

- [ ] Generate strong random tokens (32+ characters)
- [ ] Enable HTTPS via Caddy or another reverse proxy
- [ ] Configure firewall rules to restrict access
- [ ] Use Docker secrets for sensitive data
- [ ] Enable health checks and monitoring
- [ ] Set resource limits on containers
- [ ] Run as non-root user
- [ ] Use read-only volumes where possible
- [ ] Rotate tokens periodically
- [ ] Enable audit logging
- [ ] Regular security updates for base images
- [ ] Network isolation via dedicated Docker network

## Monitoring and Maintenance

### Health Checks

```bash
# Check container health
docker-compose ps

# Check logs
docker-compose logs -f portainer-mcp

# Test endpoint
curl http://localhost:8080/health
```

### Common Issues and Solutions

**Issue**: Container fails to start
- Check environment variables in `.env`
- Verify Portainer URL is accessible
- Check token validity

**Issue**: Authentication failures
- Verify MCP_ACCESS_TOKEN matches client configuration
- Check Bearer token format in requests

**Issue**: Connection timeout
- Verify network connectivity between machines
- Check firewall rules
- Ensure port 8080 is not blocked

**Issue**: High memory usage
- Check for zombie processes
- Consider process restart policies
- Monitor concurrent session count

### Update Strategy

1. Watch Portainer MCP releases: https://github.com/portainer/portainer-mcp/releases
2. Update `PORTAINER_MCP_VERSION` in Dockerfile
3. Test in development environment
4. Rebuild image: `docker-compose build`
5. Rolling update: `docker-compose up -d`
6. Verify functionality
7. Monitor logs for errors
