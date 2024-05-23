# Use the official Golang image to create a build artifact
FROM golang:1.21-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the working directory inside the container
COPY . .

# Build the Go app
RUN go build -o kvrest .

##################################################################

# Start a new stage from scratch
FROM alpine:latest  

WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/kvrest .
#COPY --from=builder /app/data ./data

# Expose port 8080 to the outside world
EXPOSE 8080

CMD ["./kvrest"]