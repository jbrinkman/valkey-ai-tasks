# Valkey MCP Task Management Server

[![Build and Publish Container](https://github.com/jbrinkman/valkey-ai-tasks/actions/workflows/publish-container-image.yml/badge.svg)](https://github.com/jbrinkman/valkey-ai-tasks/actions/workflows/publish-container-image.yml)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg)](https://conventionalcommits.org)

A task management system that implements the Model Context Protocol (MCP) for seamless integration with agentic AI tools. This system allows AI agents to create, manage, and track tasks within plans using Valkey as the persistence layer.

## Features

- Plan management (create, read, update, delete)
- Task management (create, read, update, delete)
- Task ordering and prioritization
- Status tracking for tasks
- Notes support with Markdown formatting for both plans and tasks
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
├── cmd/                  # Command-line applications
│   └── mcpserver/        # MCP server entry point
├── examples/             # Example files and templates
│   └── agent_prompts.md  # Example agent prompts for using notes
├── internal/             # Internal packages
│   ├── models/           # Data models
│   ├── mcp/              # MCP server implementation
│   ├── storage/          # Valkey storage layer
│   └── utils/            # Utility functions
│       └── markdown/     # Markdown processing utilities
├── tests/                # Test files
│   ├── integration/      # Integration tests
│   └── utils/            # Test utilities
├── Dockerfile            # Docker build file
├── docker-compose.yml    # Docker Compose configuration
├── go.mod                # Go module definition
└── README.md             # This file
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
   go mod download
   ```

3. Run the MCP server:
   ```bash
   make run
   # or directly with:
   go run cmd/mcpserver/main.go
   ```

### Running Tests

The project includes a Makefile with targets for running tests:

1. Run all tests:
   ```bash
   make test
   ```

2. Run integration tests only:
   ```bash
   make integ-test
   ```

3. Run tests with a filter:
   ```bash
   make test filter=TestName
   ```

4. Run tests with verbose output:
   ```bash
   make test verbose=1
   ```

5. Generate test coverage report:
   ```bash
   make coverage
   ```

6. View all available Makefile targets:
   ```bash
   make help
   ```

#### Docker Deployment

The MCP server is designed to run one protocol at a time for simplicity. By default, all protocols are disabled and you need to explicitly enable the one you want to use.

### Prerequisites

1. Create a named volume for Valkey data persistence:
   ```bash
   docker volume create valkey-data
   ```

### Running with SSE (Recommended for most use cases)

```bash
docker run -d --name valkey-mcp \
  -p 8080:8080 \
  -p 6379:6379 \
  -v valkey-data:/data \
  -e ENABLE_SSE=true \
  valkey-tasks-mcp-server:latest
```

### Running with Streamable HTTP

```bash
docker run -d --name valkey-mcp \
  -p 8080:8080 \
  -p 6379:6379 \
  -v valkey-data:/data \
  -e ENABLE_STREAMABLE_HTTP=true \
  valkey-tasks-mcp-server:latest
```

### Running with STDIO (For direct process communication)

```bash
docker run -i --rm --name valkey-mcp \
  -v valkey-data:/data \
  -e ENABLE_STDIO=true \
  valkey-tasks-mcp-server:latest
```

### Building the Docker Image

```bash
docker build -t valkey-tasks-mcp-server:latest .
```

## Environment Variables

The MCP server can be configured using the following environment variables:

### Database Configuration
- `VALKEY_HOST`: Valkey server hostname (default: "localhost")
- `VALKEY_PORT`: Valkey server port (default: 6379)
- `VALKEY_USERNAME`: Valkey username (default: "")
- `VALKEY_PASSWORD`: Valkey password (default: "")

### Server Configuration
- `SERVER_PORT`: MCP server port (default: 8080)

### Transport Configuration (Only one should be enabled at a time)
- `ENABLE_SSE`: Enable SSE transport (default: "false")
- `SSE_ENDPOINT`: URL path for SSE transport (default: "/sse")
- `SSE_KEEP_ALIVE`: Enable keep-alive for SSE (default: "true")
- `SSE_KEEP_ALIVE_INTERVAL`: Interval for SSE keep-alive messages in seconds (default: 15)
- `ENABLE_STREAMABLE_HTTP`: Enable Streamable HTTP transport (default: "false")
- `STREAMABLE_HTTP_ENDPOINT`: URL path for Streamable HTTP transport (default: "/mcp")
- `STREAMABLE_HTTP_HEARTBEAT_INTERVAL`: Interval for Streamable HTTP heartbeat messages in seconds (default: 30)
- `STREAMABLE_HTTP_STATELESS`: Enable stateless mode for Streamable HTTP (default: "false")
- `ENABLE_STDIO`: Enable STDIO transport (default: "false")
- `STDIO_ERROR_LOG`: Log errors to stderr when using STDIO (default: "true")

### HTTP Server Configuration
- `SERVER_READ_TIMEOUT`: Maximum duration for reading the entire request in seconds (default: 60)
- `SERVER_WRITE_TIMEOUT`: Maximum duration for writing the response in seconds (default: 60)

## CI/CD Workflow

This project uses GitHub Actions for continuous integration and delivery, automatically building and publishing container images to GitHub Container Registry (ghcr.io).

### Conventional Commits

All commits to this repository must follow the [Conventional Commits](https://www.conventionalcommits.org/) specification. This enables automated versioning, changelog generation, and release management.

The commit message format is:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Where `type` is one of:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation changes
- `style`: Changes that don't affect code meaning (formatting, etc.)
- `refactor`: Code changes that neither fix bugs nor add features
- `perf`: Performance improvements
- `test`: Adding or correcting tests
- `build`: Changes to build system or dependencies
- `ci`: Changes to CI configuration
- `chore`: Other changes that don't modify source or test files

Breaking changes must be indicated by `!` after the type/scope or by including `BREAKING CHANGE:` in the footer.

### Automated Versioning

The project uses [Semantic Release](https://github.com/semantic-release/semantic-release) to automatically determine the next version number based on conventional commits:

- `fix:` commits increment the patch version (1.0.0 → 1.0.1)
- `feat:` commits increment the minor version (1.0.0 → 1.1.0)
- Breaking changes increment the major version (1.0.0 → 2.0.0)

### Container Image Workflow

The GitHub Actions workflow automatically:

1. Validates that commits follow the conventional format
2. Determines the next semantic version based on commit messages
3. Builds the Docker image using the project's Dockerfile
4. Tags the image with the semantic version and 'latest' tag
5. Pushes the image to GitHub Container Registry (ghcr.io)
6. Creates a Git tag for the version
7. Generates a changelog
8. Creates a GitHub Release with release notes

#### Required GitHub Secrets

To enable the CI/CD workflow to function properly, the following GitHub secrets need to be configured in your repository:

- `GITHUB_TOKEN`: This is automatically provided by GitHub Actions and is used for authentication with GitHub Container Registry. Make sure it has the necessary permissions for `packages:read` and `packages:write`.

No additional secrets are required as the workflow uses the built-in `GITHUB_TOKEN` for all authentication needs.

### Using the Container Images

The container images are published to GitHub Container Registry and can be pulled using:

```bash
docker pull ghcr.io/jbrinkman/valkey-ai-tasks:latest
# or a specific version
docker pull ghcr.io/jbrinkman/valkey-ai-tasks:1.2.3
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

#### Project Management

- `create_project`: Create a new project
- `get_project`: Get a project by ID
- `list_projects`: List all projects
- `list_projects_by_application`: List all projects for a specific application
- `update_project`: Update an existing project
- `delete_project`: Delete a project by ID
- `update_project_notes`: Update notes for a project
- `get_project_notes`: Get notes for a project

#### Task Management

- `create_task`: Create a new task in a project
- `get_task`: Get a task by ID
- `list_tasks_by_project`: List all tasks in a project
- `list_tasks_by_status`: List all tasks with a specific status
- `update_task`: Update an existing task
- `delete_task`: Delete a task by ID
- `reorder_task`: Change the order of a task within its project
- `update_task_notes`: Update notes for a task
- `get_task_notes`: Get notes for a task

## MCP Configuration

### Local MCP Configuration

To configure an AI agent to use the local MCP server, add the following to your `~/.codeium/windsurf/mcp_config.json` file:

#### Using SSE Transport (Default)

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

```json
{
  "mcpServers": {
    "valkey-tasks": {
      "serverUrl": "http://localhost:8080/mcp",
      "transport": "streamable-http"
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
        "-e", "ENABLE_SSE=false",
        "-e", "ENABLE_STREAMABLE_HTTP=false",
        "-e", "ENABLE_STDIO=true",
        "valkey-mcp-server"
      ]
    }
  }
}
```

Alternatively, for a local binary:

```json
{
  "mcpServers": {
    "valkey-tasks": {
      "command": "/path/to/mcpserver",
      "args": [],
      "env": {
        "ENABLE_SSE": "false",
        "ENABLE_STREAMABLE_HTTP": "false",
        "ENABLE_STDIO": "true"
      },
      "transport": "stdio"
    }
  }
}
```

When using STDIO transport, the server must be started with the `ENABLE_STDIO=true` environment variable, and the client must be configured to use the `stdio` transport type. No `serverUrl` is required for STDIO transport as communication happens directly through stdin/stdout.

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

#### Using Streamable HTTP Transport

```json
{
  "mcpServers": {
    "valkey-tasks": {
      "serverUrl": "http://valkey-mcp-server:8080/mcp",
      "transport": "streamable-http"
    }
  }
}
```

If accessing from outside the Docker network:

```json
{
  "mcpServers": {
    "valkey-tasks": {
      "serverUrl": "http://localhost:8080/mcp",
      "transport": "streamable-http"
    }
  }
}
```

#### Using STDIO Transport with Docker

For Docker environments, agentic tools can use this configuration to start and manage the MCP server:

```json
{
  "mcpServers": {
    "valkey-tasks": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e", "ENABLE_SSE=false",
        "-e", "ENABLE_STREAMABLE_HTTP=false",
        "-e", "ENABLE_STDIO=true",
        "valkey-mcp-server"
      ],
      "transport": "stdio"
    }
  }
}
```

For manual testing or development, you can run the container directly:

```bash
# Run the MCP server with STDIO transport enabled
docker run -i --rm \
  -e ENABLE_SSE=false \
  -e ENABLE_STREAMABLE_HTTP=false \
  -e ENABLE_STDIO=true \
  valkey-mcp-server
```

Alternatively, you can use the provided `docker-compose.stdio.yml` file:

```bash
docker-compose -f docker-compose.stdio.yml up mcpserver-stdio
```

Note that the `-i` flag (or `stdin_open: true` in docker-compose) is essential for STDIO transport to work properly, as it keeps stdin open for communication.

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

## Using with AI Agents

AI agents can interact with this task management system through the MCP API using either SSE or Streamable HTTP transport. Here are examples for both transport protocols:

### Using SSE Transport

1. The agent calls `/sse/list_functions` to discover available functions
2. The agent calls `/sse/invoke/create_project` with parameters:
   ```json
   {
     "application_id": "my-app",
     "name": "New Feature Development",
     "description": "Implement new features for the application",
     "notes": "# Project Notes\n\nThis project aims to implement the following features:\n\n- Feature A\n- Feature B\n- Feature C"
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
5. The agent can add or update notes for a project using `/sse/invoke/update_project_notes`:
   ```json
   {
     "id": "project-123",
     "notes": "# Updated Project Notes\n\nAdded new requirements:\n\n- Requirement X\n- Requirement Y"
   }
   ```
6. The agent can add or update notes for a task using `/sse/invoke/update_task_notes`:
   ```json
   {
     "id": "task-456",
     "notes": "# Updated Task Notes\n\nFound a better approach:\n\n```go\nfunc betterSolution() {\n  // code here\n}\n```"
   }
   ```

### Using Streamable HTTP Transport

1. The agent sends a POST request to `/mcp` with the following JSON to discover available functions:
   ```json
   {
     "method": "list_functions",
     "params": {}
   }
   ```

2. The agent sends a POST request to `/mcp` to create a project:
   ```json
   {
     "method": "invoke",
     "params": {
       "function": "create_project",
       "params": {
         "application_id": "my-app",
         "name": "New Feature Development",
         "description": "Implement new features for the application",
         "notes": "# Project Notes\n\nThis project aims to implement the following features:\n\n- Feature A\n- Feature B\n- Feature C"
       }
     }
   }
   ```

3. The agent can add tasks to the project by sending a POST request to `/mcp`:
   ```json
   {
     "method": "invoke",
     "params": {
       "function": "bulk_create_tasks",
       "params": {
         "project_id": "project-123",
         "tasks_json": "[{\"title\": \"Task 1\", \"description\": \"Description for task 1\", \"priority\": \"high\", \"status\": \"pending\"}, {\"title\": \"Task 2\", \"description\": \"Description for task 2\", \"priority\": \"medium\", \"status\": \"pending\"}]"
       }
     }
   }
   ```

### Sample Agent Prompt

Here's a sample prompt that would trigger an AI agent to use the MCP task management system:

```
I need to organize work for my new application called "inventory-manager". 
Create a project for this application with the following project notes:
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
1. Create a new project with application_id "inventory-manager" and the specified Markdown-formatted notes
2. Add the five specified tasks to the project
3. Add detailed Markdown-formatted notes to the database schema task
4. Set appropriate priorities for each task
5. Update the status of the first two tasks to "in_progress"
6. Return a summary of the created project and tasks

## License

This project is licensed under the BSD-3-Clause License.
