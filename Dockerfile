FROM golang:1.24 AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o mcpserver ./cmd/mcpserver

# Create a production image
FROM ubuntu:22.04

# Add CA certificates for HTTPS and other required packages
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/mcpserver .

# Expose the server port
EXPOSE 8080

# Set environment variables with defaults
ENV VALKEY_HOST=valkey
ENV VALKEY_PORT=6379
ENV VALKEY_USERNAME=""
ENV VALKEY_PASSWORD=""
ENV SERVER_PORT=8080

# Default transport configuration
ENV ENABLE_SSE=true
ENV ENABLE_STREAMABLE_HTTP=false
ENV ENABLE_STDIO=false

# Run the server with appropriate configuration
# Note: When using STDIO transport, container must be run with -i (interactive) flag
# Example: docker run -i --rm mcpserver
CMD ["./mcpserver"]
