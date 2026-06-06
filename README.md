# 🛡️ Localtunnel-Go: Zero-Trust Local Folder Gateway


**Localtunnel-Go** is a secure reverse tunneling application built in Go that allows a local folder to be exposed over the internet while keeping all files stored on the local device.


It is designed for authenticated, secure, high-speed file and video transfer using a custom reverse tunnel architecture, raw TCP socket routing, HTTP protocol hijacking, security audit logging, health checks, and optimized memory management.


---


## 🚀 Project Overview


This project works as a private gateway between a public browser and a local file server running behind a NAT/firewall.


Instead of uploading files to a cloud storage provider, the data remains on the local machine. External users can access the folder only after proper authentication through a secure tunneling mechanism.


The system is built with a **Zero-Trust Network Access-style approach**, meaning every public request is authenticated, logged, and routed through a controlled tunnel.


---


## ✨ Key Features


### 🔐 Zero-Trust Access Control


- HTTP Basic Authentication for public users

- Client handshake authentication using secure initialization tokens

- Public users cannot directly access the local machine

- Every request is verified before being forwarded


---


### 🌐 Secure Reverse Tunneling


- Exposes a local folder to the internet without directly opening local ports

- Uses a VPS relay server as the public entry point

- Local client connects outward to the VPS, making it NAT/firewall friendly

- Keeps the actual data stored only on the local device


---


### ⚡ High-Speed File and Video Transfer


- Supports large file transfers

- Supports video streaming through raw TCP socket forwarding

- Uses `io.Copy` for efficient bidirectional data streaming

- Avoids unnecessary buffering for better performance


---


### 🧠 Aggressive Memory Management


- Automatically closes broken or idle connections

- Prevents zombie sockets and memory leaks

- Uses timeout-based connection cleanup

- Frees resources immediately after transfer completion


---


### 🧵 Concurrency Protection


- Uses a mutex-based busy-lock mechanism

- Allows one controlled transfer session at a time

- Prevents race conditions during large file or video transfer

- Drops unnecessary secondary requests when the tunnel is already busy


---


### 📊 Security Audit Logging


- Logs visitor IP addresses

- Logs requested paths

- Logs access time

- Sends audit telemetry from the VPS relay back to the local client

- Stores logs locally in `security_report.json`


---


### ❤️ Health Check and Monitoring


- Heartbeat mechanism using `PING/PONG`

- Detects disconnected clients

- Monitors tunnel availability

- Helps maintain stable long-running tunnel sessions


---


## 🏗️ Architecture Overview


```text

[ Public Browser ]

        |

        | HTTP Request + Basic Authentication

        v

=================================================

[ VPS Relay Server ]

-------------------------------------------------

1. Accepts public browser requests

2. Verifies authentication

3. Checks tunnel availability

4. Sends request signal to local client

5. Waits for data connection

6. Bridges browser and local file server

=================================================

        ^                              |

        | Control Channel              | Data Channel

        | Port 9000                    | Port 9001

        |                              v

=================================================

[ Local Go Client ]

-------------------------------------------------

1. Maintains secure control connection

2. Sends heartbeat signals

3. Receives relay instructions

4. Opens temporary data socket

5. Connects to local file server

6. Writes audit logs locally

=================================================

        |

        v

[ Local File Server ]

Port 8081

```


---


## 🔌 Port Usage


| Port | Component | Purpose |

|---|---|---|

| `8080` | Public Proxy Server | Public browser access point |

| `9000` | Control Channel | Persistent client-server tunnel control |

| `9001` | Data Channel | Temporary high-speed data transfer socket |

| `8081` | Local File Server | Serves local folder privately |


---


## 📁 Suggested Project Structure


```text

localtunnel-go/

│

├── cmd/

│   ├── server/

│   │   └── main.go

│   │

│   └── client/

│       └── main.go

│

├── internal/

│   ├── auth/

│   ├── tunnel/

│   ├── proxy/

│   ├── logger/

│   └── health/

│

├── shared_folder/

│   └── example files

│

├── security_report.json

├── go.mod

├── go.sum

└── README.md

```


---


## ⚙️ Prerequisites


Before running this project, make sure you have:


- Go `1.20` or higher

- A Linux VPS or local machine for testing

- Basic knowledge of terminal commands

- Required ports opened on the VPS firewall


---


---

## 🐳 Docker Deployment

You can run the **FolderExposer Relay Server** instantly using the official Docker image published on Docker Hub.

Docker image:

```text
arnabpachal/folder-exposer:latest
```

This is useful when you want to deploy the public Relay Server on a VPS/cloud server without manually building the Go server.

---

### 1. Pull the Docker Image

```bash
docker pull arnabpachal/folder-exposer:latest
```

---

### 2. Run the Relay Server with Docker

Run this command on your VPS or cloud server:

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

### 3. Docker Port Mapping

| Host Port | Container Port | Purpose |
|---|---|---|
| `80` | `8080` | Public browser access without typing a port |
| `9000` | `9000` | Persistent Control Channel |
| `9001` | `9001` | Ephemeral Data Channel for high-speed file/video transfer |

> **Note:** Port `80` is mapped to container port `8080`, so users can access the public gateway using only the VPS IP address, for example `http://YOUR_VPS_PUBLIC_IP`.

---

### 4. Start the Local Client

On the machine containing the folder you want to expose, download the latest client binary from the project **Releases** tab.

You can also run the client directly using Go:

```bash
go run ./cmd/client/main.go
```

Expected output:

```text
[Client] Preparing to expose folder: /myshared_folder
[Client] Security Handshake Complete: Secure tunnel verified and open
```

---

### 5. Access the Shared Folder

Open any browser and visit:

```text
http://YOUR_VPS_PUBLIC_IP
```

You will be asked for authentication through the browser's native login prompt.

Example credentials:

```text
Username: anything
Password: secret_folder_123
```

After successful authentication, the Relay Server will securely forward the request to your local client through the tunnel.

---

## 📦 Releases

This project also provides release builds so users can download the latest binaries directly instead of building everything manually.

Users can download the latest client/server binaries from the GitHub **Releases** tab and run the required executable for their operating system.

Example usage after downloading the client binary:

```bash
./folder-exposer-client
```

Or, on Windows:

```powershell
folder-exposer-client.exe
```


## 🚀 Getting Started


### 1. Clone the Repository


```bash

git clone git@github.com:Arnab-Pachal1234/FolderExposer.git

cd FolderExposer

```


---


### 2. Start the Relay Server


Run this command on the VPS or public machine:


```bash

go run ./cmd/server/main.go

```


Expected output:


```text

[Server] Secured Cloud Relay listening on port :9000

[Server] Public Proxy listening on port :8080

[Server] Data Channel listening on port :9001

```


---


### 3. Start the Local Client


Run this command on your local machine:


```bash

go run ./cmd/client/main.go

```


Expected output:


```text

[Client] Preparing to expose folder: ./shared_folder

[Client] Internal file server running on localhost:8081

[Client] Security handshake completed successfully

[Client] Secure tunnel is now active

```


---


### 4. Access the Shared Folder


Open your browser and visit:


```text

http://localhost:8080

```


Or, if running on a VPS:


```text

http://YOUR_VPS_PUBLIC_IP:8080

```


You will be asked for authentication.


Example credentials:


```text

Username: anything

Password: secret_folder_123

```


---


## 🔐 Authentication Flow


```text

Browser Request

      |

      v

HTTP Basic Auth Check

      |

      v

Relay Server Validates Access

      |

      v

Relay Signals Local Client

      |

      v

Local Client Opens Data Tunnel

      |

      v

File Transfer Starts

```


Only authenticated users are allowed to access the exposed folder.


---


## 📊 Security Audit Logging


Every successful authenticated request is logged and sent back to the local client.


Example `security_report.json`:


```json

{"time":"14:32:05","ip":"192.168.1.45:54321","path":"/"}

{"time":"14:45:12","ip":"10.0.0.9:65432","path":"/secret_video.mp4"}

```


This helps track:


- Who accessed the tunnel

- When they accessed it

- Which file or path they requested


---


## 🧠 Memory and Connection Management


The project includes aggressive socket cleanup techniques to avoid memory leaks.


### Implemented safety mechanisms:


- EOF detection

- Idle timeout cleanup

- Broken socket detection

- Bidirectional teardown

- Dead connection removal

- Automatic resource release


When either side of the connection closes, the tunnel immediately tears down the paired socket.


---


## 🧵 Concurrency Control


To avoid unstable parallel transfers, the relay server uses a busy-lock mechanism.


```text

Request 1: Accepted

Request 2: Dropped if tunnel is busy

Request 3: Dropped if tunnel is busy

```


This is useful when transferring large files or streaming videos, where multiple background browser requests may otherwise interfere with the main transfer.


---


## ❤️ Health Check System


The client and server use heartbeat messages to verify tunnel health.


Example:


```text

Client -> Server: PING

Server -> Client: PONG

```


If the heartbeat fails, the server can detect that the local client is disconnected.


---


## 🛠️ Debugging


The system prints detailed logs for observing tunnel behavior.


Example server logs:


```text

[DEBUG] Public browser connected

[DEBUG] Authentication successful

[DEBUG] Tunnel is available

[DEBUG] Signal sent to local client

[DEBUG] Data socket connected

[DEBUG] Starting bidirectional stream

[DEBUG] EOF detected, closing sockets

```


Example client logs:


```text

[Client] Received tunnel request

[Client] Opening data channel

[Client] Connected to local file server

[Client] Streaming data

[Client] Transfer completed

```


---


## 🧪 Testing Locally


You can test the entire system on your own machine using multiple terminals.


### Terminal 1: Start Server


```bash

go run ./cmd/server/main.go

```


### Terminal 2: Start Client


```bash

go run ./cmd/client/main.go

```


### Terminal 3: Access Public Proxy


```bash

curl -u user:secret_folder_123 http://localhost:8080

```


---


## 🧩 Technologies Used


- Go

- Raw TCP sockets

- HTTP server

- HTTP Basic Authentication

- Goroutines

- Mutex locking

- Bidirectional `io.Copy`

- JSON logging

- Custom timeout connection wrappers


---


## 📌 Use Cases


- Securely sharing local folders over the internet

- Private file transfer without uploading to cloud storage

- Temporary public access to local files

- Secure remote file browsing

- Learning reverse tunneling architecture

- Understanding Go networking and socket management


---


## ⚠️ Security Notes


This project is designed for educational and experimental security architecture learning.


For production usage, you should add:


- HTTPS/TLS support

- Strong token rotation

- Rate limiting

- IP allowlisting

- Brute-force protection

- Encrypted audit logs

- More advanced session management


---


## 👨‍💻 Author


**Arnab Pachal**


Built as a deep-dive project into:


- Low-level networking

- Secure reverse tunneling

- Zero-trust access design

- Go concurrency

- Socket lifecycle management

- High-speed file transfer systems


---



## ⭐ Final Note


Localtunnel-Go demonstrates how a secure reverse tunnel can expose local resources to the internet without moving the actual data to the cloud.


The main goal of this project is to explore how secure tunneling, authentication, socket forwarding, telemetry logging, and memory-safe connection handling can work together in a real-world networking system.


this was the prevous 