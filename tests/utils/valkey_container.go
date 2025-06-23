// Package utils provides testing utilities for the valkey-ai-tasks project
package utils

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/valkey"
	"github.com/testcontainers/testcontainers-go/wait"
	glide "github.com/valkey-io/valkey-glide/go/v2"
	"github.com/valkey-io/valkey-glide/go/v2/config"
)

const (
	// ValkeyImage is the Docker image for Valkey
	ValkeyImage = "valkey/valkey:latest"
	// ValkeyStartupTimeout is the timeout for Valkey container startup
	ValkeyStartupTimeout = 30 * time.Second
)

// ValkeyContainer represents a Valkey container for testing
type ValkeyContainer struct {
	Container *valkey.ValkeyContainer
	URI       string
	Client    *glide.Client
}

// StartValkeyContainer starts a Valkey container for testing
func StartValkeyContainer(ctx context.Context, t *testing.T) (*ValkeyContainer, error) {
	t.Helper()

	req := require.New(t)

	// Create Valkey container request
	valkeyContainer, err := valkey.RunContainer(ctx,
		testcontainers.WithImage(ValkeyImage),
		valkey.WithLogLevel("notice"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(ValkeyStartupTimeout),
		),
	)
	req.NoError(err, "Failed to start Valkey container")

	// Get container endpoint
	endpoint, err := valkeyContainer.Endpoint(ctx, "")
	req.NoError(err, "Failed to get Valkey container endpoint")

	// Parse the endpoint to get host and port
	parts := strings.Split(endpoint, ":")
	req.Equal(2, len(parts), "Expected endpoint format host:port")
	host := parts[0]
	port, err := strconv.Atoi(parts[1])
	req.NoError(err, "Failed to parse Valkey container port")

	// Create Valkey client configuration
	address := &config.NodeAddress{Host: host, Port: port}
	clientConfig := config.NewClientConfiguration().WithAddress(address)

	// Create Valkey client
	client, err := glide.NewClient(clientConfig)
	req.NoError(err, "Failed to create Valkey client")

	// Test connection
	pong, err := client.Ping(ctx)
	req.NoError(err, "Failed to ping Valkey container")
	req.Equal("PONG", pong, "Expected PONG response from Valkey")

	return &ValkeyContainer{
		Container: valkeyContainer,
		URI:       fmt.Sprintf("redis://%s", endpoint),
		Client:    client,
	}, nil
}

// StopValkeyContainer stops a Valkey container
func StopValkeyContainer(ctx context.Context, t *testing.T, container *ValkeyContainer) {
	t.Helper()

	if container == nil || container.Container == nil {
		return
	}

	req := require.New(t)

	// Close client connection
	if container.Client != nil {
		// Close the client connection
		container.Client.Close()
	}

	// Terminate container
	err := container.Container.Terminate(ctx)
	req.NoError(err, "Failed to terminate Valkey container")
}

// SetupValkeyTest sets up a Valkey container for testing
// It returns a context, Valkey container, and cleanup function
func SetupValkeyTest(t *testing.T) (context.Context, *ValkeyContainer, func()) {
	t.Helper()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

	// Start Valkey container
	container, err := StartValkeyContainer(ctx, t)
	require.NoError(t, err, "Failed to start Valkey container")

	// Return cleanup function
	cleanup := func() {
		StopValkeyContainer(ctx, t, container)
		cancel()
	}

	return ctx, container, cleanup
}

// ParseEndpoint parses a host:port endpoint string and returns the host and port
func ParseEndpoint(endpoint string) (string, int, error) {
	parts := strings.Split(endpoint, ":")
	if len(parts) != 2 {
		return "", 0, errors.New("invalid endpoint format, expected host:port")
	}

	host := parts[0]
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse port: %w", err)
	}

	return host, port, nil
}
