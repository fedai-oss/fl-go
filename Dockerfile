# FL-Go Dockerfile
# Multi-stage build for efficient containerization

# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git protobuf-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fx cmd/fx/main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates python3 py3-pip py3-numpy

# Create app user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/fx .

# Copy Python training script
COPY --from=builder /app/scripts/create_initial_model.py ./scripts/

# Create necessary directories
RUN mkdir -p data save logs src

# Change ownership to app user
RUN chown -R appuser:appgroup /app

# Switch to app user
USER appuser

# Expose default ports
EXPOSE 50051 50052 50053

# Set default command
ENTRYPOINT ["./fx"]

# Default command shows help
CMD ["--help"]
