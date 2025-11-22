# Phase 1 Prerequisites Checklist

Before starting Phase 1 (Architecture & Design) of the Portainer MCP Docker Wrapper project, ensure you have the following prerequisites in place.

## Development Environment

### Required Software

- [ ] **Go 1.23+** installed
  - Verify: `go version`
  - Download: https://go.dev/dl/

- [ ] **Docker** installed and running
  - Verify: `docker --version` and `docker ps`
  - Download: https://www.docker.com/products/docker-desktop

- [ ] **Docker Compose** installed
  - Verify: `docker-compose --version`
  - Usually included with Docker Desktop

- [ ] **Git** for version control
  - Verify: `git --version`

- [ ] **Text editor/IDE** (VS Code, GoLand, etc.)

### Optional Tools

- [ ] **curl** or **wget** for testing HTTP endpoints
- [ ] **jq** for JSON formatting
- [ ] **make** for build automation (optional)
- [ ] **openssl** for token generation

## Access Requirements

### Portainer Instance

- [ ] Running Portainer instance (version 2.31.2+ recommended)
  - URL/IP address noted
  - Network accessible from Docker host

- [ ] Portainer API token generated
  - Go to Portainer → User settings → Access tokens
  - Create new token with appropriate permissions
  - Save token securely (starts with `ptr_`)

- [ ] Test Portainer API access
  ```bash
  curl -H "X-API-Key: YOUR_TOKEN" http://PORTAINER_URL/api/endpoints
  ```

### Network Setup

- [ ] Docker host identified
  - Linux machine or VM with Docker installed
  - Network accessible from Windows machine

- [ ] Network connectivity verified
  - Can ping Docker host from Windows
  - Port 8080 available on Docker host
  - Firewall rules documented

- [ ] Windows machine preparation
  - Claude Code installed
  - Can access Docker host network
  - Admin/config access to `%APPDATA%\Claude\claude_desktop_config.json`

## Knowledge Prerequisites

### Technical Understanding

- [ ] Basic understanding of:
  - Docker and containerization
  - Docker Compose
  - HTTP/REST APIs
  - Authentication (Bearer tokens)
  - Network basics (ports, firewalls, IP addresses)

- [ ] Familiarity with:
  - Go programming (basic level)
  - MCP protocol concepts
  - Portainer functionality

### Documentation Access

- [ ] Read project brief ([docs/01-project-brief.md](docs/01-project-brief.md))
- [ ] Reviewed implementation plan ([docs/02-implementation-plan.md](docs/02-implementation-plan.md))
- [ ] Understand technical architecture ([docs/03-technical-architecture.md](docs/03-technical-architecture.md))

- [ ] Bookmarked key resources:
  - MCP Go SDK: https://github.com/modelcontextprotocol/go-sdk
  - Portainer MCP: https://github.com/portainer/portainer-mcp
  - MCP Specification: https://modelcontextprotocol.io/
  - Portainer API Docs: https://docs.portainer.io/api/

## Security Preparation

### Token Generation

- [ ] Method to generate secure random tokens
  ```bash
  openssl rand -base64 32
  ```

- [ ] Secure storage plan for:
  - Portainer API token
  - MCP access token
  - `.env` file (git-ignored)

### Security Policies

- [ ] Understand two-token security model
  - MCP access token (for wrapper authentication)
  - Portainer API token (for Portainer access)

- [ ] Plan for secret management
  - Environment variables (development)
  - Docker secrets (production consideration)

## Project Setup

### Repository Structure

- [ ] Project directory created
  ```
  portainer-mcp-docker-wrapper/
  ├── cmd/
  │   └── wrapper/
  ├── internal/
  │   ├── auth/
  │   ├── bridge/
  │   └── config/
  ├── docs/
  ├── Dockerfile
  ├── docker-compose.yml
  ├── .env.example
  ├── .gitignore
  ├── go.mod
  └── README.md
  ```

- [ ] Git repository initialized
  ```bash
  git init
  ```

- [ ] `.gitignore` configured
  ```
  .env
  .env.local
  *.log
  bin/
  dist/
  ```

### Go Module Setup

- [ ] Go module initialized
  ```bash
  go mod init portainer-mcp-wrapper
  ```

- [ ] MCP Go SDK dependency planned
  ```bash
  go get github.com/modelcontextprotocol/go-sdk
  ```

## Testing Environment

### Local Testing Setup

- [ ] Test Portainer instance available
  - Can be same as production or separate
  - Read-only mode recommended for initial testing

- [ ] Network test plan
  - Same machine (localhost)
  - Same network (LAN)
  - Different network (if applicable)

### Validation Criteria

- [ ] Success criteria defined:
  - Container builds successfully
  - Health endpoint responds
  - Authentication works
  - MCP protocol handshake succeeds
  - Can list Portainer environments from Claude Code

## Phase 1 Specific Requirements

### Architecture Design

- [ ] Reviewed architecture options:
  - Go HTTP Wrapper (recommended)
  - Node.js Bridge
  - Direct Fork & Modify

- [ ] Decision documented on approach selection

### Design Decisions to Make

- [ ] Transport bridge approach
- [ ] Authentication strategy
- [ ] Configuration management method
- [ ] Error handling approach
- [ ] Logging strategy

### Deliverables for Phase 1

Phase 1 should produce:
- [ ] Architecture diagram
- [ ] Component specifications
- [ ] Security design document
- [ ] Configuration schema
- [ ] File structure plan
- [ ] Dependencies list

## Estimated Time

**Phase 1 Duration**: 2-4 hours

This includes:
- Reviewing all documentation
- Making architectural decisions
- Designing component interfaces
- Planning file structure
- Documenting decisions

## Next Steps After Prerequisites

Once all prerequisites are met:

1. Review architecture options in detail
2. Make final architectural decisions
3. Design component interfaces
4. Plan error handling and logging
5. Document all decisions
6. Move to Phase 2: Development

## Support and Resources

### Getting Help

- MCP Discord: https://discord.gg/modelcontextprotocol
- Portainer Forums: https://community.portainer.io/
- Project Documentation: See `/docs` directory

### Reference Implementation

- Portainer MCP Source: Review for understanding CLI args and configuration
- MCP Go SDK Examples: Check repository for transport examples

## Validation Checklist

Before proceeding to Phase 2, verify:

- [ ] All software installed and working
- [ ] Portainer instance accessible
- [ ] Network connectivity confirmed
- [ ] Tokens generated and stored securely
- [ ] Documentation reviewed
- [ ] Project structure created
- [ ] Go module initialized
- [ ] Architecture decisions documented
- [ ] Ready to write code

---

## Quick Start Commands

```bash
# 1. Verify Go installation
go version

# 2. Verify Docker
docker --version
docker ps

# 3. Create project directory
mkdir portainer-mcp-docker-wrapper
cd portainer-mcp-docker-wrapper

# 4. Initialize Git
git init

# 5. Initialize Go module
go mod init portainer-mcp-wrapper

# 6. Generate MCP access token
openssl rand -base64 32

# 7. Test Portainer API
curl -H "X-API-Key: YOUR_PORTAINER_TOKEN" http://PORTAINER_URL/api/endpoints

# 8. Create basic structure
mkdir -p cmd/wrapper internal/{auth,bridge,config} docs

# 9. Ready to start Phase 1!
```

## Common Issues During Setup

**Issue**: Go not in PATH
- Solution: Add Go bin directory to system PATH

**Issue**: Docker daemon not running
- Solution: Start Docker Desktop or Docker service

**Issue**: Cannot access Portainer
- Solution: Check firewall, network connectivity, and API token

**Issue**: Port 8080 already in use
- Solution: Use different port via `MCP_PORT` environment variable

---

**Status**: Complete this checklist before beginning Phase 1 development.
