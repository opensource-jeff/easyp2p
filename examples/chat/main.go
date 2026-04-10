package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"easyp2p"
)

func main() {
	ctx := context.Background()

	// 1. Initialize the node with default settings
	fmt.Println("Starting chat node...")
	node := easyp2p.Must(easyp2p.NewNode(ctx, easyp2p.DefaultConfig()))
	defer node.Close()

	node.PrintDescribe()

	// 2. Wait for the network to be ready
	fmt.Println("Waiting for network...")
	if err := node.WaitForNetwork(ctx, 30*time.Second); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	// 3. Join a chat room (topic)
	fmt.Println("Joining 'irc-general' chat room...")
	topic, err := node.JoinTopic("irc-general")
	if err != nil {
		panic(err)
	}
	defer topic.Close()

	// 4. Set up message handler
	topic.OnMessage(func(msg string, sender string) {
		fmt.Printf("\n[%s]: %s\n> ", sender[:8], msg)
	})

	// 5. Input loop for sending messages
	fmt.Println("Chat ready! Type your message and press Enter.")
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		if text == "/quit" {
			break
		}

		if err := topic.Publish(text); err != nil {
			fmt.Printf("Error sending: %v\n", err)
		}
		fmt.Print("> ")
	}
}
