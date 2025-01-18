# Stage 1: Build the Go binary
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy the Go modules files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build arguments for target platform
ARG TARGETARCH
ARG TARGETOS

# Build the kubeidle binary for the target platform
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o kubeidle cmd/kubeidle/main.go

# Stage 2: Create a minimal final image
FROM --platform=$TARGETPLATFORM alpine:latest

# Install CA certificates for network requests
RUN apk --no-cache add ca-certificates

# Set the working directory inside the container
WORKDIR /root/

# Copy the kubeidle binary from the builder stage
COPY --from=builder /app/kubeidle .

# Set environment variables (optional defaults, can be overridden)
ENV START_TIME=1800
ENV STOP_TIME=0800

# Expose metrics port
EXPOSE 9095

# Set the entrypoint
ENTRYPOINT ["./kubeidle"]
