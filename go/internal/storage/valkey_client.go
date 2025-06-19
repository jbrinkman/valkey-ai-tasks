package storage

import (
	"context"
	"fmt"

	glide "github.com/valkey-io/valkey-glide/go/v2"
	"github.com/valkey-io/valkey-glide/go/v2/config"
)

// ValkeyClient wraps the Valkey-Glide client for our application
type ValkeyClient struct {
	client *glide.Client
}

// NewValkeyClient creates a new Valkey client with the given connection options
func NewValkeyClient(address string, port int, username, password string) (*ValkeyClient, error) {
	clientConfig := config.NewClientConfiguration().
		WithAddress(&config.NodeAddress{Host: address, Port: port})

	if username != "" && password != "" {
		clientConfig.WithCredentials(config.NewServerCredentials(username, password))
	}

	client, err := glide.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Valkey client: %w", err)
	}

	return &ValkeyClient{
		client: client,
	}, nil
}

// Ping checks the connection to the Valkey server
func (vc *ValkeyClient) Ping(ctx context.Context) error {
	_, err := vc.client.Ping(ctx)
	return err
}

// Close closes the Valkey client connection
func (vc *ValkeyClient) Close() error {
	vc.client.Close()
	return nil
}

// Keys used for storing data in Valkey
const (
	// Plan keys (formerly Project)
	planKeyPrefix = "plan:"
	plansListKey  = "plans"
	// Legacy project keys (kept for backward compatibility)
	projectKeyPrefix = "project:"
	projectsListKey  = "projects"

	// Task keys
	taskKeyPrefix   = "task:"
	planTasksPrefix = "plan_tasks:"
	// Legacy project tasks keys (kept for backward compatibility)
	projectTasksPrefix = "project_tasks:"
)

// GetPlanKey returns the key for a specific plan
func GetPlanKey(planID string) string {
	return planKeyPrefix + planID
}

// GetProjectKey returns the key for a specific project (legacy)
func GetProjectKey(projectID string) string {
	return projectKeyPrefix + projectID
}

// GetTaskKey returns the key for a specific task
func GetTaskKey(taskID string) string {
	return taskKeyPrefix + taskID
}

// GetPlanTasksKey returns the key for a plan's tasks list
func GetPlanTasksKey(planID string) string {
	return planTasksPrefix + planID
}

// GetProjectTasksKey returns the key for a project's tasks list (legacy)
func GetProjectTasksKey(projectID string) string {
	return projectTasksPrefix + projectID
}
