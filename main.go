package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// Configuration
const (
	// URL WebSocket de GeckoTerminal (AnyCable - compatible ActionCable)
	wsURL = "wss://cables.geckoterminal.com/cable"

	// Headers
	origin    = "https://www.geckoterminal.com"
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36"
)

// Message structures
type ActionCableMessage struct {
	Type       string `json:"type,omitempty"`
	Command    string `json:"command,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	Message    string `json:"message,omitempty"`
}

type ChannelIdentifier struct {
	Channel string `json:"channel"`
	PoolID  string `json:"pool_id,omitempty"` // Pool ID (numeric string like "147971598")
}

func main() {
	fmt.Println("🦎 CoinGecko Terminal WebSocket Client")
	fmt.Println("=" + string(make([]byte, 60)))

	// Setup interrupt handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Connect to WebSocket with headers
	log.Printf("Connexion à %s...", wsURL)

	headers := map[string][]string{
		"Origin":     {origin},
		"User-Agent": {userAgent},
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
	if err != nil {
		log.Fatal("Erreur de connexion:", err)
	}
	defer conn.Close()

	log.Println("✅ Connecté au WebSocket")

	// Channel for messages
	done := make(chan struct{})

	// Read messages goroutine
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Erreur de lecture:", err)
				return
			}

			timestamp := time.Now().Format("15:04:05")
			fmt.Printf("[%s] Message reçu: %s\n", timestamp, string(message))

			// Parse message type
			var msg ActionCableMessage
			if err := json.Unmarshal(message, &msg); err == nil {
				handleMessage(conn, msg)
			}
		}
	}()

	// Wait for welcome message
	time.Sleep(2 * time.Second)

	// Subscribe to pool and swaps (using pool_id from the network capture)
	// Pool ID 147971598 corresponds to ETH/USDC pool on Ethereum
	poolID := "147971598"
	subscribeToPoolChannel(conn, poolID)
	subscribeToSwapChannel(conn, poolID)

	// Wait for interrupt or done
	select {
	case <-done:
		log.Println("Connexion fermée")
	case <-interrupt:
		log.Println("Interruption reçue, fermeture...")

		// Cleanly close the connection
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("Erreur fermeture:", err)
			return
		}
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	}
}

func handleMessage(conn *websocket.Conn, msg ActionCableMessage) {
	switch msg.Type {
	case "welcome":
		log.Println("📨 Message de bienvenue reçu")

	case "ping":
		// Respond to ping with pong
		pong := ActionCableMessage{
			Type: "pong",
		}
		if err := conn.WriteJSON(pong); err != nil {
			log.Println("Erreur pong:", err)
		}

	case "confirm_subscription":
		log.Printf("✅ Subscription confirmée: %s", msg.Identifier)

	case "reject_subscription":
		log.Printf("❌ Subscription rejetée: %s", msg.Identifier)
	}
}

func subscribeToPoolChannel(conn *websocket.Conn, poolID string) {
	log.Printf("📡 Subscription à PoolChannel pour pool_id=%s...", poolID)

	// Create channel identifier
	identifier := ChannelIdentifier{
		Channel: "PoolChannel",
		PoolID:  poolID,
	}

	identifierJSON, _ := json.Marshal(identifier)

	// Create subscribe command
	subscribeMsg := ActionCableMessage{
		Command:    "subscribe",
		Identifier: string(identifierJSON),
	}

	if err := conn.WriteJSON(subscribeMsg); err != nil {
		log.Println("Erreur subscription PoolChannel:", err)
		return
	}

	log.Println("✅ Commande PoolChannel envoyée")
}

func subscribeToSwapChannel(conn *websocket.Conn, poolID string) {
	log.Printf("📡 Subscription à SwapChannel pour pool_id=%s...", poolID)

	// Create channel identifier
	identifier := ChannelIdentifier{
		Channel: "SwapChannel",
		PoolID:  poolID,
	}

	identifierJSON, _ := json.Marshal(identifier)

	// Create subscribe command
	subscribeMsg := ActionCableMessage{
		Command:    "subscribe",
		Identifier: string(identifierJSON),
	}

	if err := conn.WriteJSON(subscribeMsg); err != nil {
		log.Println("Erreur subscription SwapChannel:", err)
		return
	}

	log.Println("✅ Commande SwapChannel envoyée")
}
