# =========================
# BUILD STAGE
# =========================

FROM golang:1.25.1-alpine AS builder

WORKDIR /app

# Install git (required for some Go module resolutions)
RUN apk add --no-cache git

# Copy dependency files first for layer caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy application source
COPY . .

# Build optimized static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-w -s" \
  -o server ./cmd/api

# =========================
# RUNTIME STAGE
# =========================

FROM alpine:3.20

# Install runtime certificates
RUN apk --no-cache add ca-certificates

# Create non-root user/group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Application directory
WORKDIR /home/appuser

# Copy binary from builder
COPY --from=builder /app/server .

# Ensure correct ownership
RUN chown -R appuser:appgroup /home/appuser

# Switch to least-privileged user
USER appuser

# Application port
EXPOSE 8080

# Start application
CMD ["./server"]