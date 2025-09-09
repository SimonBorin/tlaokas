FROM golang:1.24.4-alpine

# Install git (required by go modules sometimes)
RUN apk update && apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
ENV GOPROXY=direct
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary from cmd/server
RUN go build -o server ./cmd/server

# Expose port for HTTP server
EXPOSE 8080

# Run the server
CMD ["./server"]