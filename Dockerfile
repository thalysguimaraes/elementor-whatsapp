# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X 'main.version=$(git describe --tags --always --dirty)' -X 'main.commit=$(git rev-parse --short HEAD)' -X 'main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
    -o ewctl \
    ./cmd/ewctl

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 ewctl && \
    adduser -D -u 1000 -G ewctl ewctl

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/ewctl /usr/local/bin/ewctl

# Copy example configuration
COPY --from=builder /build/examples/config.example.yaml /app/config.example.yaml

# Create config directory
RUN mkdir -p /home/ewctl/.config/ewctl && \
    chown -R ewctl:ewctl /home/ewctl

# Switch to non-root user
USER ewctl

# Set home directory
ENV HOME=/home/ewctl

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ewctl version || exit 1

# Default command
ENTRYPOINT ["ewctl"]