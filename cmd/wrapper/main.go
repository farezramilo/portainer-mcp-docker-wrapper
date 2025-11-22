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

	log.Printf("Starting Portainer MCP Wrapper")
	log.Printf("Portainer URL: %s", cfg.PortainerURL)
	log.Printf("MCP Port: %d", cfg.MCPPort)
	if cfg.ReadOnlyMode {
		log.Printf("WARNING: Running in READ-ONLY mode")
	}

	// Create context for lifecycle management
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create MCP server factory function
	// This is called for each new HTTP/SSE session
	getServer := func(req *http.Request) *mcp.Server {
		server, err := bridge.CreateMCPServer(ctx, cfg)
		if err != nil {
			log.Printf("Failed to create MCP server: %v", err)
			return nil
		}
		return server
	}

	// Create HTTP handler with SSE support using MCP SDK
	handler := mcp.NewStreamableHTTPHandler(getServer, &mcp.StreamableHTTPOptions{
		SessionTimeout: 5 * time.Minute,
		Stateless:      false, // Enable session tracking
	})

	// Wrap with authentication middleware
	authHandler := auth.NewAuthMiddleware(cfg.MCPAccessToken)(handler)

	// Create HTTP mux with routes
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

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}

		cancel() // Cancel main context
	}()

	// Start server
	log.Printf("Portainer MCP Wrapper listening on :%d", cfg.MCPPort)
	log.Printf("Health check available at http://localhost:%d/health", cfg.MCPPort)
	log.Printf("MCP endpoint available at http://localhost:%d/", cfg.MCPPort)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("Server stopped gracefully")
}

// healthCheckHandler provides a simple health check endpoint
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"portainer-mcp-wrapper","version":"1.0.0"}`)
}
