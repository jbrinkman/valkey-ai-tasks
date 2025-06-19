# Valkey MCP Task Management Server

A task management system that implements the Model Context Protocol (MCP) for seamless integration with agentic AI tools. This system allows AI agents to create, manage, and track tasks within plans using Valkey as the persistence layer.

## Features

- Plan management (create, read, update, delete)
- Task management (create, read, update, delete)
- Task ordering and prioritization
- Status tracking for tasks
- MCP server for AI agent integration
- Docker container support for easy deployment

## Architecture

The system is built using:

- **Go**: For the backend implementation
- **Valkey**: For data persistence
- **Valkey-Glide v2**: Official Go client for Valkey
- **Model Context Protocol**: For AI agent integration

## Directory Structure

```
valkey-ai-tasks/
├── go/                    # Go implementation
│   ├── cmd/               # Command-line applications
│   │   └── mcpserver/     # MCP server entry point
│   ├── internal/          # Internal packages
│   │   ├── models/        # Data models
│   │   ├── mcp/           # MCP server implementation
│   │   └── storage/       # Valkey storage layer
│   └── Dockerfile         # Docker build file for Go implementation
├── docker-compose.yml     # Docker Compose configuration
└── README.md              # This file
```

## Getting Started

### Prerequisites

- Go 1.24 or later
- Valkey server
- Docker and Docker Compose (for containerized deployment)
- GNU Make (for running Makefile commands)

### Installation

#### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/jbrinkman/valkey-ai-tasks.git
   cd valkey-ai-tasks
   ```

2. Install dependencies:
   ```bash
   cd go
   go mod download
   ```

3. Run the MCP server:
   ```bash
   go run cmd/mcpserver/main.go
   ```

### Running Tests

The project includes a Makefile in the `go` directory with targets for running tests:

1. Run all tests:
   ```bash
   cd go
   make test
   ```

2. Run integration tests only:
   ```bash
   cd go
   make integ-test
   ```

3. Run tests with a filter:
   ```bash
   cd go
   make test filter=TestName
   ```

4. Run tests with verbose output:
   ```bash
   cd go
   make test verbose=1
   ```

5. Generate test coverage report:
   ```bash
   cd go
   make coverage
   ```

6. View all available Makefile targets:
   ```bash
   cd go
   make help
   ```

#### Docker Deployment

1. Clone the repository:
   ```bash
   git clone https://github.com/jbrinkman/valkey-ai-tasks.git
   cd valkey-ai-tasks
   ```

2. Start the containers:
   ```bash
   docker-compose up -d
   ```

3. The MCP server will be available at `http://localhost:8080`

## Environment Variables

The MCP server can be configured using the following environment variables:

- `VALKEY_HOST`: Valkey server hostname (default: "localhost")
- `VALKEY_PORT`: Valkey server port (default: 6379)
- `VALKEY_USERNAME`: Valkey username (default: "")
- `VALKEY_PASSWORD`: Valkey password (default: "")
- `SERVER_PORT`: MCP server port (default: 8080)

## MCP API Reference

The MCP server exposes the following endpoints:

- `GET /mcp/list_functions`: Lists all available functions
- `POST /mcp/invoke/{function_name}`: Invokes a function with the given parameters

### Available Functions

#### Project Management

- `create_project`: Create a new project
- `get_project`: Get a project by ID
- `list_projects`: List all projects
- `list_projects_by_application`: List all projects for a specific application
- `update_project`: Update an existing project
- `delete_project`: Delete a project by ID

#### Task Management

- `create_task`: Create a new task in a project
- `get_task`: Get a task by ID
- `list_tasks_by_project`: List all tasks in a project
- `list_tasks_by_status`: List all tasks with a specific status
- `update_task`: Update an existing task
- `delete_task`: Delete a task by ID
- `reorder_task`: Change the order of a task within its project

## MCP Configuration

### Local MCP Configuration

To configure an AI agent to use the local MCP server, add the following to your `~/.codeium/windsurf/mcp_config.json` file:

```json
{
  "mcpServers": {
    "valkey-tasks": {
      "serverUrl": "http://localhost:8080/sse"
    }
  }
}
```

### Docker MCP Configuration

When running in Docker, use the container name as the hostname:

```json
{
  "mcpServers": {
    "valkey-tasks": {
      "serverUrl": "http://valkey-mcp-server:8080/sse
    }
  }
}
```

If accessing from outside the Docker network:

```json
{
  "mcpServers": {
    "valkey-tasks": {
      "serverUrl": "http://localhost:8080/sse"
    }
  }
}
```

## Using with AI Agents

AI agents can interact with this task management system through the MCP API. Here's an example of how an agent might use the API:

1. The agent calls `/sse/list_functions` to discover available functions
2. The agent calls `/sse/invoke/create_project` with parameters:
   ```json
   {
     "application_id": "my-app",
     "name": "New Feature Development",
     "description": "Implement new features for the application"
   }
   ```
3. The agent can add tasks to the project using either:
   - Individual task creation with `/sse/invoke/create_task`
   - Bulk task creation with `/sse/invoke/bulk_create_tasks` for multiple tasks at once:
     ```json
     {
       "project_id": "project-123",
       "tasks_json": "[
         {
           \"title\": \"Task 1\",
           \"description\": \"Description for task 1\",
           \"priority\": \"high\",
           \"status\": \"pending\"
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
Create a project for this application and add the following tasks:
1. Set up database schema
2. Implement REST API endpoints
3. Create user authentication system
4. Design frontend dashboard
5. Implement inventory tracking features

Prioritize the tasks appropriately and set the first two tasks as "in_progress".
```

With this prompt, an AI agent with access to the Valkey MCP Task Management Server would:
1. Create a new project with application_id "inventory-manager"
2. Add the five specified tasks to the project
3. Set appropriate priorities for each task
4. Update the status of the first two tasks to "in_progress"
5. Return a summary of the created project and tasks

## License

This project is licensed under the BSD-3-Clause License.
