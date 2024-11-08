# Start with an official Go image as the base image
FROM golang:1.22-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy the Go modules files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the application code
COPY . .

COPY .env .env


# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api-gateway ./cmd/main.go

# Multi-stage for minimal final image
FROM alpine:latest

# Set up a non-root user
RUN adduser -D -g '' appuser

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/api-gateway .

# Copy the HTML templates to the final image
COPY --from=builder /app/templates ./templates

# Grant permissions for the non-root user
RUN chown -R appuser /app && chmod +x /app/api-gateway

# Set the user to run the app
USER appuser

# Expose the port the app will run on
EXPOSE 8080

# Command to run the executable
CMD ["./api-gateway"]
