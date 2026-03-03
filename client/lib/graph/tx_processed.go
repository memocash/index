package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func WaitForTx(ctx context.Context, graphUrl, txHash string) error {
	wsUrl := strings.Replace(graphUrl, "https://", "wss://", 1)
	wsUrl = strings.Replace(wsUrl, "http://", "ws://", 1)
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		Subprotocols:     []string{"graphql-transport-ws"},
	}
	conn, _, err := dialer.Dial(wsUrl, nil)
	if err != nil {
		return fmt.Errorf("error connecting websocket for wait tx; %w", err)
	}
	defer conn.Close()
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	if err := conn.WriteJSON(map[string]string{"type": "connection_init"}); err != nil {
		return fmt.Errorf("error sending connection_init for wait tx; %w", err)
	}
	var ack wsMessage
	if err := conn.ReadJSON(&ack); err != nil {
		return fmt.Errorf("error reading connection_ack for wait tx; %w", err)
	}
	if ack.Type != "connection_ack" {
		return fmt.Errorf("error expected connection_ack, got %s", ack.Type)
	}
	subscribeMsg := map[string]interface{}{
		"id":   "1",
		"type": "subscribe",
		"payload": map[string]interface{}{
			"query":     `subscription ($hash: Hash!) { tx(hash: $hash) { hash } }`,
			"variables": map[string]string{"hash": txHash},
		},
	}
	if err := conn.WriteJSON(subscribeMsg); err != nil {
		return fmt.Errorf("error sending subscribe for wait tx; %w", err)
	}
	for {
		var msg wsMessage
		if err := conn.ReadJSON(&msg); err != nil {
			return fmt.Errorf("error reading message for wait tx; %w", err)
		}
		switch msg.Type {
		case "next":
			return nil
		case "error":
			return fmt.Errorf("error subscription error for wait tx: %s", string(msg.Payload))
		case "complete":
			return nil
		}
	}
}

type wsMessage struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}
