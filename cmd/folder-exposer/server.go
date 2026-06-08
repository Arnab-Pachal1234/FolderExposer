package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Arnab-Pachal1234/FolderExposer/pkg/tunnel"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/acme/autocert"
)

const SecureSystemToken = "dev_secure_99"

var (
	publicPort  int
	controlPort int
	dataPort    int
	rootDomain  string
)

type IdleTimeoutConn struct {
	net.Conn
	Timeout time.Duration
}

func (i *IdleTimeoutConn) Read(b []byte) (int, error) {
	i.Conn.SetReadDeadline(time.Now().Add(i.Timeout))
	return i.Conn.Read(b)
}

func (i *IdleTimeoutConn) Write(b []byte) (int, error) {
	i.Conn.SetWriteDeadline(time.Now().Add(i.Timeout))
	return i.Conn.Write(b)
}

// STRIPPED DOWN: No more passwords, no more locks!
type ActiveTunnel struct {
	ControlConn net.Conn
}

type RelayServer struct {
	tunnels map[string]*ActiveTunnel
	mu      sync.Mutex
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the public VPS relay server",
	Run: func(cmd *cobra.Command, args []string) {
		server := &RelayServer{tunnels: make(map[string]*ActiveTunnel)}

		controlAddr := fmt.Sprintf(":%d", controlPort)
		listener, err := net.Listen("tcp", controlAddr)
		if err != nil {
			fmt.Printf("Server boot failure: %v\n", err)
			return
		}
		defer listener.Close()
		fmt.Printf("[Server] High-Speed Cloud Relay listening on port %s...\n", controlAddr)

		go server.startSecureGateway(fmt.Sprintf(":%d", publicPort), fmt.Sprintf(":%d", dataPort))

		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			go server.handleClient(conn)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVarP(&publicPort, "public-port", "p", 8080, "Port for public web traffic")
	serverCmd.Flags().IntVarP(&controlPort, "control-port", "c", 9000, "Port for the local client heartbeat")
	serverCmd.Flags().IntVarP(&dataPort, "data-port", "d", 9001, "Port for the ephemeral data sockets")
	serverCmd.Flags().StringVar(&rootDomain, "domain", "", "Your base domain for Auto-TLS (e.g., arnabpachal.site)")
}

func (s *RelayServer) handleClient(conn net.Conn) {
	var initialMessage tunnel.Message
	if err := tunnel.ReadJSON(conn, &initialMessage); err != nil {
		conn.Close()
		return
	}

	if initialMessage.AuthToken != SecureSystemToken {
		conn.Close()
		return
	}

	requestedSubdomain := initialMessage.Payload

	s.mu.Lock()
	if requestedSubdomain == "" {
		requestedSubdomain = generateRandomSubdomain()
	}

	if _, exists := s.tunnels[requestedSubdomain]; exists {
		s.mu.Unlock()
		tunnel.WriteJSON(conn, tunnel.Message{Type: "AUTH_FAILURE", Payload: "Subdomain already in use."})
		conn.Close()
		return
	}

	s.tunnels[requestedSubdomain] = &ActiveTunnel{
		ControlConn: conn,
	}
	s.mu.Unlock()

	fmt.Printf("[Server] Tunnel '%s' is now online and open to the public.\n", requestedSubdomain)

	tunnel.WriteJSON(conn, tunnel.Message{Type: "INIT_ACK", Payload: requestedSubdomain})

	defer func() {
		s.mu.Lock()
		delete(s.tunnels, requestedSubdomain)
		s.mu.Unlock()
		conn.Close()
		fmt.Printf("[Server] Client '%s' disconnected. Path removed.\n", requestedSubdomain)
	}()

	for {
		var msg tunnel.Message
		if err := tunnel.ReadJSON(conn, &msg); err != nil {
			break
		}
		if msg.Type == "PING" {
			tunnel.WriteJSON(conn, tunnel.Message{Type: "PONG"})
		}
	}
}

func (s *RelayServer) startSecureGateway(publicPort string, dataPort string) {
	cert, err := tls.LoadX509KeyPair("/root/certs/tunnel.crt", "/root/certs/tunnel.key")
	if err != nil {
		fmt.Printf("[FATAL] Could not load internal tunnel keys: %v\n", err)
		return
	}

	// 2. Configure the server to require TLS
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	// 3. UPGRADE: Listen using TLS instead of raw TCP
	dataListener, err := tls.Listen("tcp", dataPort, tlsConfig)
	if err != nil {
		fmt.Printf("[FATAL] Failed to start encrypted data channel: %v\n", err)
		return
	}

	dataConns := make(chan net.Conn)

	go func() {
		for {
			conn, err := dataListener.Accept()
			if err != nil {
				fmt.Printf("[Server] Rejected unencrypted data connection attempt.\n")
				continue
			}
			dataConns <- conn
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		host := req.Host
		requestedSubdomain := strings.Split(host, ".")[0]

		s.mu.Lock()
		tunnelInfo, exists := s.tunnels[requestedSubdomain]
		s.mu.Unlock()

		if !exists {
			http.Error(w, "Tunnel is offline", http.StatusNotFound)
			return
		}

		// NO MORE BASIC AUTH OR BUSY LOCKS!
		// Traffic flows freely and concurrently.

		hijacker, ok := w.(http.Hijacker)
		if !ok {
			// This will never trigger again because we disabled HTTP/2 below!
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		publicConn, _, _ := hijacker.Hijack()

		tunnel.WriteJSON(tunnelInfo.ControlConn, tunnel.Message{Type: "NEW_REQUEST"})

		var laptopDataConn net.Conn
		select {
		case laptopDataConn = <-dataConns:
		case <-time.After(5 * time.Second):
			publicConn.Close()
			return
		}

		req.Header.Set("Connection", "close")
		req.Close = true
		req.Write(laptopDataConn)

		idleDuration := 60 * time.Second
		safePublicConn := &IdleTimeoutConn{Conn: publicConn, Timeout: idleDuration}
		safeLaptopConn := &IdleTimeoutConn{Conn: laptopDataConn, Timeout: idleDuration}

		go func() {
			go io.Copy(safePublicConn, safeLaptopConn)
			io.Copy(safeLaptopConn, safePublicConn)
			publicConn.Close()
			laptopDataConn.Close()
		}()
	})

	if rootDomain != "" {
		fmt.Printf("[Server] Enterprise TLS Enabled for domain: %s\n", rootDomain)

		certManager := &autocert.Manager{
			Prompt: autocert.AcceptTOS,
			Cache:  autocert.DirCache("certs"),
			HostPolicy: func(ctx context.Context, host string) error {
				if host == rootDomain || strings.HasSuffix(host, "."+rootDomain) {
					return nil
				}
				return fmt.Errorf("acme/autocert: host not allowed: %s", host)
			},
		}

		server := &http.Server{
			Addr: ":443",
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
			},
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),

			// ADD THIS LINE: Discard background TLS handshake noise
			ErrorLog: log.New(io.Discard, "", 0),
		}

		go http.ListenAndServe(":80", certManager.HTTPHandler(nil))

		if err := server.ListenAndServeTLS("", ""); err != nil {
			fmt.Printf("TLS Server crashed: %v\n", err)
		}
	} else {
		http.ListenAndServe(publicPort, nil)
	}
}

func generateRandomSubdomain() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
