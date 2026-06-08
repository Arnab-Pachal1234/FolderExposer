package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync" // <-- ADDED
	"time"

	"github.com/Arnab-Pachal1234/FolderExposer/pkg/tunnel"
	"github.com/spf13/cobra"
)

var (
	vpsIP      string
	rootDomain string
	subdomain  string
	localPort  int
)

var exposeCmd = &cobra.Command{
	Use:   "expose [directory]",
	Short: "Exposes a local directory to the encrypted relay server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		folderToExpose := args[0]
		fmt.Println("[Client] Initializing Authenticated Local Agent...")

		os.MkdirAll(folderToExpose, os.ModePerm)
		fmt.Printf("[Client] Preparing to expose folder: %s\n", folderToExpose)

		localAddr := fmt.Sprintf(":%d", localPort)
		go startInternalFileServer(folderToExpose, localAddr)

		controlAddr := fmt.Sprintf("%s:9000", vpsIP)
		conn, err := net.Dial("tcp", controlAddr)
		if err != nil {
			fmt.Printf("\n[FATAL] Connection error to VPS %s: %v\n", controlAddr, err)
			return
		}
		defer conn.Close()

		initMsg := tunnel.Message{
			Type:      "INIT",
			Payload:   subdomain,
			AuthToken: "dev_secure_99",
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
		publicURL := fmt.Sprintf("https://%s.%s", finalSubdomain, rootDomain)

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

				if msg.Type == "NEW_REQUEST" {
					fmt.Println("\n[DEBUG-CLIENT] VPS requested data! Opening TLS encrypted tunnel bridge...")

					dataAddr := fmt.Sprintf("%s:9001", vpsIP)
					tlsConfig := &tls.Config{InsecureSkipVerify: true}

					// Robust retry logic for the data channel
					var vpsDataConn *tls.Conn
					var err1 error
					for i := 0; i < 3; i++ {
						vpsDataConn, err1 = tls.Dial("tcp", dataAddr, tlsConfig)
						if err1 == nil {
							break
						}
						time.Sleep(500 * time.Millisecond)
					}

					if err1 != nil {
						fmt.Printf("[DEBUG-CLIENT] Error connecting securely: %v\n", err1)
						continue
					}

					localApp, err2 := net.Dial("tcp", fmt.Sprintf("localhost:%d", localPort))
					if err2 != nil {
						vpsDataConn.Close()
						continue
					}

					fmt.Println("[DEBUG-CLIENT] TLS Sockets established. Streaming...")

					go func() {
						var wg sync.WaitGroup
						wg.Add(2)
						go func() { defer wg.Done(); io.Copy(vpsDataConn, localApp) }()
						go func() { defer wg.Done(); io.Copy(localApp, vpsDataConn) }()
						wg.Wait()
						vpsDataConn.Close()
						localApp.Close()
						fmt.Println("[DEBUG-CLIENT] Bridge torn down cleanly.")
					}()
				}
			}
		}()

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := tunnel.WriteJSON(conn, tunnel.Message{Type: "PING"}); err != nil {
				break
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(exposeCmd)
	exposeCmd.Flags().StringVarP(&vpsIP, "server", "s", "187.127.142.6", "The raw IP address of your VPS relay")
	exposeCmd.Flags().StringVarP(&rootDomain, "domain", "r", "arnabpachal.site", "The root domain")
	exposeCmd.Flags().StringVarP(&subdomain, "subdomain", "d", "", "Requested subdomain")
	exposeCmd.Flags().IntVarP(&localPort, "local-port", "l", 8081, "Local port")
}

func startInternalFileServer(folderPath string, port string) {
	absolutePath, _ := filepath.Abs(folderPath)
	fs := http.FileServer(http.Dir(absolutePath))
	http.Handle("/", fs)
	http.ListenAndServe(port, nil)
}
