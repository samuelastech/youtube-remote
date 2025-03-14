package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type Command struct {
	Action string `json:"action"`
	Value  string `json:"value,omitempty"`
}

func main() {
	serverAddr := flag.String("server", "localhost:8080", "server address in format host:port")
	flag.Parse()

	conn, err := net.Dial("tcp", *serverAddr)
	if err != nil {
		log.Fatal("Failed to connect to server:", err)
	}
	defer conn.Close()

	fmt.Println("Connected to YouTube Remote Control Server")
	fmt.Println("Available commands:")
	fmt.Println("- play: Play/Resume video")
	fmt.Println("- pause: Pause video")
	fmt.Println("- next: Next video")
	fmt.Println("- previous: Previous video")
	fmt.Println("- volumeUp: Increase volume")
	fmt.Println("- volumeDown: Decrease volume")
	fmt.Println("- open <url>: Open YouTube URL")
	fmt.Println("- quit: Exit the application")

	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(conn)

	for {
		fmt.Print("\nEnter command: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "quit" {
			break
		}

		parts := strings.SplitN(input, " ", 2)
		action := parts[0]

		var cmd Command
		if action == "open" && len(parts) > 1 {
			cmd = Command{Action: action, Value: parts[1]}
		} else {
			cmd = Command{Action: action}
		}

		err := encoder.Encode(cmd)
		if err != nil {
			log.Printf("Error sending command: %v\n", err)
			continue
		}
	}
}
