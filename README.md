# YouTube Remote Control

A simple client-server application that allows you to control YouTube playback from another computer on the same network.

## Prerequisites

- Go 1.16 or higher
- Google Chrome browser
- YouTube should be open and active in Chrome

## Getting Started

1. Start the server on the computer that has YouTube open:
```bash
go run cmd/server/main.go
```

2. Start the client on another computer (update the server address in the client code if needed):
```bash
go run cmd/client/main.go
```

## Available Commands

- `play`: Play/Resume video
- `pause`: Pause video
- `next`: Next video
- `previous`: Previous video
- `volumeUp`: Increase volume
- `volumeDown`: Decrease volume
- `quit`: Exit the client application

## Network Setup

By default, the server listens on port 8080. Make sure this port is accessible between the client and server machines.
