package bridge

import (
	"context"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"portainer-mcp-wrapper/internal/config"
)

// CreateMCPServer creates an MCP server that proxies to Portainer MCP subprocess
// NOTE: This is a simplified version. The actual bridge implementation will be
// refined once we can test with the Portainer MCP binary in Docker.
func CreateMCPServer(ctx context.Context, cfg *config.Config) (*mcp.Server, error) {
	// Build command arguments for Portainer MCP
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

	// Create command for Portainer MCP
	cmd := exec.CommandContext(ctx, cfg.MCPBinaryPath, args...)

	// For now, store the command (we'll implement the actual bridge in Docker)
	_ = cmd

	// Create MCP server with implementation details
	impl := &mcp.Implementation{
		Name:    "portainer-mcp-wrapper",
		Version: "1.0.0",
	}

	// Create server with minimal options
	server := mcp.NewServer(impl, &mcp.ServerOptions{})

	// TODO: Implement actual transport bridge
	// This will be completed when testing in Docker with Portainer MCP binary

	return server, nil
}

// GetPortainerCommand builds the Portainer MCP command for manual execution
// This is a helper function for debugging
func GetPortainerCommand(cfg *config.Config) *exec.Cmd {
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

	return exec.Command(cfg.MCPBinaryPath, args...)
}
