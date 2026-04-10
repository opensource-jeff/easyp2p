package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"easyp2p"
)

func main() {
	ctx := context.Background()

	// 1. Initialize the node with default settings
	fmt.Println("Starting chat node...")
	node, err := easyp2p.NewNode(ctx, easyp2p.DefaultConfig())
	if err != nil {
		panic(err)
	}
	defer node.Close()

	fmt.Printf("My Peer ID: %s\n", node.ID())
	fmt.Printf("Listening on: %v\n", node.Addrs())

	// 2. Join a chat room (topic)
	fmt.Println("Joining 'irc-general' chat room...")
	topic, err := node.JoinTopic("irc-general")
	if err != nil {
		panic(err)
	}
	defer topic.Close()

	// 3. Set up message handler
	topic.OnMessage(func(msg string, sender string) {
		fmt.Printf("\n[%s]: %s\n> ", sender[:8], msg)
	})

	// 4. Input loop for sending messages
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
