# MCP Resources Documentation

This document provides detailed information about the MCP resources available in the Valkey AI Tasks system. These resources allow AI agents to access structured data directly through the Model Context Protocol (MCP).

## Plan Resource

The Plan Resource provides a complete view of a plan, including its tasks and notes. This resource is particularly useful for AI agents that need to understand the full context of a plan without making multiple API calls.

### URI Patterns

The Plan Resource supports the following URI patterns:

| URI Pattern | Description |
|-------------|-------------|
| `ai-tasks://plans/{id}/full` | Returns a specific plan with its tasks |
| `ai-tasks://plans/full` | Returns all plans with their tasks |
| `ai-tasks://applications/{app_id}/plans/full` | Returns all plans for a specific application |

### Resource Structure

#### Single Plan Response

When requesting a single plan (`ai-tasks://plans/{id}/full`), the response will be a JSON object with the following structure:

```json
{
  "id": "plan-123",
  "application_id": "my-app",
  "name": "New Feature Development",
  "description": "Implement new features for the application",
  "status": "new",
  "notes": "# Project Notes\n\nThis project aims to implement the following features...",
  "created_at": "2025-06-27T14:00:21Z",
  "updated_at": "2025-07-01T13:04:01Z",
  "tasks": [
    {
      "id": "task-456",
      "plan_id": "plan-123",
      "title": "Task 1",
      "description": "Description for task 1",
      "status": "pending",
      "priority": "high",
      "order": 0,
      "notes": "# Task Notes\n\nThis task requires the following steps...",
      "created_at": "2025-06-27T14:00:50Z",
      "updated_at": "2025-07-01T12:04:27Z"
    },
    // Additional tasks...
  ]
}
```

#### Multiple Plans Response

When requesting all plans (`ai-tasks://plans/full`) or plans for a specific application (`ai-tasks://applications/{app_id}/plans/full`), the response will be a JSON array of plan objects:

```json
[
  {
    "id": "plan-123",
    "application_id": "my-app",
    "name": "New Feature Development",
    "description": "Implement new features for the application",
    "status": "new",
    "notes": "# Project Notes\n\nThis project aims to implement the following features...",
    "created_at": "2025-06-27T14:00:21Z",
    "updated_at": "2025-07-01T13:04:01Z",
    "tasks": [
      // Tasks for this plan...
    ]
  },
  {
    "id": "plan-456",
    "application_id": "my-app",
    "name": "Bug Fixes",
    "description": "Fix critical bugs in the application",
    "status": "in_progress",
    "notes": "# Bug Fix Notes\n\nThis plan addresses the following critical bugs...",
    "created_at": "2025-06-28T10:15:30Z",
    "updated_at": "2025-07-01T09:22:15Z",
    "tasks": [
      // Tasks for this plan...
    ]
  }
  // Additional plans...
]
```

### Error Handling

The Plan Resource implements robust error handling with specific error types and detailed error messages:

| Error Type | Description | Example |
|------------|-------------|---------|
| `ErrInvalidURI` | The URI format is not supported | "invalid resource URI: 'ai-tasks://invalid/uri' does not match any supported pattern" |
| `ErrPlanNotFound` | The requested plan does not exist | "plan not found: plan with ID 'non-existent-id' does not exist" |
| `ErrInvalidPlanID` | The plan ID is invalid or empty | "invalid plan ID: empty plan ID" |
| `ErrInvalidAppID` | The application ID is invalid or empty | "invalid application ID: empty application ID" |
| `ErrMarshalFailure` | Failed to marshal the resource to JSON | "failed to marshal resource: failed to marshal plan resource for plan '123'" |
| `ErrInternalStorage` | An internal storage error occurred | "internal storage error: failed to get plan with ID '123'" |

### Empty Results Handling

When a request would normally return an empty array (e.g., no plans exist for an application), the resource returns an empty JSON array (`[]`) instead of an error. This is consistent with REST API best practices.

## Using MCP Resources

AI agents can access these resources using the MCP resource API. Here's an example of how to read a resource:

```json
{
  "action": "read_resource",
  "params": {
    "uri": "ai-tasks://plans/123/full"
  }
}
```

### Example: Reading a Single Plan

```json
// Request
{
  "action": "read_resource",
  "params": {
    "uri": "ai-tasks://plans/f50c031d-112e-420b-b5ef-68177cebe43b/full"
  }
}

// Response
{
  "uri": "ai-tasks://plans/f50c031d-112e-420b-b5ef-68177cebe43b/full",
  "mime_type": "application/json",
  "text": "{\"id\":\"f50c031d-112e-420b-b5ef-68177cebe43b\",\"application_id\":\"valkey-ai-tasks\",\"name\":\"Implement MCP Full Plan Resource\",\"description\":\"Implement a new MCP resource that provides a complete view of a plan including its tasks and notes\",\"status\":\"new\",\"notes\":\"## Implementation Approach\\n\\n* *Important Directives**: \\n- Complete one task at a time and seek approval before continuing to the next task.\\n- Update the task status before starting on the next task.\\n\\n## Plan Overview\\nThis plan aims to implement a new MCP resource that provides a complete view of a plan including its tasks and notes. The PlanResource will combine all related data into a single resource for easier consumption by clients.\\n\\n## Tasks Sequence\\n1. Define PlanResource struct\\n2. Implement ResourceProvider interface\\n3. Add resource registration to server\\n4. Implement resource URI handling\\n5. Add error handling\\n6. ~~Add unit tests~~ (Cancelled - minimal business logic to test)\\n7. Update documentation\\n8. Add integration tests\\n\\n## Progress Tracking\\n- Completed tasks:\\n  1. Define PlanResource struct\\n  2. Implement ResourceProvider interface\\n  3. Add resource registration to server\\n  4. Implement resource URI handling\\n  5. Add error handling\\n- Cancelled tasks:\\n  6. Add unit tests (Minimal business logic to test)\\n- Current task: Update documentation\\n- Status: In progress\",\"created_at\":\"2025-06-27T14:00:21Z\",\"updated_at\":\"2025-07-01T13:04:01Z\",\"tasks\":[...]}"
}
```

### Example: Reading All Plans

```json
// Request
{
  "action": "read_resource",
  "params": {
    "uri": "ai-tasks://plans/full"
  }
}

// Response
{
  "uri": "ai-tasks://plans/full",
  "mime_type": "application/json",
  "text": "[{\"id\":\"f50c031d-112e-420b-b5ef-68177cebe43b\",\"application_id\":\"valkey-ai-tasks\",\"name\":\"Implement MCP Full Plan Resource\",...},{\"id\":\"132db63e-9c41-43ec-b6b2-e4a666c67b19\",\"application_id\":\"valkey-ai-tasks\",\"name\":\"Add Backup and Export Functionality\",...}]"
}
```

### Example: Reading Plans for an Application

```json
// Request
{
  "action": "read_resource",
  "params": {
    "uri": "ai-tasks://applications/valkey-ai-tasks/plans/full"
  }
}

// Response
{
  "uri": "ai-tasks://applications/valkey-ai-tasks/plans/full",
  "mime_type": "application/json",
  "text": "[{\"id\":\"f50c031d-112e-420b-b5ef-68177cebe43b\",\"application_id\":\"valkey-ai-tasks\",\"name\":\"Implement MCP Full Plan Resource\",...},{\"id\":\"132db63e-9c41-43ec-b6b2-e4a666c67b19\",\"application_id\":\"valkey-ai-tasks\",\"name\":\"Add Backup and Export Functionality\",...}]"
}
```

## Benefits of Using MCP Resources

1. **Efficiency**: Get all the data you need in a single request
2. **Consistency**: Ensure that plan and task data is consistent
3. **Simplicity**: Simpler code for AI agents that need to work with plans and tasks
4. **Performance**: Reduced number of API calls means better performance

## Best Practices

1. Use resources when you need a complete view of a plan and its tasks
2. Use individual MCP tools when you only need to perform a specific action on a single entity
3. Handle potential errors gracefully, especially "not found" errors
4. Be aware of the response size when requesting all plans, as it could be large
