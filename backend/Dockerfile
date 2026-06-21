# Start from Golang official bookworm image to build the Go binary
FROM golang:1.26-bookworm AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire backend directory
COPY . .

# Build the Go application
# Note: we must keep CGO enabled since onnxruntime_go bindings require CGO
RUN CGO_ENABLED=1 GOOS=linux go build -o parkintel-backend main.go

# Runtime stage using debian:bookworm-slim
FROM debian:bookworm-slim

WORKDIR /app

# Install dependencies (curl to download libonnxruntime, ca-certificates for secure requests)
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Download and unpack official ONNX Runtime C library release
RUN curl -sSL -o onnxruntime.tgz https://github.com/microsoft/onnxruntime/releases/download/v1.26.0/onnxruntime-linux-x64-1.26.0.tgz \
    && tar -xzf onnxruntime.tgz \
    && mv onnxruntime-linux-x64-1.26.0/lib/libonnxruntime.so.1.26.0 /usr/lib/ \
    && ln -s /usr/lib/libonnxruntime.so.1.26.0 /usr/lib/libonnxruntime.so \
    && rm -rf onnxruntime.tgz onnxruntime-linux-x64-1.26.0

# Copy the compiled binary from the build stage
COPY --from=builder /app/parkintel-backend .

# Copy the database schema file (it is needed by the backend)
COPY --from=builder /app/schema.sql .

# Expose backend port
EXPOSE 8080

# Run the backend binary
CMD ["./parkintel-backend"]
