# Developer Guide for Valkey MCP Task Management Server

This guide provides information for developers who want to contribute to or modify the Valkey MCP Task Management Server codebase.

## Development Environment Setup

### Prerequisites

- Go 1.24 or later
- Valkey server
- Docker and Docker Compose (for containerized development)
- GNU Make (for running Makefile commands)

### Getting Started

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

## Project Structure

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

## Testing

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

## Building the Docker Image

```bash
docker build -t valkey-tasks-mcp-server:latest .
```

## CI/CD Workflow

This project uses GitHub Actions for continuous integration and delivery, automatically building and publishing container images to GitHub Container Registry (ghcr.io).

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

## Development Guidelines

### Code Style

- Follow standard Go coding conventions
- Use `gofmt` to format your code
- Follow the recommendations in [Effective Go](https://golang.org/doc/effective_go)
- Run `golint` and `go vet` to catch common issues

### Testing Guidelines

- Write tests for all new features and bug fixes
- Maintain or improve test coverage with each PR
- Use table-driven tests when appropriate
- Integration tests should be in the `tests/integration` directory
- Unit tests should be in the same package as the code they test

### Documentation

- Update documentation when changing functionality
- Use clear, concise language
- Include examples where appropriate
- Document exported functions, types, and constants

## Contribution Process

Please refer to [CONTRIBUTING.md](CONTRIBUTING.md) for detailed information on how to contribute to this project, including:

- Code of Conduct
- Development Workflow
- Pull Request Process
- Conventional Commits guidelines
- Issue Reporting

## Automated Versioning

The project uses [Semantic Release](https://github.com/semantic-release/semantic-release) to automatically determine the next version number based on conventional commits:

- `fix:` commits increment the patch version (1.0.0 → 1.0.1)
- `feat:` commits increment the minor version (1.0.0 → 1.1.0)
- Breaking changes increment the major version (1.0.0 → 2.0.0)

## License

This project is licensed under the BSD-3-Clause License.
