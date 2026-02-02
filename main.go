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
	// URL WebSocket de GeckoTerminal (ActionCable)
	wsURL = "wss://www.geckoterminal.com/cable"
)

// Message structures
type ActionCableMessage struct {
	Type       string `json:"type,omitempty"`
	Command    string `json:"command,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	Message    string `json:"message,omitempty"`
}

type ChannelIdentifier struct {
	Channel     string `json:"channel"`
	PoolAddress string `json:"pool_address,omitempty"`
	NetworkID   string `json:"network_id,omitempty"`
}

func main() {
	fmt.Println("🦎 CoinGecko Terminal WebSocket Client")
	fmt.Println("=" + string(make([]byte, 60)))

	// Setup interrupt handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Connect to WebSocket
	log.Printf("Connexion à %s...", wsURL)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
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

	// Subscribe to a pool (ETH/USDC Uniswap V3 on Ethereum)
	subscribeToPool(conn, "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640", "eth")

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

func subscribeToPool(conn *websocket.Conn, poolAddress, network string) {
	log.Printf("📡 Subscription au pool %s sur %s...", poolAddress, network)

	// Create channel identifier
	identifier := ChannelIdentifier{
		Channel:     "PoolChannel",
		PoolAddress: poolAddress,
	}

	identifierJSON, _ := json.Marshal(identifier)

	// Create subscribe command
	subscribeMsg := ActionCableMessage{
		Command:    "subscribe",
		Identifier: string(identifierJSON),
	}

	if err := conn.WriteJSON(subscribeMsg); err != nil {
		log.Println("Erreur subscription:", err)
		return
	}

	log.Println("✅ Commande de subscription envoyée")
}
