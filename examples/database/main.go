package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/opensource-jeff/easyp2p"
	"github.com/libp2p/go-libp2p/core/peer"
)

func main() {
	ctx := context.Background()

	// 1. Local database
	myDB := map[string]string{
		"golang":  "A powerful programming language.",
		"libp2p":  "A modular network stack for P2P apps.",
		"easyp2p": "A beginner friendly P2P library.",
	}

	// 2. Initialize the node
	node := easyp2p.Must(easyp2p.NewNode(ctx, easyp2p.DefaultConfig()))
	defer node.Close()

	node.PrintDescribe()

	// 3. Set protocol for database search
	const dbProtocol = "/easyp2p/db-search/1.0"

	// 4. Set handler for incoming searches
	node.HandleProtocol(dbProtocol, func(s *easyp2p.Stream) {
		defer s.Close()
		
		// Read search query
		queryBuf := make([]byte, 256)
		n, _ := s.Read(queryBuf)
		query := string(queryBuf[:n])
		
		fmt.Printf("\nSearching for: %s\n", query)
		
		// Respond with result if found
		if result, ok := myDB[query]; ok {
			s.Write([]byte(result))
		} else {
			s.Write([]byte("Not found!"))
		}
	})

	// 5. Wait for network (optional but good for visibility)
	go func() {
		fmt.Println("Attempting to find peers...")
		if err := node.WaitForNetwork(ctx, 30*time.Second); err != nil {
			fmt.Printf("Warning: %v\n", err)
		} else {
			fmt.Println("Network ready! Peers found.")
		}
	}()

	// 6. Input loop for searching other nodes
	fmt.Println("\nEnter a Peer ID and a keyword to search (e.g. [PeerID] golang)")
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		input := strings.Split(scanner.Text(), " ")
		if len(input) < 2 {
			fmt.Println("Usage: [PeerID] [Keyword]")
			continue
		}

		targetID, err := peer.Decode(input[0])
		if err != nil {
			fmt.Printf("Invalid Peer ID: %v\n", err)
			continue
		}
		keyword := input[1]

		fmt.Printf("Querying %s for '%s'...\n", targetID.String()[:8], keyword)
		
		// Connect and search
		stream, err := node.ConnectTo(targetID, dbProtocol)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		
		stream.Write([]byte(keyword))
		
		// Read result
		resultBuf := make([]byte, 512)
		n, _ := stream.Read(resultBuf)
		fmt.Printf("Result from node: %s\n", string(resultBuf[:n]))
		stream.Close()
		fmt.Print("> ")
	}
}
