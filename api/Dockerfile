# Use the official Golang image as the base image
FROM golang:1.20 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download Go modules
RUN go mod download

# Copy the application source code
COPY . .

# Build the Go application
RUN go build -o main .

# Use a minimal base image for the final container
FROM alpine:latest

# Install SQLite for database support
RUN apk --no-cache add sqlite

# Set the working directory inside the container
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Copy the SQLite database and schema files
COPY ./game.db ./schema.sql ./

# Copy the .env file for sensitive keys
COPY .env .

# Load environment variables from the .env file
ENV $(cat .env | xargs)

# Initialize the database schema
RUN sqlite3 game.db < schema.sql

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./main"]