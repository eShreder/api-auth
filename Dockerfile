# Build stage
FROM golang:1.24 AS builder

# Set working directory
WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o user-server ./cmd/server

# Final stage
FROM alpine:latest

# Install required packages
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

# Create required directories
WORKDIR /app
RUN mkdir -p /app/keys

# Copy binary from build stage
COPY --from=builder /app/user-server .
COPY --from=builder /app/.env.example .env

# Set permissions
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./user-server"] 