# Dockerfile
# Multi-stage build with proper multi-arch support

# Build stage - automatically uses the correct Go image for the target platform
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

# Build arguments for cross-compilation
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application for the target platform
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s -extldflags '-static' \
    -X main.version=${VERSION} \
    -X main.commit=${COMMIT} \
    -X main.date=${BUILD_DATE} \
    -X main.builtBy=docker" \
    -a -installsuffix cgo \
    -o k8s-diff-informer \
    ./cmd/k8s-diff-informer

# Final stage - use Alpine with CA certificates
FROM alpine:3.20

# Install CA certificates and create user
RUN apk --no-cache add ca-certificates && \
    adduser -D -u 65532 -g 65532 k8s-diff-informer

# Copy the binary from builder stage
COPY --from=builder /app/k8s-diff-informer /usr/local/bin/k8s-diff-informer

# Ensure binary is executable
RUN chmod +x /usr/local/bin/k8s-diff-informer

# Use non-root user
USER k8s-diff-informer

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/k8s-diff-informer"]
