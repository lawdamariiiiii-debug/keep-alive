# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Copy source code
COPY *.go ./

# Build the application
RUN go build -o file-keepalive

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/file-keepalive .

# Set environment variables (can be overridden at runtime)
ENV INTERVAL_DAYS=7
ENV MAX_AGE_DAYS=30

# Run the application
CMD ["./file-keepalive"]
