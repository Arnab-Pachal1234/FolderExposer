package tunnel

import (
	"encoding/json"
	"net"
)

type Message struct {
	Type      string `json:"type"`    // "INIT", "INIT_ACK", "PING", "PONG"
	Payload   string `json:"payload"` // The requested subdomain string
	AuthToken string `json:"auth_token"`
	Password  string `json:"password,omitempty"` // The developer's secure key
}

func WriteJSON(conn net.Conn, msg Message) error {
	return json.NewEncoder(conn).Encode(msg)
}

func ReadJSON(conn net.Conn, msg *Message) error {
	return json.NewDecoder(conn).Decode(msg)
}
