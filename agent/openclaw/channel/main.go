package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	gatewayURL string
	gatewayToken string

	// Batch mode flags
	batchSession string
	batchMessage string

	// API mode flags
	apiPort int
	apiHost string

	// MCP mode flags
	mcpTransport string

	// Enhanced mode flags
	useEnhanced bool
	configFile  string

	// Connector mode flags
	remoteServerURL string
)

var rootCmd = &cobra.Command{
	Use:   "openclaw-channel",
	Short: "OpenClaw Channel - A Go-based channel for OpenClaw interaction",
	Long: `OpenClaw Channel is a Go implementation that provides multiple modes
for interacting with OpenClaw gateway:

- Interactive mode: Real-time command-line chat
- Batch mode: Process messages from stdin or arguments
- Web RESTful API mode: HTTP endpoints for integration
- MCP mode: Model Context Protocol server for AI assistants`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&gatewayURL, "gateway", "g", "ws://127.0.0.1:18789", "OpenClaw gateway WebSocket URL")
	rootCmd.PersistentFlags().StringVarP(&gatewayToken, "token", "t", "", "OpenClaw gateway auth token")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Configuration file path (YAML or JSON)")
}

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Start interactive chat mode",
	Long:  "Start an interactive command-line chat session with OpenClaw",
	Aliases: []string{"i", "chat"},
	Run: func(cmd *cobra.Command, args []string) {
		client := NewClient(gatewayURL, gatewayToken)
		if err := RunInteractive(client); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var batchCmd = &cobra.Command{
	Use:   "batch [message]",
	Short: "Send a message in batch mode",
	Long: `Send a single message to OpenClaw and print the response.
If message is not provided as argument, reads from stdin.

Examples:
  openclaw-channel batch "Hello, how are you?"
  echo "What time is it?" | openclaw-channel batch
  cat questions.txt | openclaw-channel batch --session my-session`,
	Aliases: []string{"b", "send"},
	Run: func(cmd *cobra.Command, args []string) {
		client := NewClient(gatewayURL, gatewayToken)
		message := batchMessage
		if len(args) > 0 {
			message = args[0]
		}
		if err := RunBatch(client, message, batchSession); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start RESTful API server",
	Long: `Start an HTTP server that provides RESTful endpoints for OpenClaw interaction.

Endpoints:
  POST /chat         - Send a message (body: {"message": "...", "session": "..."})
  GET  /health       - Health check
  GET  /status       - Gateway status
  WS   /ws           - WebSocket passthrough`,
	Aliases: []string{"a", "serve", "server"},
	Run: func(cmd *cobra.Command, args []string) {
		client := NewClient(gatewayURL, gatewayToken)
		if err := RunAPIServer(client, apiHost, apiPort); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP (Model Context Protocol) server",
	Long: `Start an MCP server that exposes OpenClaw as tools for AI assistants.

Available tools:
  openclaw_chat    - Send a message to OpenClaw and get a response
  openclaw_status  - Get the current status of the OpenClaw gateway`,
	Aliases: []string{"m"},
	Run: func(cmd *cobra.Command, args []string) {
		client := NewClient(gatewayURL, gatewayToken)
		if err := RunMCPServer(client, mcpTransport); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var apiEnhancedCmd = &cobra.Command{
	Use:   "api-enhanced",
	Short: "Start enhanced RESTful API server with auto-reconnect",
	Long: `Start an enhanced HTTP server with advanced features:

Features:
  - Auto-reconnect on disconnection
  - Heartbeat/keep-alive
  - Connection state monitoring
  - Graceful shutdown
  - Client statistics

Endpoints:
  POST /chat         - Send a message (body: {"message": "...", "session": "..."})
  GET  /health       - Health check with connection state
  GET  /status       - Gateway status
  GET  /stats        - Client statistics
  WS   /ws           - WebSocket passthrough`,
	Aliases: []string{"ae", "serve-enhanced"},
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		config := GetConfigManager().Get()

		// Load from config file if specified
		if configFile != "" {
			if err := GetConfigManager().LoadFromFile(configFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}
			config = GetConfigManager().Get()
		}

		// Override with command-line flags
		if gatewayURL != "ws://127.0.0.1:18789" {
			config.Gateway.URL = gatewayURL
		}
		if gatewayToken != "" {
			config.Gateway.Token = gatewayToken
		}
		if apiHost != "0.0.0.0" {
			config.API.Host = apiHost
		}
		if apiPort != 8080 {
			config.API.Port = apiPort
		}

		// Load from environment
		if err := GetConfigManager().LoadFromEnv(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: error loading env config: %v\n", err)
		}

		if err := RunEnhancedAPIServer(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var reverseCmd = &cobra.Command{
	Use:   "reverse",
	Short: "Start reverse connection server (for public deployment)",
	Long: `Start a reverse connection server that waits for OpenClaw Gateway to connect.

This mode is designed for deploying on a public server (e.g., itsnot.fun).
The OpenClaw Gateway running on an internal network will connect to this server.

Architecture:
  [External Clients] --> [This Server (public)] <-- [OpenClaw Gateway (internal)]

The Gateway initiates the WebSocket connection to this server, enabling
communication from external clients to the internal OpenClaw instance.

Endpoints:
  GET  /gateway - WebSocket endpoint for Gateway connection
  GET  /health  - Health check
  GET  /status  - Gateway status (requires connected Gateway)
  POST /chat    - Send a message to OpenClaw
  GET  /stats   - Server statistics`,
	Aliases: []string{"r", "serve-reverse"},
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		config := GetConfigManager().Get()

		// Load from config file if specified
		if configFile != "" {
			if err := GetConfigManager().LoadFromFile(configFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}
			config = GetConfigManager().Get()
		}

		// Override with command-line flags
		if gatewayToken != "" {
			config.Gateway.Token = gatewayToken
		}
		if apiHost != "0.0.0.0" {
			config.API.Host = apiHost
		}
		if apiPort != 8080 {
			config.API.Port = apiPort
		}

		// Load from environment
		if err := GetConfigManager().LoadFromEnv(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: error loading env config: %v\n", err)
		}

		server := NewReverseServer(config)
		if err := server.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var connectorCmd = &cobra.Command{
	Use:   "connector",
	Short: "Start connector (bridge local Gateway to remote reverse server)",
	Long: `Start a connector that bridges the local OpenClaw Gateway to a remote reverse server.

This mode is designed to run on the same machine as the OpenClaw Gateway.
It connects to both the local Gateway and the remote reverse server, bridging
them together.

Architecture:
  [Local OpenClaw Gateway] <-- [Connector] --> [Remote Reverse Server]

This allows the remote server to communicate with your local OpenClaw instance.

Usage:
  1. Start the reverse server on the public server:
     openclaw-channel reverse --port 18770 --token YOUR_TOKEN

  2. Run this connector on your local machine:
     openclaw-channel connector --remote ws://itsnot.fun:18770/gateway --token YOUR_TOKEN`,
	Aliases: []string{"c", "bridge"},
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		config := GetConfigManager().Get()

		// Load from config file if specified
		if configFile != "" {
			if err := GetConfigManager().LoadFromFile(configFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}
			config = GetConfigManager().Get()
		}

		// Override with command-line flags
		if gatewayURL != "ws://127.0.0.1:18789" {
			config.Gateway.URL = gatewayURL
		}
		if gatewayToken != "" {
			config.Gateway.Token = gatewayToken
		}
		if remoteServerURL != "" {
			config.API.ExternalAddress = remoteServerURL
		}

		// Load from environment
		if err := GetConfigManager().LoadFromEnv(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: error loading env config: %v\n", err)
		}

		if config.API.ExternalAddress == "" {
			fmt.Fprintf(os.Stderr, "Error: remote server URL is required. Use --remote flag or set OPENCLAW_REMOTE_URL environment variable.\n")
			os.Exit(1)
		}

		if err := RunConnector(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	// Batch command flags
	batchCmd.Flags().StringVarP(&batchSession, "session", "s", "agent:main:main", "Session key to use")
	batchCmd.Flags().StringVarP(&batchMessage, "message", "m", "", "Message to send (reads from stdin if not provided)")

	// API command flags
	apiCmd.Flags().IntVarP(&apiPort, "port", "p", 8080, "API server port")
	apiCmd.Flags().StringVarP(&apiHost, "host", "H", "0.0.0.0", "API server host")

	// Enhanced API command flags
	apiEnhancedCmd.Flags().IntVarP(&apiPort, "port", "p", 8080, "API server port")
	apiEnhancedCmd.Flags().StringVarP(&apiHost, "host", "H", "0.0.0.0", "API server host")

	// MCP command flags
	mcpCmd.Flags().StringVarP(&mcpTransport, "transport", "T", "stdio", "MCP transport (stdio or sse)")

	// Reverse command flags
	reverseCmd.Flags().IntVarP(&apiPort, "port", "p", 8080, "Reverse server port")
	reverseCmd.Flags().StringVarP(&apiHost, "host", "H", "0.0.0.0", "Reverse server host")

	// Connector command flags
	connectorCmd.Flags().StringVarP(&remoteServerURL, "remote", "r", "", "Remote reverse server URL (e.g., ws://itsnot.fun:18770/gateway)")

	rootCmd.AddCommand(interactiveCmd)
	rootCmd.AddCommand(batchCmd)
	rootCmd.AddCommand(apiCmd)
	rootCmd.AddCommand(apiEnhancedCmd)
	rootCmd.AddCommand(reverseCmd)
	rootCmd.AddCommand(connectorCmd)
	rootCmd.AddCommand(mcpCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
