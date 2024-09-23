# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go binary
RUN go build -o yt2pc .

# Final image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ffmpeg python3 py3-pip ca-certificates \
    && pip install --no-cache-dir --break-system-packages yt-dlp

# Copy the Go binary from the builder stage
COPY --from=builder /app/yt2pc /yt2pc

# Create the audio cache directory
RUN mkdir -p /audio_cache

# Expose the service port
EXPOSE 8080

# Set the working directory
WORKDIR /

# Start the service
CMD ["/yt2pc"]
