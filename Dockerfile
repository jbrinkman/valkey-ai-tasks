FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o mcpserver ./cmd/mcpserver

# Create a minimal production image
FROM alpine:latest

# Add CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

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

# Run the server
CMD ["./mcpserver"]
