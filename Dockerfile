# Use a more optimized base image with specific version
FROM golang:1.21.0-alpine3.18 AS builder

# Enable Go modules optimization
ENV GO111MODULE=on
ENV GOOS=linux
ENV CGO_ENABLED=0

WORKDIR /app

# Copy only go.mod first
COPY go.mod ./

# Download dependencies with retry mechanism
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy source code
COPY . .

# Build with optimizations
RUN go build -ldflags="-s -w" -o kvcache .

# Use specific alpine version
FROM alpine:3.18.0

# Add CA certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder stage
COPY --from=builder /app/kvcache /

EXPOSE 7171

ENTRYPOINT ["/kvcache"]