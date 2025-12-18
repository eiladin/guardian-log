# Multi-stage, multi-architecture Dockerfile for Guardian-Log
# Supports: linux/amd64, linux/arm64

# Stage 1: Build Frontend
FROM --platform=$BUILDPLATFORM node:20-alpine AS frontend-builder

WORKDIR /app/web

# Copy package files
COPY web/package*.json ./

# Install dependencies (including devDependencies for build)
RUN npm ci

# Copy source files
COPY web/ ./

# Build frontend
RUN npm run build

# Stage 2: Build Backend
FROM --platform=$BUILDPLATFORM golang:alpine AS backend-builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Copy built frontend from previous stage
COPY --from=frontend-builder /app/web/dist ./web/dist

# Prepare for embedding
RUN mkdir -p webfs/web && cp -r web/dist webfs/web/

# Build binary with optimizations
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w -X main.version=${VERSION:-dev}" \
    -o guardian-log \
    ./cmd/guardian-log

# Stage 3: Runtime
FROM alpine:latest

# Add ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 guardian && \
    adduser -D -u 1000 -G guardian guardian

WORKDIR /app

# Copy binary from builder
COPY --from=backend-builder /app/guardian-log .

# Create data directory
RUN mkdir -p /app/data && chown -R guardian:guardian /app

# Switch to non-root user
USER guardian

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Run the application
CMD ["./guardian-log"]
