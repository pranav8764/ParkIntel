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

ARG ONNXRUNTIME_VERSION=1.27.0

# Install runtime dependencies and curl to fetch the ONNX Runtime C library.
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    libgomp1 \
    libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

# Download and unpack official ONNX Runtime C library release
RUN set -eux; \
    case "$(dpkg --print-architecture)" in \
        amd64) ort_arch="x64" ;; \
        arm64) ort_arch="aarch64" ;; \
        *) echo "Unsupported ONNX Runtime architecture: $(dpkg --print-architecture)" >&2; exit 1 ;; \
    esac; \
    ort_dir="onnxruntime-linux-${ort_arch}-${ONNXRUNTIME_VERSION}"; \
    curl -fsSL -o onnxruntime.tgz "https://github.com/microsoft/onnxruntime/releases/download/v${ONNXRUNTIME_VERSION}/${ort_dir}.tgz"; \
    tar -xzf onnxruntime.tgz; \
    mv "${ort_dir}/lib/libonnxruntime.so.${ONNXRUNTIME_VERSION}" "/usr/lib/"; \
    ln -sf "/usr/lib/libonnxruntime.so.${ONNXRUNTIME_VERSION}" /usr/lib/libonnxruntime.so; \
    rm -rf onnxruntime.tgz "${ort_dir}"

ENV ONNX_RUNTIME_LIB_PATH=/usr/lib/libonnxruntime.so.${ONNXRUNTIME_VERSION}
ENV ONNX_MODEL_DIR=/app/models/onnx
ENV LD_LIBRARY_PATH=/usr/lib

# Copy the compiled binary from the build stage
COPY --from=builder /app/parkintel-backend .

# Copy the database schema file (it is needed by the backend)
COPY --from=builder /app/schema.sql .

# Copy ONNX model artifacts into the production image. Railway does not have
# Docker Compose bind mounts, so the models must be baked into the image.
COPY models/onnx ./models/onnx

# Expose backend port
EXPOSE 8080

# Run the backend binary
CMD ["./parkintel-backend"]
