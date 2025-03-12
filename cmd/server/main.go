package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os/exec"
	"runtime"
)

type Command struct {
	Action string `json:"action"`
	Value  string `json:"value,omitempty"`
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
	defer listener.Close()

	fmt.Println("Server listening on :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	var cmd Command

	for {
		err := decoder.Decode(&cmd)
		if err != nil {
			return
		}

		executeCommand(cmd)
	}
}

func executeCommand(cmd Command) {
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
	}

	if err != nil {
		log.Printf("Error executing command %s: %v\n", cmd.Action, err)
	} else {
		log.Printf("Successfully executed command: %s\n", cmd.Action)
	}
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
