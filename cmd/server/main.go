package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"strings"

	"github.com/gorilla/websocket"
)

type Command struct {
	Action string `json:"action"`
	Value  string `json:"value,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

const (
	serverPort = 8080
)

func main() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting YouTube Remote Control server...")



	// Create a router for our web server
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/", fs)

	// WebSocket endpoint
	mux.HandleFunc("/ws", handleWebSocket)

	// Start HTTP/WebSocket server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", serverPort),
		Handler: mux,
	}

	networkIPs := getNetworkIPs()
	log.Printf("\nYouTube Remote Control is running!\n")
	log.Printf("\nAccess from your devices using any of these addresses:\n")
	for _, ip := range networkIPs {
		log.Printf("http://%s:%d\n", ip, serverPort)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
	}()

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal("Failed to start web server:", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection: %v\n", err)
		return
	}
	defer conn.Close()

	log.Printf("New WebSocket connection from %s\n", conn.RemoteAddr())

	// Send welcome message
	welcomeMsg := map[string]string{"status": "connected"}
	if err := conn.WriteJSON(welcomeMsg); err != nil {
		log.Printf("Failed to send welcome message: %v\n", err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v\n", err)
			break
		}

		log.Printf("Received message: %s\n", string(message))

		var cmd Command
		if err := json.Unmarshal(message, &cmd); err != nil {
			log.Printf("Failed to parse command: %v\n", err)
			// Send error back to client
			conn.WriteJSON(map[string]string{"error": "invalid command format"})
			continue
		}

		log.Printf("Executing command: %+v\n", cmd)
		if err := executeCommand(cmd); err != nil {
			log.Printf("Error executing command: %v\n", err)
			conn.WriteJSON(map[string]string{"error": err.Error()})
			continue
		}

		// Send confirmation back to client
		conn.WriteJSON(map[string]string{"status": "command executed"})
	}
}

func executeCommand(cmd Command) error {
	var err error
	switch cmd.Action {
	case "play", "pause":
		err = sendKeyPress("k")
	case "next":
		err = sendKeyPress("l")
	case "previous":
		err = sendKeyPress("j")
	case "volumeUp":
		err = sendKeyPress("up")
	case "volumeDown":
		err = sendKeyPress("down")
	case "open":
		if cmd.Value != "" {
			err = openYouTubeURL(cmd.Value)
		} else {
			err = fmt.Errorf("no URL provided")
		}
	}

	if err != nil {
		log.Printf("Error executing command %s: %v\n", cmd.Action, err)
		return err
	}
	log.Printf("Successfully executed command: %s\n", cmd.Action)
	return nil
}

func sendKeyPress(key string) error {
	if runtime.GOOS == "darwin" {
		script := fmt.Sprintf(`
			tell application "Google Chrome"
				activate
				delay 0.1
				tell application "System Events"
					keystroke "%s"
				end tell
			end tell
		`, key)
		cmd := exec.Command("osascript", "-e", script)
		return cmd.Run()
	}
	return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
}

func openYouTubeURL(url string) error {
	if runtime.GOOS == "darwin" {
		script := fmt.Sprintf(`
			tell application "Google Chrome"
				activate
				open location "%s"
			end tell
		`, url)
		cmd := exec.Command("osascript", "-e", script)
		return cmd.Run()
	}
	return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
}

func getNetworkIPs() []string {
	var ips []string

	// Always include localhost for local access
	ips = append(ips, "localhost")

	// Get all network interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Failed to get network interfaces: %v\n", err)
		return ips
	}

	// For each interface
	for _, iface := range ifaces {
		// Skip loopback and inactive interfaces
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Get addresses for this interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// For each address
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				// Skip loopback and non-IPv4
				ipv4 := v.IP.To4()
				if ipv4 == nil || ipv4.IsLoopback() {
					continue
				}

				// Only include private network IPs
				ip := ipv4.String()
				if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.") {
					log.Printf("Found network interface: %s (%s) - %s\n", iface.Name, iface.HardwareAddr, ip)
					ips = append(ips, ip)
				}
			}
		}
	}

	return ips
}
