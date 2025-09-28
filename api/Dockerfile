# Use the official Golang image as the base image
FROM golang:1.24.0 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy the application source code
COPY . .

# Build the Go application. A standard dynamic build is fine.
RUN go build -o main .

# Use a debian-slim image which is compatible with the builder's libraries
FROM debian:12-slim

# Install SQLite using apt-get
RUN apt-get update && apt-get install -y sqlite3 && rm -rf /var/lib/apt/lists/*

# Set the working directory inside the container
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Copy the SQLite database and schema files
COPY ./game.db ./schema.sql ./

# Copy the .env file for sensitive keys
COPY .env .

# Initialize the database schema
RUN sqlite3 game.db < schema.sql

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./main"]