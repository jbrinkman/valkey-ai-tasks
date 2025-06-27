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
FROM valkey/valkey-bundle:8-bookworm

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/mcpserver .

# Create a custom entrypoint script that leverages the bundle-docker-entrypoint.sh
# but also starts our MCP server
RUN echo '#!/bin/bash\n\
    # First run the original bundle-docker-entrypoint.sh to set up Valkey\n\
    # but with the --daemonize flag to run it in the background\n\
    VALKEY_ARGS="$@"\n\
    if [[ "$VALKEY_ARGS" == "valkey-server" ]]; then\n\
    VALKEY_ARGS="valkey-server --daemonize yes --save 60 1 --loglevel warning --appendonly yes --appendfsync everysec --dir /data --dbfilename valkey.db --appendfilename valkey.aof"\n\
    fi\n\
    /usr/local/bin/bundle-docker-entrypoint.sh $VALKEY_ARGS\n\
    \n\
    # Wait for Valkey to be ready\n\
    until valkey-cli ping; do\n\
    echo "Waiting for Valkey to start..."\n\
    sleep 1\n\
    done\n\
    echo "Valkey is ready!"\n\
    \n\
    # Run the MCP server with appropriate configuration\n\
    exec ./mcpserver\n\
    ' > /app/custom-entrypoint.sh && chmod +x /app/custom-entrypoint.sh

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
