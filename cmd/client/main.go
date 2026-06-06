package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/arnabpachal/localtunnel/pkg/tunnel"
)

func main() {
	fmt.Println("[Client] Initializing Authenticated Local Agent...")

	folderToExpose := "/home/arnabpanchal/Desktop/LocaltunnelChecker/myshared_folder"
	os.MkdirAll(folderToExpose, os.ModePerm)
	fmt.Printf("[Client] Preparing to expose folder: %s\n", folderToExpose)

	go startInternalFileServer(folderToExpose, ":8081")

	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		fmt.Printf("Connection error: %v\n", err)
		return
	}
	defer conn.Close()

	// FIX 1: Added the folder password to the initialization message
	initMsg := tunnel.Message{
		Type:      "INIT",
		Payload:   "arnab-dev-tunnel",
		AuthToken: "dev_secure_99",
		Password:  "secret_folder_124", // This is what Client 2 must type in the browser!
	}

	if err := tunnel.WriteJSON(conn, initMsg); err != nil {
		return
	}

	var resp tunnel.Message
	if err := tunnel.ReadJSON(conn, &resp); err != nil {
		fmt.Println("[Client] Connection refused by server.")
		return
	}

	if resp.Type == "AUTH_FAILURE" {
		fmt.Printf("[Client] Authentication Denied: %s\n", resp.Payload)
		return
	}

	fmt.Printf("[Client] Security Handshake Complete: %s\n", resp.Payload)

	// Background worker listening for VPS commands
	// Background worker listening for VPS commands
	go func() {
		for {
			var msg tunnel.Message
			err := tunnel.ReadJSON(conn, &msg)
			if err != nil {
				fmt.Println("[Client] Control channel disconnected.")
				break
			}

			if msg.Type == "PONG" {
				// Heartbeat verified
			} else if msg.Type == "AUDIT_LOG" {
				f, err := os.OpenFile("security_report.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err == nil {
					f.WriteString(msg.Payload + "\n")
					f.Close()
				}
			} else if msg.Type == "NEW_REQUEST" {
				fmt.Println("\n[DEBUG-CLIENT] VPS requested data! Opening tunnel bridge...")

				vpsDataConn, err1 := net.Dial("tcp", "localhost:9001")
				if err1 != nil {
					fmt.Printf("[DEBUG-CLIENT] Error connecting to VPS 9001: %v\n", err1)
				}

				localApp, err2 := net.Dial("tcp", "localhost:8081")
				if err2 != nil {
					fmt.Printf("[DEBUG-CLIENT] Error connecting to Local File Server 8081: %v\n", err2)
				}

				if vpsDataConn != nil && localApp != nil {
					fmt.Println("[DEBUG-CLIENT] Sockets established. Streaming data bidirectionally...")

					go func() {
						done := make(chan string, 2)

						go func() {
							io.Copy(vpsDataConn, localApp)
							done <- "Local File Server finished."
						}()
						go func() {
							io.Copy(localApp, vpsDataConn)
							done <- "VPS closed the pipe."
						}()

						reason := <-done
						fmt.Printf("[DEBUG-CLIENT] Bridge torn down. Reason: %s\n", reason)

						vpsDataConn.Close()
						localApp.Close()
						close(done) //here closing done
						fmt.Println("[DEBUG-CLIENT] Local sockets cleanly released.")
					}()
				}
			}
		}
	}()

	// Hold connection open with heartbeats (ONLY sending PINGs now)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := tunnel.WriteJSON(conn, tunnel.Message{Type: "PING"}); err != nil {
			fmt.Println("[Client] Broken tunnel pipe detected.")
			break
		}
	}
}

func startInternalFileServer(folderPath string, port string) {
	absolutePath, _ := filepath.Abs(folderPath)
	fs := http.FileServer(http.Dir(absolutePath))
	http.Handle("/", fs)

	fmt.Printf("[Client] Internal File Server running silently on localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("[Client] File server crashed: %v\n", err)
	}
}
