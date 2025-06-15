package storage

import (
	"context"
	"fmt"
	"github.com/valkey-io/valkey-glide/go/v2"
)

// ValkeyClient wraps the Valkey-Glide client for our application
type ValkeyClient struct {
	client *glide.Client
}

// NewValkeyClient creates a new Valkey client with the given connection options
func NewValkeyClient(address string, port int, username, password string) (*ValkeyClient, error) {
	opts := &glide.ClientOptions{
		Addresses: []string{fmt.Sprintf("%s:%d", address, port)},
		Username:  username,
		Password:  password,
	}

	client, err := glide.NewClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create Valkey client: %w", err)
	}

	return &ValkeyClient{
		client: client,
	}, nil
}

// Ping checks the connection to the Valkey server
func (vc *ValkeyClient) Ping(ctx context.Context) error {
	return vc.client.Ping(ctx).Err()
}

// Close closes the Valkey client connection
func (vc *ValkeyClient) Close() error {
	return vc.client.Close()
}

// Keys used for storing data in Valkey
const (
	// Project keys
	projectKeyPrefix = "project:"
	projectsListKey  = "projects"

	// Task keys
	taskKeyPrefix      = "task:"
	projectTasksPrefix = "project_tasks:"
)

// GetProjectKey returns the key for a specific project
func GetProjectKey(projectID string) string {
	return projectKeyPrefix + projectID
}

// GetTaskKey returns the key for a specific task
func GetTaskKey(taskID string) string {
	return taskKeyPrefix + taskID
}

// GetProjectTasksKey returns the key for a project's tasks list
func GetProjectTasksKey(projectID string) string {
	return projectTasksPrefix + projectID
}
