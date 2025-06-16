# Valkey MCP Task Management Server

A task management system that implements the Model Context Protocol (MCP) for seamless integration with agentic AI tools. This system allows AI agents to create, manage, and track tasks within projects using Valkey as the persistence layer.

## Features

- Project management (create, read, update, delete)
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

## Using with AI Agents

AI agents can interact with this task management system through the MCP API. Here's an example of how an agent might use the API:

1. The agent calls `/mcp/list_functions` to discover available functions
2. The agent calls `/mcp/invoke/create_project` to create a new project
3. The agent calls `/mcp/invoke/create_task` to add tasks to the project
4. The agent calls `/mcp/invoke/update_task` to update task status as work progresses

## License

This project is licensed under the BSD-3-Clause License.
