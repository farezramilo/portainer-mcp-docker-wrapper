# Portainer MCP Docker Wrapper - Project Brief

## Project Overview

Create a dockerized version of the Portainer MCP server that enables remote control of a Portainer instance from Claude Code running on a different Windows machine. The solution will expose the MCP server over HTTP/SSE transport instead of the default stdio transport.

## Goals

- **Primary**: Enable remote MCP connectivity from Windows Claude Code to Portainer instance
- **Secondary**: Maintain security best practices for remote MCP server access
- **Tertiary**: Create reusable, configurable Docker deployment

## Current State

- **Portainer MCP**: Go-based server using stdio transport (subprocess model)
- **Configuration**: Requires Portainer address, API token, optional tools YAML
- **Platform Support**: Linux (amd64, arm64), macOS (arm64)
- **Version**: Supports Portainer 2.31.2 (v0.6.0)

## Technical Challenges

1. **Transport Mismatch**: Portainer MCP uses stdio; remote access requires HTTP/SSE
2. **Network Security**: Remote MCP server needs authentication and encryption
3. **Cross-Machine Architecture**: Windows client → Docker host → Portainer instance
4. **Configuration Management**: Securely manage API tokens and endpoints

## Success Criteria

- Docker container successfully exposes MCP server over HTTP/SSE
- Windows Claude Code can connect and execute Portainer commands remotely
- Secure authentication mechanism implemented
- Easy deployment via Docker Compose
- Comprehensive documentation for setup and usage

## Architecture Decision

**Recommended Approach**: Create a thin HTTP/SSE wrapper around the existing Portainer MCP binary using the MCP Go SDK's streamable HTTP transport capabilities.

**Rationale**: This approach minimizes code changes, leverages official SDK transports, and maintains compatibility with upstream Portainer MCP updates.

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

This wrapper bridges the transport gap while maintaining security and simplicity.
