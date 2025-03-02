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

# Install AWS CLI
RUN apk add --no-cache aws-cli

# Install Fonts
RUN mkdir /usr/share/fonts
COPY --from=builder /app/fonts /usr/share/fonts
RUN chmod -R 644 /usr/share/fonts
RUN fc-cache -fv

# Copy the Go binary from the builder stage
COPY --from=builder /app/main .
RUN chmod +x ./main

# Copy the static files
# COPY --from=builder /app/static ./static
COPY --from=builder /app/entrypoint.sh .
RUN chmod +x ./entrypoint.sh

# Command to run the executable
CMD ["./entrypoint.sh"]