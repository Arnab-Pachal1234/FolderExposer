package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Arnab-Pachal1234/FolderExposer/pkg/tunnel"
	"github.com/spf13/cobra"
)

const SecureSystemToken = "dev_secure_99"

var (
	publicPort  int
	controlPort int
	dataPort    int
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
	serverCmd.Flags().IntVarP(&dataPort, "data-port", "d", 9001, "Port for the ephemeral encrypted data sockets")
	// Removed the rootDomain flag entirely as Cloudflare handles HTTPS now!
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

// THE UPGRADE: Generates an unbreakable TLS certificate completely in RAM!
func generateEphemeralTLSCert() (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"FolderExposer Auto-TLS"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  priv,
	}, nil
}

func (s *RelayServer) startSecureGateway(publicPort string, dataPort string) {
	// 1. Generate the certificate in memory (NO HARD DRIVE REQUIRED!)
	cert, err := generateEphemeralTLSCert()
	if err != nil {
		fmt.Printf("[FATAL] Could not generate ephemeral TLS certificate: %v\n", err)
		return
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	// 2. Start the encrypted listener on port 9001
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

		hijacker, ok := w.(http.Hijacker)
		if !ok {
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

	// Clean, stripped-down HTTP listener. Cloudflare handles the HTTPS encryption for the public!
	fmt.Printf("[Server] Public Gateway ready on port %s\n", publicPort)
	if err := http.ListenAndServe(publicPort, nil); err != nil {
		fmt.Printf("[FATAL] Public server crashed: %v\n", err)
	}
}

func generateRandomSubdomain() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
