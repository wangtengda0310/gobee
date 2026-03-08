package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// RunInteractive starts an interactive chat session
func RunInteractive(client *Client) error {
	// Connect to gateway
	fmt.Println("Connecting to OpenClaw gateway...")
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	fmt.Println("Connected! Type your message and press Enter to send.")
	fmt.Println("Commands: /quit, /exit, /status, /clear, /help")
	fmt.Println()

	// Get initial status
	status, err := client.GetStatus()
	if err == nil && status != nil {
		fmt.Printf("Gateway: %s\n", client.url)
		if status.Gateway != nil {
			fmt.Printf("  Running: %v, Version: %s\n", status.Gateway.Running, status.Gateway.Version)
		}
		fmt.Println()
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start event printer
	go printEvents(client)

	reader := bufio.NewReader(os.Stdin)
	sessionKey := "agent:main:main"

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(input, "/") {
			switch strings.ToLower(input) {
			case "/quit", "/exit", "/q":
				fmt.Println("Goodbye!")
				return nil
			case "/status":
				printStatus(client)
				continue
			case "/clear":
				fmt.Print("\033[H\033[2J")
				continue
			case "/help":
				printHelp()
				continue
			case "/history":
				printHistory(client, sessionKey)
				continue
			default:
				if strings.HasPrefix(input, "/session ") {
					sessionKey = strings.TrimSpace(strings.TrimPrefix(input, "/session "))
					fmt.Printf("Switched to session: %s\n", sessionKey)
					continue
				}
				fmt.Printf("Unknown command: %s\n", input)
				continue
			}
		}

		// Send message
		response, err := client.Chat(sessionKey, input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		fmt.Println()
		fmt.Println(response)
		fmt.Println()
	}

	return nil
}

func printEvents(client *Client) {
	for event := range client.Events() {
		switch event.Event {
		case "chat":
			// Chat events are handled in Chat() method
		case "system-presence":
			// Presence updates
		default:
			// Log other events if needed
		}
	}
}

func printStatus(client *Client) {
	status, err := client.GetStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting status: %v\n", err)
		return
	}

	fmt.Println("=== Gateway Status ===")
	if status.Gateway != nil {
		fmt.Printf("Running: %v\n", status.Gateway.Running)
		fmt.Printf("Reachable: %v\n", status.Gateway.Reachable)
		fmt.Printf("Host: %s\n", status.Gateway.Host)
		fmt.Printf("Port: %d\n", status.Gateway.Port)
		fmt.Printf("Version: %s\n", status.Gateway.Version)
	}

	if len(status.Sessions) > 0 {
		fmt.Println("\n=== Sessions ===")
		for _, s := range status.Sessions {
			fmt.Printf("  %s [%s] - %s\n", s.Key, s.Kind, s.Model)
		}
	}
	fmt.Println()
}

func printHistory(client *Client, sessionKey string) {
	history, err := client.GetHistory(sessionKey, 10)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting history: %v\n", err)
		return
	}

	fmt.Println("=== Recent Messages ===")
	if payload, ok := history.(map[string]interface{}); ok {
		if messages, ok := payload["messages"].([]interface{}); ok {
			for _, item := range messages {
				fmt.Printf("%v\n", item)
			}
		}
	} else {
		fmt.Printf("%v\n", history)
	}
	fmt.Println()
}

func printHelp() {
	fmt.Println("=== Commands ===")
	fmt.Println("  /quit, /exit, /q  - Exit interactive mode")
	fmt.Println("  /status           - Show gateway status")
	fmt.Println("  /history          - Show recent messages")
	fmt.Println("  /session <key>    - Switch to a different session")
	fmt.Println("  /clear            - Clear the screen")
	fmt.Println("  /help             - Show this help")
	fmt.Println()
	fmt.Println("Just type your message and press Enter to chat with OpenClaw.")
	fmt.Println()
}
