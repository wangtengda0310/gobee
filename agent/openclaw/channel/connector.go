package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// Connector bridges local OpenClaw Gateway to remote reverse server
type Connector struct {
	localGatewayURL  string
	remoteServerURL  string
	token            string

	localConn  *websocket.Conn
	remoteConn *websocket.Conn
	localMu    sync.Mutex
	remoteMu   sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// HTTP client for health checks
	httpClient *http.Client
}

// NewConnector creates a new connector
func NewConnector(localGatewayURL, remoteServerURL, token string) *Connector {
	ctx, cancel := context.WithCancel(context.Background())
	return &Connector{
		localGatewayURL: localGatewayURL,
		remoteServerURL: remoteServerURL,
		token:           token,
		ctx:             ctx,
		cancel:          cancel,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Start starts the connector
func (c *Connector) Start() error {
	log.Printf("Starting connector...")
	log.Printf("  Local Gateway: %s", c.localGatewayURL)
	log.Printf("  Remote Server: %s", c.remoteServerURL)

	// Connect to local gateway first
	log.Println("Connecting to local OpenClaw Gateway...")
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	localConn, _, err := dialer.Dial(c.localGatewayURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to local gateway: %w", err)
	}
	c.localMu.Lock()
	c.localConn = localConn
	c.localMu.Unlock()
	log.Println("Connected to local Gateway")

	// Connect to remote reverse server
	log.Println("Connecting to remote reverse server...")
	remoteConn, _, err := dialer.Dial(c.remoteServerURL, nil)
	if err != nil {
		localConn.Close()
		return fmt.Errorf("failed to connect to remote server: %w", err)
	}
	c.remoteMu.Lock()
	c.remoteConn = remoteConn
	c.remoteMu.Unlock()
	log.Println("Connected to remote server")

	// Start bidirectional relay
	c.wg.Add(2)
	go c.relayLocalToRemote()
	go c.relayRemoteToLocal()

	log.Println("Connector started successfully - bridging connections")

	// Wait for shutdown
	<-c.ctx.Done()
	c.wg.Wait()

	return nil
}

// relayLocalToRemote forwards messages from local gateway to remote server
func (c *Connector) relayLocalToRemote() {
	defer c.wg.Done()
	defer c.cancel()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		c.localMu.Lock()
		localConn := c.localConn
		c.localMu.Unlock()

		if localConn == nil {
			return
		}

		_, data, err := localConn.ReadMessage()
		if err != nil {
			log.Printf("Error reading from local gateway: %v", err)
			return
		}

		log.Printf("[Local->Remote] %s", truncate(string(data), 200))

		c.remoteMu.Lock()
		remoteConn := c.remoteConn
		c.remoteMu.Unlock()

		if remoteConn == nil {
			return
		}

		if err := remoteConn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Error writing to remote server: %v", err)
			return
		}
	}
}

// relayRemoteToLocal forwards messages from remote server to local gateway
func (c *Connector) relayRemoteToLocal() {
	defer c.wg.Done()
	defer c.cancel()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		c.remoteMu.Lock()
		remoteConn := c.remoteConn
		c.remoteMu.Unlock()

		if remoteConn == nil {
			return
		}

		_, data, err := remoteConn.ReadMessage()
		if err != nil {
			log.Printf("Error reading from remote server: %v", err)
			return
		}

		log.Printf("[Remote->Local] %s", truncate(string(data), 200))

		c.localMu.Lock()
		localConn := c.localConn
		c.localMu.Unlock()

		if localConn == nil {
			return
		}

		if err := localConn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Error writing to local gateway: %v", err)
			return
		}
	}
}

// Stop stops the connector
func (c *Connector) Stop() {
	c.cancel()

	c.localMu.Lock()
	if c.localConn != nil {
		c.localConn.Close()
	}
	c.localMu.Unlock()

	c.remoteMu.Lock()
	if c.remoteConn != nil {
		c.remoteConn.Close()
	}
	c.remoteMu.Unlock()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// RunConnector runs the connector with the given configuration
func RunConnector(config *Config) error {
	// Determine the remote URL - ExternalAddress should be the full WebSocket URL
	remoteURL := config.API.ExternalAddress
	if remoteURL == "" {
		return fmt.Errorf("remote server URL is required")
	}

	connector := NewConnector(
		config.Gateway.URL,
		remoteURL,
		config.Gateway.Token,
	)

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		errCh <- connector.Start()
	}()

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		log.Printf("Received signal %v, shutting down...", sig)
		connector.Stop()
		return nil
	}
}
