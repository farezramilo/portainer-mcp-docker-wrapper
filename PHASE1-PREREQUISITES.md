# Phase 1 Prerequisites - Quick Reference

This document provides a quick checklist of prerequisites before starting Phase 1 (Architecture & Design). For detailed information, see [docs/00-prerequisites-phase1.md](docs/00-prerequisites-phase1.md).

## Essential Software

- [ ] **Go 1.23+** - `go version`
- [ ] **Docker** - `docker --version` && `docker ps`
- [ ] **Docker Compose** - `docker-compose --version`
- [ ] **Git** - `git --version`

## Portainer Access

- [ ] Running Portainer instance (v2.31.2+)
- [ ] Portainer API token generated (starts with `ptr_`)
- [ ] Can access Portainer API:
  ```bash
  curl -H "X-API-Key: YOUR_TOKEN" http://PORTAINER_URL/api/endpoints
  ```

## Network Setup

- [ ] Docker host identified and accessible
- [ ] Port 8080 available on Docker host
- [ ] Windows machine can reach Docker host
- [ ] Firewall rules documented

## Security Preparation

- [ ] Can generate secure tokens: `openssl rand -base64 32`
- [ ] Plan for storing secrets securely
- [ ] Understand two-token model:
  - MCP access token (wrapper auth)
  - Portainer API token (Portainer access)

## Documentation Review

- [ ] Read [Project Brief](docs/01-project-brief.md)
- [ ] Reviewed [Implementation Plan](docs/02-implementation-plan.md)
- [ ] Understand [Technical Architecture](docs/03-technical-architecture.md)

## Project Setup

- [ ] Project directory created
- [ ] Git initialized: `git init`
- [ ] Go module initialized: `go mod init portainer-mcp-wrapper`
- [ ] `.gitignore` configured (exclude `.env`, `*.log`, `bin/`, `dist/`)

## Knowledge Requirements

- [ ] Basic Docker & containerization
- [ ] HTTP/REST APIs & authentication
- [ ] Basic Go programming
- [ ] MCP protocol concepts

## Quick Start Commands

```bash
# 1. Create project
mkdir portainer-mcp-docker-wrapper
cd portainer-mcp-docker-wrapper

# 2. Initialize
git init
go mod init portainer-mcp-wrapper

# 3. Create structure
mkdir -p cmd/wrapper internal/{auth,bridge,config} docs

# 4. Generate token
openssl rand -base64 32

# 5. Test Portainer
curl -H "X-API-Key: YOUR_TOKEN" http://PORTAINER_URL/api/endpoints
```

## Phase 1 Deliverables

Once prerequisites are complete, Phase 1 will produce:

- [ ] Architecture diagram
- [ ] Component specifications
- [ ] Security design document
- [ ] Configuration schema
- [ ] File structure plan
- [ ] Dependencies list

**Estimated Time**: 2-4 hours

## Next Steps

After completing this checklist:

1. ✅ Proceed to Phase 1: Architecture & Design
2. Review architecture options in [Implementation Plan](docs/02-implementation-plan.md#phase-1-architecture--design)
3. Make architectural decisions
4. Design component interfaces
5. Document decisions
6. Move to Phase 2: Development

## Support Resources

- **Full Prerequisites Guide**: [docs/00-prerequisites-phase1.md](docs/00-prerequisites-phase1.md)
- **Documentation Index**: [docs/README.md](docs/README.md)
- **MCP Go SDK**: https://github.com/modelcontextprotocol/go-sdk
- **Portainer MCP**: https://github.com/portainer/portainer-mcp

---

**Status**: Complete this checklist before beginning Phase 1 development.
