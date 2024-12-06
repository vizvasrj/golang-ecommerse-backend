# Use the official Golang image as the base image
FROM golang:1.22.4 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o server cmd/server/main.go

# Use a minimal base image to reduce the size of the final image
FROM gcr.io/distroless/base-debian10

# Copy the binary from the builder stage
COPY --from=builder /app/server /server

# Expose port 8080 to the outside world
EXPOSE 3000

# Command to run the executable
CMD ["/server"]