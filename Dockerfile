# Build stage
FROM golang:1.23-alpine AS builder

# Set necessary environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod go.sum ./
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -ldflags="-s -w" -o shorturl-api ./cmd/api

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /build/shorturl-api .

# Build a small image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create appuser to run the application
RUN addgroup -g 1001 appgroup && adduser -u 1001 -G appgroup -s /bin/sh -D appuser

WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /dist/shorturl-api .

# Change ownership to appuser
RUN chown appuser:appgroup shorturl-api

# Switch to appuser
USER appuser

# Expose port 8080
EXPOSE 8080

# Command to run when starting the container
CMD ["./shorturl-api"]