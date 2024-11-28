# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build/langspace

# Copy go mod
COPY go.mod ./

# Copy go sum
# COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build-main

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

WORKDIR /app/

# Copy the binary from builder
COPY --from=builder /build/langspace/langspace .

# Command to run
ENTRYPOINT ["./langspace"]
