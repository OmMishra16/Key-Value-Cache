# Use a more optimized base image
FROM golang:1.21-alpine AS builder

# Enable Go modules optimization
ENV GO111MODULE=on
ENV GOOS=linux
ENV CGO_ENABLED=0

WORKDIR /app

# Copy only go.mod first
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .
RUN go build -ldflags="-s -w" -o kvcache .

# Use alpine for final image (instead of scratch to have shell access)
FROM alpine:latest

# Copy binary from builder stage
COPY --from=builder /app/kvcache /

EXPOSE 7171

ENTRYPOINT ["/kvcache"]