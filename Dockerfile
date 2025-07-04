FROM golang:bookworm AS builder

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
FROM valkey/valkey:8

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/mcpserver .

# Create a custom entrypoint script that leverages the bundle-docker-entrypoint.sh
# but also starts our MCP server
RUN cat <<EOF > /app/custom-entrypoint.sh
#!/bin/bash
# First run the original bundle-docker-entrypoint.sh to set up Valkey
# but with the --daemonize flag to run it in the background
VALKEY_ARGS="$@"

valkey-server --daemonize yes --save 60 1 --loglevel warning --appendonly yes --appendfsync everysec --dir /data --dbfilename valkey.db --appendfilename valkey.aof

# Wait for Valkey to be ready
until valkey-cli ping; do
echo "Waiting for Valkey to start..."
sleep 1
done
echo "Valkey is ready!"

# Run the MCP server with appropriate configuration
exec ./mcpserver
EOF
RUN chmod +x /app/custom-entrypoint.sh
# Expose both Valkey and MCP server ports
EXPOSE 6379 8080

# Set environment variables with defaults
ENV VALKEY_HOST=localhost
ENV VALKEY_PORT=6379
ENV VALKEY_USERNAME=""
ENV VALKEY_PASSWORD=""
ENV SERVER_PORT=8080

# Default transport configuration
ENV ENABLE_SSE=false
ENV ENABLE_STREAMABLE_HTTP=false
ENV ENABLE_STDIO=false
ENV STDIO_ERROR_LOG=true

# Use our custom entrypoint script
ENTRYPOINT ["/app/custom-entrypoint.sh"]
CMD ["valkey-server"]
