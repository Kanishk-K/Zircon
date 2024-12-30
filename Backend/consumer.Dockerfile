# Step 1: Build the Go application using the official Go image
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files first to cache dependencies
COPY go.mod go.sum ./

# Download dependencies (cached if they haven't changed)
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application binary
RUN go build -o main ./cmd/Consumer

# Step 2: Create a minimal production image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Install ffmpeg
RUN apk add --no-cache ffmpeg

# Copy the Go binary from the builder stage
COPY --from=builder /app/main .
RUN chmod +x ./main

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]