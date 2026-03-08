package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// RunBatch processes messages in batch mode
func RunBatch(client *Client, message string, sessionKey string) error {
	// Connect to gateway
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// If no message provided, read from stdin
	if message == "" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return fmt.Errorf("no message provided and stdin is a terminal")
		}

		reader := bufio.NewReader(os.Stdin)
		var lines []string
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			lines = append(lines, strings.TrimSuffix(line, "\n"))
		}
		message = strings.Join(lines, "\n")
	}

	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("empty message")
	}

	// Send message and get response
	response, err := client.Chat(sessionKey, message)
	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}

	// Print response to stdout
	fmt.Println(response)
	return nil
}
