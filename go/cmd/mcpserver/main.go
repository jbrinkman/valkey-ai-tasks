package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jbrinkman/valkey-ai-tasks/go/internal/mcp"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/storage"
)

func main() {
	// Get environment variables or use defaults
	valkeyHost := getEnv("VALKEY_HOST", "localhost")
	valkeyPortStr := getEnv("VALKEY_PORT", "6379")
	valkeyPort, err := strconv.Atoi(valkeyPortStr)
	if err != nil {
		log.Fatalf("Invalid VALKEY_PORT: %v", err)
	}
	valkeyUsername := getEnv("VALKEY_USERNAME", "")
	valkeyPassword := getEnv("VALKEY_PASSWORD", "")
	serverPort := getEnv("SERVER_PORT", "8080")

	// Initialize Valkey client
	valkeyClient, err := storage.NewValkeyClient(valkeyHost, valkeyPort, valkeyUsername, valkeyPassword)
	if err != nil {
		log.Fatalf("Failed to initialize Valkey client: %v", err)
	}
	defer valkeyClient.Close()

	// Ping Valkey to ensure connection
	ctx := context.Background()
	if err := valkeyClient.Ping(ctx); err != nil {
		log.Fatalf("Failed to connect to Valkey: %v", err)
	}
	log.Printf("Connected to Valkey at %s:%d", valkeyHost, valkeyPort)

	// Initialize repositories
	projectRepo := storage.NewProjectRepository(valkeyClient)
	taskRepo := storage.NewTaskRepository(valkeyClient)

	// Create MCP server handler
	mcpHandler := mcp.NewMCPServer(projectRepo, taskRepo)

	// Set up HTTP server
	server := &http.Server{
		Addr:    ":" + serverPort,
		Handler: mcpHandler,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("MCP server starting on port %s", serverPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
