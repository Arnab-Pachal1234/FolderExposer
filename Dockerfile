# ==========================================
# STAGE 1: The Builder
# ==========================================
FROM golang:alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod ./
# COPY go.sum ./  <-- Uncomment this if you have a go.sum file!

# Download all dependencies (like Cobra and AutoCert)
RUN go mod download

# Copy the rest of your source code
COPY . .

# Compile the unified Cobra CLI binary statically
# Pointing to the new folder-exposer directory!
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server-bin ./cmd/folder-exposer/

# ==========================================
# STAGE 2: The Production Image
# ==========================================
FROM alpine:latest

# Install CA certificates for Let's Encrypt Auto-TLS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy ONLY the compiled binary from the builder stage
COPY --from=builder /app/server-bin .

# Expose standard HTTP/HTTPS and your custom tunnel ports
EXPOSE 80 443 9000 9001

# Start the binary, but DON'T run the server command yet. 
# We will pass "server --domain XYZ" dynamically from the VPS!
ENTRYPOINT ["./server-bin"]