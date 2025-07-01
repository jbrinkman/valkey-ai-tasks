# Valkey MCP Task Management Server

[![Lint](https://github.com/jbrinkman/valkey-ai-tasks/actions/workflows/lint.yml/badge.svg)](https://github.com/jbrinkman/valkey-ai-tasks/actions/workflows/lint.yml)
[![Publish Container](https://github.com/jbrinkman/valkey-ai-tasks/actions/workflows/publish-container-image.yml/badge.svg)](https://github.com/jbrinkman/valkey-ai-tasks/actions/workflows/publish-container-image.yml)
[![Release](https://github.com/jbrinkman/valkey-ai-tasks/actions/workflows/create-release.yml/badge.svg)](https://github.com/jbrinkman/valkey-ai-tasks/actions/workflows/create-release.yml)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg)](https://conventionalcommits.org)

A task management system that implements the Model Context Protocol (MCP) for seamless integration with agentic AI tools. This system allows AI agents to create, manage, and track tasks within plans using Valkey as the persistence layer.

## Features

- Plan management (create, read, update, delete)
- Task management (create, read, update, delete)
- Task ordering and prioritization
- Status tracking for tasks
- Notes support with Markdown formatting for both plans and tasks
- MCP server for AI agent integration
- Supports STDIO, SSE and Streamable HTTP transport protocols
- Docker container support for easy deployment

## Architecture

The system is built using:

- **Go**: For the backend implementation
- **Valkey**: For data persistence
- **Valkey-Glide v2**: Official Go client for Valkey
- **Model Context Protocol**: For AI agent integration

## Quick Start

### Docker Deployment

The MCP server is designed to run one protocol at a time for simplicity. By default, all protocols are disabled and you need to explicitly enable the one you want to use.

#### Prerequisites

1. Create a named volume for Valkey data persistence:
   ```bash
   docker volume create valkey-data
   ```

#### Running with SSE (Recommended for most use cases)

```bash
docker run -d --name valkey-mcp \
  -p 8080:8080 \
  -p 6379:6379 \
  -v valkey-data:/data \
  -e ENABLE_SSE=true \
  ghcr.io/jbrinkman/valkey-ai-tasks:latest
```

#### Running with Streamable HTTP

```bash
docker run -d --name valkey-mcp \
  -p 8080:8080 \
  -p 6379:6379 \
  -v valkey-data:/data \
  -e ENABLE_STREAMABLE_HTTP=true \
  ghcr.io/jbrinkman/valkey-ai-tasks:latest
```

#### Running with STDIO (For direct process communication)

```bash
docker run -i --rm --name valkey-mcp \
  -v valkey-data:/data \
  -e ENABLE_STDIO=true \
  ghcr.io/jbrinkman/valkey-ai-tasks:latest
```

### Using the Container Images

The container images are published to GitHub Container Registry and can be pulled using:

```bash
docker pull ghcr.io/jbrinkman/valkey-ai-tasks:latest
# or a specific version
docker pull ghcr.io/jbrinkman/valkey-ai-tasks:1.1.0
```

## MCP API Reference

The MCP server supports two transport protocols: Server-Sent Events (SSE) and Streamable HTTP. Each protocol exposes similar endpoints but with different interaction patterns.

### Server-Sent Events (SSE) Endpoints

- `GET /sse/list_functions`: Lists all available functions
- `POST /sse/invoke/{function_name}`: Invokes a function with the given parameters

### Streamable HTTP Endpoints

- `POST /mcp`: Handles all MCP requests using JSON format
  - For function listing: `{"method": "list_functions", "params": {}}`
  - For function invocation: `{"method": "invoke", "params": {"function": "function_name", "params": {...}}}`

### Transport Selection

The server automatically selects the appropriate transport based on:

1. **URL Path**: Connect to the specific endpoint for your preferred transport
2. **Content Type**: When connecting to the root path (`/`), the server redirects based on content type:
   - `application/json` → Streamable HTTP
   - Other content types → SSE

### Health Check

- `GET /health`: Returns server health status

### Available Functions

#### Plan Management

- `create_plan`: Create a new plan
- `get_plan`: Get a plan by ID
- `list_plans`: List all plans
- `list_plans_by_application`: List all plans for a specific application
- `update_plan`: Update an existing plan
- `delete_plan`: Delete a plan by ID
- `update_plan_notes`: Update notes for a plan
- `get_plan_notes`: Get notes for a plan

#### Task Management

- `create_task`: Create a new task in a plan
- `get_task`: Get a task by ID
- `list_tasks_by_plan`: List all tasks in a plan
- `list_tasks_by_status`: List all tasks with a specific status
- `update_task`: Update an existing task
- `delete_task`: Delete a task by ID
- `reorder_task`: Change the order of a task within its plan
- `update_task_notes`: Update notes for a task
- `get_task_notes`: Get notes for a task

## MCP Configuration

### Local MCP Configuration

To configure an AI agent to use the local MCP server, add the following to your MCP configuration file (the exact file location depends on your AI Agent):

#### Using SSE Transport (Default)

> Note: The docker container should already be running.

```json
{
  "mcpServers": {
    "valkey-tasks": {
      "serverUrl": "http://localhost:8080/sse"
    }
  }
}
```

#### Using Streamable HTTP Transport

> Note: The docker container should already be running.

```json
{
  "mcpServers": {
    "valkey-tasks": {
      "serverUrl": "http://localhost:8080/mcp"
    }
  }
}
```

#### Using STDIO Transport

STDIO transport allows the MCP server to communicate via standard input/output, which is useful for legacy AI tools that rely on stdin/stdout for communication.

For agentic tools that need to start and manage the MCP server process, use a configuration like this:

```json
{
  "mcpServers": {
    "valkey-tasks": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-v", "valkey-data:/data"
        "-e", "ENABLE_STDIO=true",
        "ghcr.io/jbrinkman/valkey-ai-tasks:latest"
      ]    
    }
  }
}
```

### Docker MCP Configuration

When running in Docker, use the container name as the hostname:

#### Using SSE Transport (Default)

```json
{
  "mcpServers": {
    "valkey-tasks": {
      "serverUrl": "http://valkey-mcp-server:8080/sse"
    }
  }
}
```

## Notes Functionality

The system supports rich Markdown-formatted notes for both plans and tasks. This feature is particularly useful for AI agents to maintain context between sessions and document important information.

### Notes Features

- Full Markdown support including:
  - Headings, lists, and tables
  - Code blocks with syntax highlighting
  - Links and images
  - Emphasis and formatting
- Separate notes for plans and tasks
- Dedicated MCP tools for managing notes
- Notes are included in all relevant API responses

### Best Practices for Notes

1. **Maintain Context**: Use notes to document important context that should persist between sessions
2. **Document Decisions**: Record key decisions and their rationale
3. **Track Progress**: Use notes to track progress and next steps
4. **Organize Information**: Use Markdown formatting to structure information clearly
5. **Code Examples**: Include code snippets with proper syntax highlighting

### Notes Security

Notes content is sanitized to prevent XSS and other security issues while preserving Markdown formatting.

## MCP Resources

In addition to MCP tools, the system provides MCP resources that allow AI agents to access structured data directly. These resources provide a complete view of plans and tasks in a single request, which is more efficient than making multiple tool calls.

### Available Resources

#### Plan Resource

The Plan Resource provides a complete view of a plan, including its tasks and notes. It supports the following URI patterns:

- **Single Plan**: `ai-tasks://plans/{id}/full` - Returns a specific plan with its tasks
- **All Plans**: `ai-tasks://plans/full` - Returns all plans with their tasks
- **Application Plans**: `ai-tasks://applications/{app_id}/plans/full` - Returns all plans for a specific application

Each resource returns a JSON object or array with the following structure:

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

### Using MCP Resources

AI agents can access these resources using the MCP resource API. Here's an example of how to read a resource:

```json
{
  "action": "read_resource",
  "params": {
    "uri": "ai-tasks://plans/123/full"
  }
}
```

This will return the complete plan resource including all tasks, which is more efficient than making separate calls to get the plan and then its tasks.

## Using with AI Agents

AI agents can interact with this task management system through the MCP API using either SSE or Streamable HTTP transport. Here are examples for both transport protocols:

### Using SSE Transport

1. The agent calls `/sse/list_functions` to discover available functions
2. The agent calls `/sse/invoke/create_plan` with parameters:
   ```json
   {
     "application_id": "my-app",
     "name": "New Feature Development",
     "description": "Implement new features for the application",
     "notes": "# Project Notes\n\nThis project aims to implement the following features:\n\n- Feature A\n- Feature B\n- Feature C"
   }
   ```
3. The agent can add tasks to the plan using either:
   - Individual task creation with `/sse/invoke/create_task`
   - Bulk task creation with `/sse/invoke/bulk_create_tasks` for multiple tasks at once:
     ```json
     {
       "plan_id": "plan-123",
       "tasks_json": "[
         {
           \"title\": \"Task 1\",
           \"description\": \"Description for task 1\",
           \"priority\": \"high\",
           \"status\": \"pending\",
           \"notes\": \"# Task Notes\\n\\nThis task requires the following steps:\\n\\n1. Step one\\n2. Step two\\n3. Step three\"
         },
         {
           \"title\": \"Task 2\",
           \"description\": \"Description for task 2\",
           \"priority\": \"medium\",
           \"status\": \"pending\"
         }
       ]"
     }
     ```
4. The agent calls `/sse/invoke/update_task` to update task status as work progresses

### Sample Agent Prompt

Here's a sample prompt that would trigger an AI agent to use the MCP task management system:

```
I need to organize work for my new application called "inventory-manager". 
Create a plan for this application with the following plan notes:
"# Inventory Manager Project

This project aims to create a comprehensive inventory management system with the following goals:
- Track inventory levels in real-time
- Generate reports on inventory movement
- Provide alerts for low stock items"

Add the following tasks:
1. Set up database schema
2. Implement REST API endpoints
3. Create user authentication system
4. Design frontend dashboard
5. Implement inventory tracking features

For the database schema task, add these notes:
"# Database Schema Notes

The schema should include the following tables:
- Products
- Categories
- Inventory Transactions
- Users
- Roles"

Prioritize the tasks appropriately and set the first two tasks as "in_progress".
```

With this prompt, an AI agent with access to the Valkey MCP Task Management Server would:
1. Create a new plan with application_id "inventory-manager" and the specified Markdown-formatted notes
2. Add the five specified tasks to the plan
3. Add detailed Markdown-formatted notes to the database schema task
4. Set appropriate priorities for each task
5. Update the status of the first two tasks to "in_progress"
6. Return a summary of the created plan and tasks

## Developer Documentation

For information on how to set up a development environment, contribute to the project, and understand the codebase structure, please refer to the [Developer Guide](DEVELOPERS.md).

For contribution guidelines, including commit message format and pull request process, see [Contributing Guidelines](CONTRIBUTING.md).

## License

This project is licensed under the BSD-3-Clause License.
