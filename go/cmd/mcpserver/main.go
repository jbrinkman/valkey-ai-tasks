package main

import (
	"context"
	"log"
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
	serverPortStr := getEnv("SERVER_PORT", "8080")
	serverPort, err := strconv.Atoi(serverPortStr)
	if err != nil {
		log.Fatalf("Invalid SERVER_PORT: %v", err)
	}

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
	planRepo := storage.NewPlanRepository(valkeyClient)
	taskRepo := storage.NewTaskRepository(valkeyClient)

	// Create MCP server using the mark3labs/mcp-go library
	// Convert concrete types to interfaces
	var planRepoInterface storage.PlanRepositoryInterface = planRepo
	var taskRepoInterface storage.TaskRepositoryInterface = taskRepo
	mcpServer := mcp.NewMCPGoServer(planRepoInterface, taskRepoInterface)

	// Set up signal handling for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start the MCP server in a goroutine
	go func() {
		log.Printf("Initializing MCP server on port %d", serverPort)
		if err := mcpServer.Start(serverPort); err != nil {
			log.Fatalf("MCP server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server...")

	// Give the server some time to finish ongoing requests
	time.Sleep(2 * time.Second)

	log.Println("Server exited properly")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
