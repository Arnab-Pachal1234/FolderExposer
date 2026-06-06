package tunnel

import (
	"fmt"
	"net"
)

// TunnelSession tracks an active connection link between client and server
type TunnelSession struct {
	ID          string
	ControlConn net.Conn
}

// PrintStatus is a shared helper function
func PrintStatus(msg string) {
	fmt.Printf("[Tunnel Core] %s\n", msg)
}
