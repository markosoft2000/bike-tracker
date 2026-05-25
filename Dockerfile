# Stage 1: Build the binaries
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install build dependencies (CGO and librdkafka are required for confluent-kafka-go)
RUN apk add --no-cache git build-base librdkafka-dev pkgconf

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Enable CGO for the Kafka wrapper
ENV CGO_ENABLED=1

# Copy the entire project
COPY . .

# BUILD WITH CACHE MOUNTS
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -tags dynamic -o gateway-server ./cmd/gateway/main.go

# Stage 2: Final lightweight image
FROM alpine:3.21
WORKDIR /app

# Install runtime dependencies for Kafka
RUN apk add --no-cache librdkafka ca-certificates

# Copy binaries from the builder stage
COPY --from=builder /app/gateway-server .

# Copy configuration and migration files
COPY --from=builder /app/configs ./configs

# Expose gRPC and HTTP ports
EXPOSE 8080

# Command to run
CMD ["./gateway-server"]