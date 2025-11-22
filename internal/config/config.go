package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the MCP wrapper
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

// LoadConfig loads configuration from environment variables
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

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvAsIntOrDefault returns environment variable as int or default
func getEnvAsIntOrDefault(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// getEnvAsBoolOrDefault returns environment variable as bool or default
func getEnvAsBoolOrDefault(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultVal
}
