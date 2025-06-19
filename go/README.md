# Go Implementation - Valkey MCP Task Management Server

This directory contains the Go implementation of the Valkey MCP Task Management Server.

## Directory Structure

```
go/
├── cmd/                 # Application entry points
│   └── server/          # MCP server
├── internal/            # Internal packages
│   ├── api/             # API handlers and routes
│   ├── config/          # Configuration
│   ├── models/          # Data models
│   └── storage/         # Valkey storage implementation
├── pkg/                 # Public packages
│   └── mcp/             # MCP protocol implementation
└── tests/               # Tests
```

## Requirements

- Go 1.20 or later
- Valkey server running locally or remotely
- Valkey-Glide client for Go

## Setup

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Configure the application:
   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your Valkey connection details
   ```

3. Build the application:
   ```bash
   go build -o bin/mcp-server ./cmd/server
   ```

4. Run the server:
   ```bash
   ./bin/mcp-server
   ```

## API Endpoints

The MCP server exposes the following endpoints:

- `list_resources` - List available resources (plans and tasks)
- `read_resource` - Read a specific resource (plan or task)
- `create_resource` - Create a new resource
- `update_resource` - Update an existing resource
- `delete_resource` - Delete a resource

See the API documentation in the `/docs` directory for more details.
