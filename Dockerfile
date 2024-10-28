# Stage 1: Build the Go binary
FROM golang:1.19 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the kubeidle binary for Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o kubeidle cmd/kubeidle/main.go

# Stage 2: Create a minimal final image
FROM alpine:latest

# Install CA certificates for network requests
RUN apk --no-cache add ca-certificates

# Set the working directory inside the container
WORKDIR /root/

# Copy the kubeidle binary from the builder stage
COPY --from=builder /app/kubeidle .

# Set environment variables (optional defaults, can be overridden)
ENV START_TIME=1800
ENV STOP_TIME=0800

# Expose the necessary ports (if any) and set the entrypoint
ENTRYPOINT ["./kubeidle"]