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
	wsURL     = "wss://cables.geckoterminal.com/cable"
	origin    = "https://www.geckoterminal.com"
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
)

// Pools to monitor (same as aggregator benchmark)
var monitoredPools = []struct {
	Name    string
	Network string
	PoolID  string
	Chain   string // For metrics
}{
	{
		Name:    "ETH/USDC Uniswap V3",
		Network: "eth",
		PoolID:  "147971598",
		Chain:   "ethereum",
	},
	{
		Name:    "SOL/USDC Raydium",
		Network: "solana",
		PoolID:  "162715608",
		Chain:   "solana",
	},
	{
		Name:    "WETH/USDC Base",
		Network: "base",
		PoolID:  "162840764",
		Chain:   "base",
	},
	{
		Name:    "WBNB/BUSD PancakeSwap",
		Network: "bsc",
		PoolID:  "24",
		Chain:   "bnb",
	},
	{
		Name:    "WETH/USDC Arbitrum",
		Network: "arbitrum",
		PoolID:  "162634438",
		Chain:   "arbitrum",
	},
}

// Message structures
type ActionCableMessage struct {
	Type       string          `json:"type,omitempty"`
	Command    string          `json:"command,omitempty"`
	Identifier string          `json:"identifier,omitempty"`
	Message    json.RawMessage `json:"message,omitempty"`
}

type ChannelIdentifier struct {
	Channel string `json:"channel"`
	PoolID  string `json:"pool_id,omitempty"`
}

// Swap data from SwapChannel
type SwapData struct {
	Data struct {
		BlockTimestamp     int64   `json:"block_timestamp"`      // On-chain timestamp (ms)
		TxHash             string  `json:"tx_hash"`
		FromTokenAmount    string  `json:"from_token_amount"`
		ToTokenAmount      string  `json:"to_token_amount"`
		PriceFromInUSD     string  `json:"price_from_in_usd"`
		PriceToInUSD       string  `json:"price_to_in_usd"`
		FromTokenTotalUSD  string  `json:"from_token_total_in_usd"`
		ToTokenTotalUSD    string  `json:"to_token_total_in_usd"`
		TxFromAddress      string  `json:"tx_from_address"`
		FromTokenID        int     `json:"from_token_id"`
		ToTokenID          int     `json:"to_token_id"`
	} `json:"data"`
	Type string `json:"type"` // "newSwap"
}

// Head lag statistics
type HeadLagStats struct {
	Count      int
	TotalMs    int64
	MinMs      int64
	MaxMs      int64
	LastSwap   time.Time
	LastLagMs  int64
}

var stats = make(map[string]*HeadLagStats) // chain -> stats

func main() {
	fmt.Println("🦎 GeckoTerminal Head Lag Monitor")
	fmt.Println("=" + string(make([]byte, 60)))
	fmt.Println("Measuring indexation latency for GeckoTerminal WebSocket")
	fmt.Printf("Monitoring %d pools across 5 chains\n", len(monitoredPools))
	fmt.Println()

	// Initialize stats
	for _, pool := range monitoredPools {
		stats[pool.Chain] = &HeadLagStats{
			MinMs: 999999,
		}
	}

	// Setup interrupt handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Print stats every 30 seconds
	statsTicker := time.NewTicker(30 * time.Second)
	defer statsTicker.Stop()

	statsQuit := make(chan struct{})
	go func() {
		for {
			select {
			case <-statsTicker.C:
				printStats()
			case <-statsQuit:
				return
			}
		}
	}()

	// Auto-reconnect loop
	reconnectDelay := 5 * time.Second
	maxReconnectDelay := 60 * time.Second

	for {
		select {
		case <-interrupt:
			log.Println("\nInterrupt received, closing...")
			close(statsQuit)
			printStats()
			return
		default:
			err := connectAndMonitor(interrupt)
			if err != nil {
				log.Printf("⚠️  Connection lost: %v", err)
				log.Printf("🔄 Reconnecting in %v...", reconnectDelay)

				select {
				case <-interrupt:
					close(statsQuit)
					printStats()
					return
				case <-time.After(reconnectDelay):
					// Exponential backoff
					reconnectDelay = reconnectDelay * 2
					if reconnectDelay > maxReconnectDelay {
						reconnectDelay = maxReconnectDelay
					}
				}
			} else {
				// Clean disconnect, reset delay
				reconnectDelay = 5 * time.Second
			}
		}
	}
}

func connectAndMonitor(interrupt chan os.Signal) error {
	log.Printf("Connecting to %s...", wsURL)

	headers := map[string][]string{
		"Origin":     {origin},
		"User-Agent": {userAgent},
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
	if err != nil {
		return fmt.Errorf("dial error: %w", err)
	}
	defer conn.Close()

	log.Println("✅ Connected to WebSocket")

	// Channel for messages
	done := make(chan struct{})

	// Read messages goroutine
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}

			handleMessage(conn, message)
		}
	}()

	// Wait for welcome message
	time.Sleep(2 * time.Second)

	// Subscribe to SwapChannel for all monitored pools
	for _, pool := range monitoredPools {
		subscribeToSwapChannel(conn, pool.PoolID, pool.Name)
		time.Sleep(100 * time.Millisecond) // Small delay between subscriptions
	}

	log.Printf("✅ Subscribed to %d pools", len(monitoredPools))

	// Heartbeat (ping/pong)
	pingTicker := time.NewTicker(25 * time.Second)
	defer pingTicker.Stop()

	// Wait for interrupt or done
	for {
		select {
		case <-done:
			return fmt.Errorf("connection closed by server")
		case <-interrupt:
			// Cleanly close the connection
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
		case <-pingTicker.C:
			// Server sends pings, we respond with pongs (handled in handleMessage)
			// Just check connection is alive
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		}
	}
}

func handleMessage(conn *websocket.Conn, message []byte) {
	var msg ActionCableMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		return
	}

	switch msg.Type {
	case "welcome":
		log.Println("📨 Welcome message received")

	case "ping":
		// Respond to ping with pong
		pong := ActionCableMessage{
			Type: "pong",
		}
		conn.WriteJSON(pong)

	case "confirm_subscription":
		log.Printf("✅ Subscription confirmed: %s", msg.Identifier)

	case "reject_subscription":
		log.Printf("❌ Subscription rejected: %s", msg.Identifier)

	default:
		// Handle data messages
		if msg.Message != nil {
			handleDataMessage(msg.Identifier, msg.Message)
		}
	}
}

func handleDataMessage(identifier string, message json.RawMessage) {
	// Parse swap data
	var swapData SwapData
	if err := json.Unmarshal(message, &swapData); err != nil {
		return
	}

	if swapData.Type != "newSwap" {
		return
	}

	// Extract channel info to get pool
	var channelIdent ChannelIdentifier
	if err := json.Unmarshal([]byte(identifier), &channelIdent); err != nil {
		return
	}

	// Find which pool this is
	var poolChain string
	for _, pool := range monitoredPools {
		if pool.PoolID == channelIdent.PoolID {
			poolChain = pool.Chain
			break
		}
	}

	if poolChain == "" {
		return
	}

	// Calculate head lag
	receiveTime := time.Now().UTC()
	onChainTime := time.UnixMilli(swapData.Data.BlockTimestamp)
	lagMs := receiveTime.Sub(onChainTime).Milliseconds()
	lagSeconds := float64(lagMs) / 1000.0

	// Update stats
	updateStats(poolChain, lagMs)

	// Log (show all swaps for visibility)
	timestamp := receiveTime.Format("15:04:05")
	txHash := swapData.Data.TxHash
	if len(txHash) > 12 {
		txHash = txHash[:10] + "..."
	}

	volume := "N/A"
	if swapData.Data.FromTokenTotalUSD != "" && len(swapData.Data.FromTokenTotalUSD) > 2 {
		if len(swapData.Data.FromTokenTotalUSD) > 10 {
			volume = swapData.Data.FromTokenTotalUSD[:10]
		} else {
			volume = swapData.Data.FromTokenTotalUSD
		}
	}

	fmt.Printf("[%s][GECKO][%-10s] Lag: %6.2fs | Tx: %s | Vol: $%s\n",
		timestamp, poolChain, lagSeconds, txHash, volume)
}

func updateStats(chain string, lagMs int64) {
	s := stats[chain]
	s.Count++
	s.TotalMs += lagMs
	s.LastSwap = time.Now()
	s.LastLagMs = lagMs

	if lagMs < s.MinMs {
		s.MinMs = lagMs
	}
	if lagMs > s.MaxMs {
		s.MaxMs = lagMs
	}
}

func printStats() {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║         GECKOTERMINAL HEAD LAG STATISTICS                     ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════╣")

	for _, pool := range monitoredPools {
		s := stats[pool.Chain]

		if s.Count == 0 {
			fmt.Printf("║ %-12s │ No swaps received yet                      ║\n", pool.Chain)
			continue
		}

		avgMs := s.TotalMs / int64(s.Count)
		timeSinceLastSwap := time.Since(s.LastSwap)

		fmt.Printf("║ %-12s │ Swaps: %5d │ Avg: %4dms │ Min: %4dms │ Max: %5dms ║\n",
			pool.Chain, s.Count, avgMs, s.MinMs, s.MaxMs)
		fmt.Printf("║              │ Last: %4dms (%s ago)                       ║\n",
			s.LastLagMs, formatDuration(timeSinceLastSwap))
	}

	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}

func subscribeToSwapChannel(conn *websocket.Conn, poolID, poolName string) {
	log.Printf("📡 Subscribing to SwapChannel for %s (pool_id=%s)...", poolName, poolID)

	identifier := ChannelIdentifier{
		Channel: "SwapChannel",
		PoolID:  poolID,
	}

	identifierJSON, _ := json.Marshal(identifier)

	subscribeMsg := ActionCableMessage{
		Command:    "subscribe",
		Identifier: string(identifierJSON),
	}

	if err := conn.WriteJSON(subscribeMsg); err != nil {
		log.Printf("Error subscribing to %s: %v", poolName, err)
		return
	}
}
