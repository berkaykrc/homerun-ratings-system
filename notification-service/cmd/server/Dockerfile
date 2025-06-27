FROM golang:1.23-alpine AS builder

# Install dependencies including make
RUN apk update && \
    apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.sum go.mod ./
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .
RUN make build

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/notification-service .

# Copy config files
COPY --from=builder /app/config ./config

# Change ownership to appuser
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8081

# Run the application
CMD ["./notification-service"]
