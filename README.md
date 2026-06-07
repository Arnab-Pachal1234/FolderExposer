# 🛡️ FolderExposer

## Zero-Trust Multi-Tenant Network Gateway

**FolderExposer** is a secure, multi-tenant reverse tunneling application built in **Go**. It allows you to expose local folders to the public internet using dynamic subdomains, while keeping all files safely stored on your local device.

Engineered with a **Zero-Trust Network Access** approach, FolderExposer includes a unified Cobra CLI, custom raw TCP socket routing, HTTP protocol hijacking, collision-resistant subdomain generation, and an ultra-lightweight Docker deployment pipeline.

---

## ✨ Key Features

### 🚀 Unified Cobra CLI Engine

- Single powerful binary handles both the **Relay Server** and the **Local Client**.
- Supports command-line arguments for dynamic ports, passwords, and subdomains.
- Built on [`spf13/cobra`](https://github.com/spf13/cobra), the same CLI framework used by tools like Kubernetes and Docker.

### 🌍 Multi-Tenant Subdomain Routing

- True reverse proxy architecture using HTTP `Host` headers.
- Multiple users can connect simultaneously using different subdomains.

  Example:

  ```text
  movie-api.yourdomain.com
  docs.yourdomain.com
  files.yourdomain.com
  ```

- Built-in collision detection safely rejects duplicate subdomain requests.
- Automatically generates secure randomized subdomains when no subdomain is provided.

### 🔐 Zero-Trust Access Control

- HTTP Basic Authentication for public users.
- Client handshake authentication using secure initialization tokens.
- Public users never directly access the local machine.
- Every request is verified before the relay bridges the socket.

### 🐳 Production-Ready Docker Pipeline

- Ultra-lightweight multi-stage Alpine Linux container.
- Final Docker image size is approximately **15MB**.
- Simple VPS/cloud deployment using Docker.
- Automated CI/CD pipeline using **GoReleaser** and **GitHub Actions**.
- Cross-compiles Windows, macOS, and Linux binaries on every release.

### 🧠 Aggressive Memory Management & Concurrency

- Mutex-based busy locks prevent race conditions during large file or video transfers.
- Bidirectional teardown tripwires immediately free sockets when a browser tab closes or a transfer finishes.
- Drops unnecessary secondary requests when the tunnel is actively streaming data.

---

## 🏗️ Architecture Overview

```text
[ Public Browser ]
        |
        | HTTP Request
        | Host: my-api.localhost
        | Basic Auth
        v
=================================================
[ VPS Relay Server - Port 8080 ]
-------------------------------------------------
1. Parses HTTP Host Header for Subdomain
2. Verifies Basic Authentication
3. Checks Active Subdomain Socket in Memory Map
4. Sends NEW_REQUEST Signal to Correct Local Client
5. Bridges Browser and Local File Server
=================================================
        ^                              |
        | Control Channel - Port 9000  | Data Channel - Port 9001
        |                              v
=================================================
[ Local Go Client - FolderExposer CLI ]
-------------------------------------------------
1. Requests Subdomain and Maintains Heartbeat
2. Receives Relay Instructions
3. Opens Temporary Data Socket
4. Connects to Internal Local File Server
=================================================
        |
        v
[ Local File Server - Dynamic Port ]
```

---

## 🚀 Getting Started

You can run FolderExposer in three ways:

1. Run natively using Go.
2. Deploy the relay server using Docker.
3. Download a pre-built binary from GitHub Releases.

---

## 🧑‍💻 Native Setup

### 1. Clone the Repository

```bash
git clone https://github.com/Arnab-Pachal1234/FolderExposer.git
cd FolderExposer
```

---

### 2. Start the Relay Server

Run the relay server on your public machine, VPS, or cloud server.

```bash
go run ./cmd/folder-exposer server
```

By default, the relay server uses:

| Service | Default Port | Purpose |
|---|---:|---|
| Public HTTP Relay | `8080` | Public browser access |
| Control Channel | `9000` | Persistent client control connection |
| Data Channel | `9001` | Temporary high-speed data transfer |

You can also customize the ports:

```bash
go run ./cmd/folder-exposer server --public-port 80 --control-port 7000
```

---

### 3. Expose a Local Folder

On the machine containing the files you want to share, run:

```bash
go run ./cmd/folder-exposer expose ./shared_folder --subdomain my-api --password "secure123"
```

Leave the subdomain blank to let the server generate a random public URL:

```bash
go run ./cmd/folder-exposer expose ./shared_folder --password "secure123"
```

---

### ✅ Expected Output

```text
=======================================================
🚀 SUCCESS! YOUR FOLDER IS LIVE ON THE INTERNET
=======================================================
🌍 Public URL: http://my-api.localhost:8080
🔒 Password:   secure123
📡 Local Port: 8081
=======================================================
Listening for incoming connections. Press CTRL+C to stop.
```

---

## 🐳 Docker Deployment

Docker is the recommended way to deploy the **FolderExposer Relay Server** on a VPS or cloud server.

### 1. Pull and Run the Docker Image

Run this command on your cloud server:

```bash
docker run -d \
  --name folder-exposer \
  --restart unless-stopped \
  -p 80:8080 \
  -p 9000:9000 \
  -p 9001:9001 \
  arnabpachal/folder-exposer:latest
```

---

### 2. Docker Port Mapping

| Host Port | Container Port | Purpose |
|---:|---:|---|
| `80` | `8080` | Public browser access without typing a port |
| `9000` | `9000` | Persistent control channel |
| `9001` | `9001` | Ephemeral data channel for file transfers |

---

### 3. Check Running Container

```bash
docker ps
```

You should see a running container named:

```text
folder-exposer
```

---

### 4. View Logs

```bash
docker logs -f folder-exposer
```

---

### 5. Stop the Relay Server

```bash
docker stop folder-exposer
```

---

### 6. Remove the Container

```bash
docker rm folder-exposer
```

---

## 📦 Releases & Binaries

FolderExposer uses **GoReleaser** and **GitHub Actions** to automatically compile ready-to-use executables.

Instead of installing Go, users can download the latest `folder-exposer` binary for their operating system from the GitHub **Releases** page.

### Example Usage with Downloaded Binary

#### Linux / macOS

```bash
./folder-exposer expose ./my_folder --subdomain test --password "1234"
```

#### Windows

```bash
folder-exposer.exe expose ./my_folder --subdomain test --password "1234"
```

---

## 📁 Project Structure

```text
FolderExposer/
│
├── cmd/
│   └── folder-exposer/
│       ├── main.go      # Entry point
│       ├── root.go      # Base CLI command
│       ├── server.go    # Reverse proxy server logic
│       └── expose.go    # Local client agent logic
│
├── pkg/
│   └── tunnel/          # Shared network and message struct models
│
├── Dockerfile           # Multi-stage production Docker build
├── .goreleaser.yaml     # CI/CD cross-compilation config
└── go.mod               # Go module file
```

---

## 🔐 Security Notes

FolderExposer is designed for **educational**, **experimental**, and **security architecture learning** purposes.

Before using it in a production environment with highly sensitive data, it is strongly recommended to implement:

- TLS/HTTPS encryption.
- Strong authentication tokens.
- Rate limiting.
- Request logging and monitoring.
- Domain-level access controls.
- Firewall rules for relay server ports.

> Public users should never directly access your local machine. FolderExposer follows a relay-based approach where every request is verified before socket bridging.

---

## 🧪 Example Use Cases

- Share a local project demo with friends or testers.
- Expose a folder from your laptop without uploading files to a cloud provider.
- Test reverse tunneling and relay server architecture.
- Learn raw TCP socket routing and HTTP hijacking in Go.
- Build a lightweight alternative to local tunneling tools for learning purposes.

---

## 🛠️ Tech Stack

| Technology | Purpose |
|---|---|
| Go | Core application logic |
| Cobra CLI | Unified command-line interface |
| Raw TCP Sockets | Relay and data channel communication |
| HTTP Host Routing | Subdomain-based multi-tenant routing |
| Docker | Lightweight relay server deployment |
| Alpine Linux | Minimal production image |
| GoReleaser | Cross-platform release automation |
| GitHub Actions | CI/CD pipeline |

---

## 👨‍💻 Author

**Arnab Pachal**

GitHub: [Arnab-Pachal1234](https://github.com/Arnab-Pachal1234)

---

## ⭐ Support

If you find this project useful, consider giving it a star on GitHub.

```text
Secure. Lightweight. Multi-Tenant. Zero-Trust.
```
