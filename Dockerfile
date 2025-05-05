# Start from the official Go image as a build stage
FROM golang:1.23 AS builder

# Set working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o ewallet-api ./cmd/server/main.go

# Final stage: use a minimal base image
FROM alpine:latest

# Install necessary certificates
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the compiled binary from builder stage
COPY --from=builder /app/ewallet-api .

# Expose the service port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./ewallet-api"]
