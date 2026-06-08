# 🛡️ FolderExposer

## Zero-Trust Multi-Tenant Network Gateway

**FolderExposer** is a secure, multi-tenant reverse tunneling application built in **Go**. It allows users to expose local folders to the public internet using dynamic subdomains, while keeping all files safely stored on their local device.

FolderExposer is engineered with a **Zero-Trust Network Access** approach and includes a unified Cobra CLI, custom raw TCP socket routing, HTTP Host-based subdomain routing, collision-resistant subdomain generation, RSA-encrypted tunnel communication, Cloudflare-native HTTPS support, and an ultra-lightweight Docker deployment pipeline.

---

## 📌 What FolderExposer Does

FolderExposer lets you share any folder from your local computer using a public HTTPS URL.

Instead of uploading files to a cloud storage provider, your files stay on your own system. The public user connects to your VPS relay server, and the relay securely forwards the request to your local FolderExposer client.

```text
Public Browser  →  Cloudflare HTTPS  →  VPS Relay Server  →  Your Local Folder
```

---

## 🚀 How Users Can Use FolderExposer

This section is for normal users who only want to expose a local folder.

### 1. Download FolderExposer

Download the latest binary for your operating system from the GitHub **Releases** page.

Example binaries:

```text
folder-exposer-linux-amd64
folder-exposer-darwin-amd64
folder-exposer-windows-amd64.exe
```

For Linux/macOS, rename the downloaded binary to:

```text
folder-exposer
```

For Windows, keep it as:

```text
folder-exposer.exe
```

### 2. Give Execute Permission on Linux/macOS

```bash
chmod +x folder-exposer
```

Windows users can skip this step.

### 3. Open Terminal in the Folder You Want to Share

Example:

```text
my-awesome-project/
```

Open your terminal inside that folder.

### 4. Expose the Current Folder

#### Linux / macOS

```bash
./folder-exposer expose . --subdomain my-awesome-project
```

#### Windows

```bash
folder-exposer.exe expose . --subdomain my-awesome-project
```

The `.` means:

```text
Expose the current folder
```

### 5. Expose a Specific Sub-Folder

#### Linux / macOS

```bash
./folder-exposer expose ./my_shared_folder
```

#### Windows

```bash
folder-exposer.exe expose ./my_shared_folder
```

### 6. Use a Custom Subdomain

```bash
./folder-exposer expose . --subdomain portfolio
```

Expected public URL:

```text
https://portfolio.arnabpachal.site
```

Windows:

```bash
folder-exposer.exe expose . --subdomain portfolio
```

### 7. Auto-Generate a Random Subdomain

If no subdomain is provided, FolderExposer automatically generates one.

```bash
./folder-exposer expose .
```

Example output:

```text
https://blue-river-4821.arnabpachal.site
```

### ✅ Expected Output

```text
=======================================================
🚀 SUCCESS! YOUR FOLDER IS LIVE ON THE INTERNET
=======================================================
🌍 Public URL: https://my-awesome-project.arnabpachal.site
=======================================================
Listening for incoming connections. Press CTRL+C to stop.
```

Now anyone can open the public URL in a browser and access the exposed folder through the secure relay.

---

## 🧑‍💻 Common User Commands

### Expose Current Folder

```bash
./folder-exposer expose .
```

### Expose Current Folder with Custom Subdomain

```bash
./folder-exposer expose . --subdomain my-awesome-project
```

### Expose a Specific Folder

```bash
./folder-exposer expose ./my_shared_folder
```

### Windows Example

```bash
folder-exposer.exe expose . --subdomain my-awesome-project
```

### Stop Sharing

Press:

```text
CTRL + C
```

Once the client stops, the public URL will no longer serve your local folder.

---

## ✨ Key Features

### 🚀 Unified Cobra CLI Engine

- Single powerful binary handles both the **Relay Server** and the **Local Client**.
- Built on [`spf13/cobra`](https://github.com/spf13/cobra), the same CLI framework used by tools like Kubernetes and Docker.
- Simple commands for server deployment and local folder exposure.

### 🌍 Multi-Tenant Subdomain Routing

- True reverse proxy architecture using HTTP `Host` headers.
- Multiple users can connect simultaneously using different subdomains.

Example:

```text
project-one.arnabpachal.site
portfolio.arnabpachal.site
movie-api.arnabpachal.site
```

- Built-in collision detection safely rejects duplicate subdomain requests.
- Automatically generates secure randomized subdomains when no subdomain is provided.

### 🔐 Zero-Trust Architecture & RSA Encryption

- Custom RSA encryption protects the tunnel data channel.
- `tunnel.key` stays strictly local.
- Client handshake authentication uses secure initialization tokens.
- Public users never directly access the local machine.
- Every request is verified before the relay bridges the socket.

### ⚡ Cloudflare Native HTTPS Integration

- Ready for production-grade SSL/TLS using Cloudflare Proxy.
- Supports Cloudflare Flexible SSL.
- CLI outputs clean HTTPS URLs for frictionless sharing.

Example:

```text
https://my-awesome-project.arnabpachal.site
```

### 🧠 Aggressive Memory Management & Concurrency

- Mutex-based busy locks prevent race conditions during large file or video transfers.
- Bidirectional teardown tripwires immediately free sockets when a browser tab closes.
- Drops unnecessary secondary requests when the tunnel is actively streaming data.
- Designed for stable long-running relay connections.

---

## 🏗️ Architecture Overview

```text
[ Public Browser ]
        |
        | HTTPS Request
        v
[ Cloudflare Proxy - SSL Termination ]
        |
        | HTTP / HTTPS Traffic
        | Host: my-api.arnabpachal.site
        v
=================================================
[ VPS Relay Server - Docker ]
-------------------------------------------------
1. Parses HTTP Host Header for Subdomain
2. Checks Active Subdomain Socket in Memory Map
3. Sends NEW_REQUEST Signal to Correct Local Client
4. Bridges Browser and Local File Server
=================================================
        ^                              |
        | Control Channel - Port 9000  | Data Channel - Encrypted on 9001
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

## 🐳 Production Deployment on VPS

Docker is the recommended way to deploy the **FolderExposer Relay Server** on a custom VPS such as Hostinger, AWS EC2, DigitalOcean, Linode, or any Linux server.

This configuration maps standard web traffic to the internal Go relay server.

```bash
docker run -d \
  --name folder-exposer \
  --restart unless-stopped \
  -p 80:8080 \
  -p 443:8080 \
  -p 9000:9000 \
  -p 9001:9001 \
  arnabpachal/folder-exposer:latest
```

## 🐳 Docker Port Mapping

| Host Port | Container Port | Purpose |
|---:|---:|---|
| `80` | `8080` | Public HTTP access |
| `443` | `8080` | HTTPS traffic forwarded by proxy/CDN setup |
| `9000` | `9000` | Persistent control channel |
| `9001` | `9001` | Encrypted data channel |

## 🔎 Docker Management Commands

### Check Running Container

```bash
docker ps
```

### View Logs

```bash
docker logs -f folder-exposer
```

### Stop Container

```bash
docker stop folder-exposer
```

### Start Container Again

```bash
docker start folder-exposer
```

### Remove Container

```bash
docker rm folder-exposer
```

### Remove Docker Image

```bash
docker rmi arnabpachal/folder-exposer:latest
```

---

## 🔒 Enabling HTTPS with Cloudflare

To get the secure padlock icon without complex Let's Encrypt setup inside Go, use Cloudflare.

### 1. Add Your Domain to Cloudflare

Add your domain, for example:

```text
arnabpachal.site
```

### 2. Create a Wildcard DNS Record

Create this DNS record:

| Type | Name | Value |
|---|---|---|
| `A` | `*` | Your VPS Public IP |

Example:

```text
*.arnabpachal.site → VPS_PUBLIC_IP
```

### 3. Enable Cloudflare Proxy

Make sure the proxy status is:

```text
Proxied - Orange Cloud
```

### 4. Set SSL Mode

Go to:

```text
SSL/TLS → Overview
```

Set encryption mode to:

```text
Flexible
```

### 5. Final Public URL Format

After setup, users can access exposed folders like:

```text
https://my-awesome-project.arnabpachal.site
```

---

## 🧑‍💻 Developer Setup

If you are a developer and want to build or run the project from source, follow these steps.

### 1. Clone the Repository

```bash
git clone https://github.com/Arnab-Pachal1234/FolderExposer.git
cd FolderExposer
```

### 2. Start the Relay Server Locally

```bash
go run ./cmd/folder-exposer server
```

### 3. Connect the Local Client

```bash
go run ./cmd/folder-exposer expose ./shared_folder --subdomain test-api
```

### 4. Build Binary Manually

```bash
go build -o folder-exposer ./cmd/folder-exposer
```

Windows build:

```bash
go build -o folder-exposer.exe ./cmd/folder-exposer
```

---

## 📦 Releases & Binaries

FolderExposer uses **GoReleaser** and **GitHub Actions** to automatically compile ready-to-use executables.

Users do not need to install Go. They can simply download the correct binary for their operating system from the GitHub Releases page.

### Linux / macOS

```bash
./folder-exposer expose ./my_folder --subdomain test
```

### Windows

```bash
folder-exposer.exe expose ./my_folder --subdomain test
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

## 🛠️ Tech Stack

| Technology | Purpose |
|---|---|
| Go | Core application logic and memory management |
| Cobra CLI | Unified command-line interface |
| Raw TCP Sockets | Relay and data channel communication |
| HTTP Host Routing | Subdomain-based multi-tenant routing |
| RSA Crypto | Tunnel data encryption |
| Docker | Production relay deployment |
| Alpine Linux | Ultra-lightweight container image |
| GoReleaser | Cross-platform release automation |
| GitHub Actions | Zero-touch CI/CD deployment pipeline |
| Cloudflare | HTTPS, wildcard DNS, and SSL termination |

---

## 🔐 Security Notes

FolderExposer is designed for **educational**, **experimental**, and **security architecture learning** purposes.

Before using it in a production environment with highly sensitive data, it is recommended to implement:

- Strong authentication tokens
- TLS/HTTPS encryption
- Rate limiting
- Request logging and monitoring
- Firewall rules for relay ports
- Abuse protection
- Domain-level access controls

> Public users should never directly access your local machine. FolderExposer follows a relay-based architecture where every request is verified before socket bridging.

---

## 🧪 Example Use Cases

- Share a local project demo with friends or testers.
- Expose a local folder without uploading files to cloud storage.
- Share static frontend builds temporarily.
- Test reverse tunneling and relay server architecture.
- Learn raw TCP socket routing in Go.
- Build a lightweight local tunneling tool for educational purposes.
- Demonstrate secure multi-tenant subdomain routing.

---

## ❓ FAQ

### Do my files get uploaded to the server?

No. Your files remain on your local machine. The relay only forwards requests and responses.

### Can I use my own domain?

Yes. You can use your own domain with wildcard DNS and Cloudflare proxy.

Example:

```text
*.yourdomain.com
```

### Can multiple users expose folders at the same time?

Yes. FolderExposer supports multi-tenant subdomain routing.

Example:

```text
user1.arnabpachal.site
user2.arnabpachal.site
project-demo.arnabpachal.site
```

### What happens if two users request the same subdomain?

FolderExposer safely rejects duplicate subdomain requests using collision detection.

### How do I stop exposing my folder?

Press:

```text
CTRL + C
```

The tunnel will close and the public URL will stop working.

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
