# ==========================================
# STAGE 1: The Builder
# ==========================================
FROM golang:1.20-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod file
COPY go.mod ./

# Copy the rest of your source code
COPY . .

# Compile the server binary statically
# CGO_ENABLED=0 ensures the binary doesn't rely on external C libraries
# -ldflags="-w -s" strips debugging data to make the file size extremely small
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server-bin ./cmd/server/main.go

# ==========================================
# STAGE 2: The Production Image
# ==========================================
# Alpine is a highly secure, ultra-lightweight Linux distribution (~5MB)
FROM alpine:latest

# Install CA certificates just in case you add HTTPS/SSL later
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy ONLY the compiled binary from the builder stage above
COPY --from=builder /app/server-bin .

# Expose the specific ports your server needs to operate
# 8080: Public Proxy | 9000: Control Channel | 9001: Data Channel
EXPOSE 8080 9000 9001

# Command to run when the container starts
CMD ["./server-bin"]