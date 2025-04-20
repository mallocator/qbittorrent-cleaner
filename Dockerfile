# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install UPX for binary compression
RUN apk add --no-cache upx

# Copy Go module files and source code
COPY go.mod .
COPY main.go .
COPY qbittorrent/ ./qbittorrent/

# Build the Go application with optimizations for size
# -s -w: strip debugging information
# -trimpath: removes file system paths from the resulting binary
# -ldflags="-s -w": removes symbol table and DWARF debugging information
# Additional flags for maximum size reduction
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -extldflags '-static'" -trimpath -o qbt-clean . && \
    # Compress the binary with UPX using best compression (--best) and LZMA algorithm (--lzma)
    upx --best --lzma qbt-clean

# Final stage - using scratch (the smallest possible base image)
FROM scratch

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from the builder stage
COPY --from=builder /app/qbt-clean /qbt-clean

# Set environment variables with defaults
ENV DOWNLOAD_DIRS=/downloads
ENV SERVER_URL=https://10.0.0.1:8080
ENV SERVER_USER=admin
ENV SERVER_PASS=adminadmin

# Run the application
ENTRYPOINT ["/qbt-clean"]
