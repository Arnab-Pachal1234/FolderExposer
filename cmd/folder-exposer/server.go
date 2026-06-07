package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
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
	Password    string
	IsBusy      bool
	mu          sync.Mutex
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
		fmt.Printf("[Server] Secured Cloud Relay listening on port %s...\n", controlAddr)

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
}

func (s *RelayServer) handleClient(conn net.Conn) {
	var initialMessage tunnel.Message
	if err := tunnel.ReadJSON(conn, &initialMessage); err != nil {
		conn.Close()
		return
	}

	if initialMessage.AuthToken != SecureSystemToken {
		fmt.Printf("[Server] Security Alert: Unauthorized access attempt using token: %s\n", initialMessage.AuthToken)
		tunnel.WriteJSON(conn, tunnel.Message{Type: "AUTH_FAILURE", Payload: "Invalid Auth Token Provided"})
		conn.Close()
		return
	}

	tunnelID := initialMessage.Payload
	folderPassword := initialMessage.Password

	s.mu.Lock()
	s.tunnels[tunnelID] = &ActiveTunnel{
		ControlConn: conn,
		Password:    folderPassword,
		IsBusy:      false,
	}
	s.mu.Unlock()

	fmt.Printf("[Server] Authentication Successful. Tunnel '%s' is now online.\n", tunnelID)
	tunnel.WriteJSON(conn, tunnel.Message{Type: "INIT_ACK", Payload: "Secure tunnel verified and open"})

	defer func() {
		s.mu.Lock()
		delete(s.tunnels, tunnelID)
		s.mu.Unlock()
		conn.Close()
		fmt.Printf("[Server] Client '%s' disconnected. Routing path removed immediately.\n", tunnelID)
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
	dataListener, _ := net.Listen("tcp", dataPort)
	dataConns := make(chan net.Conn)

	go func() {
		for {
			conn, _ := dataListener.Accept()
			dataConns <- conn
		}
	}()

	fmt.Printf("[Server] Public Proxy listening on %s. Data Channel on %s...\n", publicPort, dataPort)

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("\n[DEBUG-SERVER] New public request from browser for path: %s\n", req.URL.Path)

		s.mu.Lock()
		tunnelInfo, exists := s.tunnels["arnab-dev-tunnel"]
		s.mu.Unlock()

		if !exists {
			fmt.Println("[DEBUG-SERVER] Rejecting: Tunnel offline.")
			http.Error(w, "Tunnel is currently offline.", http.StatusNotFound)
			return
		}

		_, pass, ok := req.BasicAuth()
		if !ok || pass != tunnelInfo.Password {
			fmt.Println("[DEBUG-SERVER] Rejecting: Incorrect or missing password.")
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted Area"`)
			http.Error(w, "Unauthorized Access", http.StatusUnauthorized)
			return
		}

		tunnelInfo.mu.Lock()
		if tunnelInfo.IsBusy {
			tunnelInfo.mu.Unlock()
			fmt.Printf("[DEBUG-SERVER] Rejecting %s: Tunnel is currently busy!\n", req.URL.Path)
			http.Error(w, "Resource currently in use by another user.", http.StatusServiceUnavailable)
			return
		}
		tunnelInfo.IsBusy = true
		tunnelInfo.mu.Unlock()
		fmt.Println("[DEBUG-SERVER] Door Locked. Securing connection for data transfer.")

		auditMsg := fmt.Sprintf(`{"time": "%s", "ip": "%s", "path": "%s"}`, time.Now().Format("15:04:05"), req.RemoteAddr, req.URL.Path)
		tunnel.WriteJSON(tunnelInfo.ControlConn, tunnel.Message{Type: "AUDIT_LOG", Payload: auditMsg})

		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		publicConn, _, _ := hijacker.Hijack()

		fmt.Println("[DEBUG-SERVER] Sending NEW_REQUEST signal to laptop...")
		tunnel.WriteJSON(tunnelInfo.ControlConn, tunnel.Message{Type: "NEW_REQUEST"})

		var laptopDataConn net.Conn
		select {
		case laptopDataConn = <-dataConns:
			fmt.Println("[DEBUG-SERVER] Laptop successfully connected to Data Port!")
		case <-time.After(5 * time.Second):
			fmt.Println("[DEBUG-SERVER] TIMEOUT ERROR: Laptop failed to connect in time. Unlocking door.")
			publicConn.Close()
			tunnelInfo.mu.Lock()
			tunnelInfo.IsBusy = false
			tunnelInfo.mu.Unlock()
			return
		}

		req.Header.Set("Connection", "close")
		req.Close = true
		req.Write(laptopDataConn)

		idleDuration := 60 * time.Second
		safePublicConn := &IdleTimeoutConn{Conn: publicConn, Timeout: idleDuration}
		safeLaptopConn := &IdleTimeoutConn{Conn: laptopDataConn, Timeout: idleDuration}

		go func() {
			done := make(chan string, 2)

			go func() {
				io.Copy(safePublicConn, safeLaptopConn)
				done <- "Laptop File Server finished sending data."
			}()
			go func() {
				io.Copy(safeLaptopConn, safePublicConn)
				done <- "Public Browser closed the tab/connection."
			}()

			reason := <-done
			fmt.Printf("[DEBUG-SERVER] Transfer stopped! Reason: %s\n", reason)

			publicConn.Close()
			laptopDataConn.Close()

			tunnelInfo.mu.Lock()
			tunnelInfo.IsBusy = false
			tunnelInfo.mu.Unlock()
			fmt.Println("[DEBUG-SERVER] Sockets wiped. Door is UNLOCKED for the next request.")
		}()
	})

	http.ListenAndServe(publicPort, nil)
}
