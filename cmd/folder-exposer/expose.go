package main

import (
	"crypto/tls" // <-- ADDED
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Arnab-Pachal1234/FolderExposer/pkg/tunnel"
	"github.com/spf13/cobra"
)

var (
	password  string
	serverIP  string
	subdomain string
	localPort int
)

var exposeCmd = &cobra.Command{
	Use:   "expose [directory]",
	Short: "Exposes a local directory to the relay server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		folderToExpose := args[0]
		fmt.Println("[Client] Initializing Authenticated Local Agent...")

		os.MkdirAll(folderToExpose, os.ModePerm)
		fmt.Printf("[Client] Preparing to expose folder: %s\n", folderToExpose)

		localAddr := fmt.Sprintf(":%d", localPort)
		go startInternalFileServer(folderToExpose, localAddr)

		controlAddr := fmt.Sprintf("%s:9000", serverIP)
		conn, err := net.Dial("tcp", controlAddr)
		if err != nil {
			fmt.Printf("Connection error to server %s: %v\n", controlAddr, err)
			return
		}
		defer conn.Close()

		initMsg := tunnel.Message{
			Type:      "INIT",
			Payload:   subdomain,
			AuthToken: "dev_secure_99",
			Password:  password,
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

		finalSubdomain := resp.Payload

		publicURL := fmt.Sprintf("http://%s.%s:8080", finalSubdomain, serverIP)

		fmt.Println("\n=======================================================")
		fmt.Println("🚀 SUCCESS! YOUR FOLDER IS LIVE ON THE INTERNET")
		fmt.Println("=======================================================")
		fmt.Printf("🌍 Public URL: %s\n", publicURL)
		fmt.Println("=======================================================")
		fmt.Println("Listening for incoming connections. Press CTRL+C to stop.")

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
					fmt.Println("\n[DEBUG-CLIENT] VPS requested data! Opening encrypted tunnel bridge...")

					dataAddr := fmt.Sprintf("%s:9001", serverIP)

					// --- UPGRADED TO TLS ---
					tlsConfig := &tls.Config{
						InsecureSkipVerify: true,
					}
					vpsDataConn, err1 := tls.Dial("tcp", dataAddr, tlsConfig)
					// -----------------------

					if err1 != nil {
						fmt.Printf("[DEBUG-CLIENT] Error connecting securely to VPS %s: %v\n", dataAddr, err1)
					}

					targetApp := fmt.Sprintf("localhost:%d", localPort)
					localApp, err2 := net.Dial("tcp", targetApp)
					if err2 != nil {
						fmt.Printf("[DEBUG-CLIENT] Error connecting to Local File Server 8081: %v\n", err2)
					}

					if vpsDataConn != nil && localApp != nil {
						fmt.Println("[DEBUG-CLIENT] Sockets established. Streaming encrypted data bidirectionally...")

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
							close(done)
							fmt.Println("[DEBUG-CLIENT] Local sockets cleanly released.")
						}()
					}
				}
			}
		}()

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := tunnel.WriteJSON(conn, tunnel.Message{Type: "PING"}); err != nil {
				fmt.Println("[Client] Broken tunnel pipe detected.")
				break
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(exposeCmd)

	exposeCmd.Flags().StringVarP(&serverIP, "server", "s", "arnabpachal.site", "The IP address of your VPS relay")
	exposeCmd.Flags().StringVarP(&subdomain, "subdomain", "d", "", "Requested subdomain (leave blank for random)")
	exposeCmd.Flags().IntVarP(&localPort, "local-port", "l", 8081, "Local port for the internal file server")
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
