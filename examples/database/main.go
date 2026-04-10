package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"easyp2p"
	"github.com/libp2p/go-libp2p/core/peer"
)

func main() {
	ctx := context.Background()

	// 1. Local database
	myDB := map[string]string{
		"golang": "A powerful programming language.",
		"libp2p": "A modular network stack for P2P apps.",
		"easyp2p": "A beginner friendly P2P library.",
	}

	// 2. Initialize the node
	node, err := easyp2p.NewNode(ctx, easyp2p.DefaultConfig())
	if err != nil {
		panic(err)
	}
	defer node.Close()

	fmt.Printf("My Peer ID: %s\n", node.ID())

	// 3. Set protocol for database search
	const dbProtocol = "/easyp2p/db-search/1.0"

	// 4. Set handler for incoming searches
	node.HandleProtocol(dbProtocol, func(s *easyp2p.Stream) {
		defer s.Close()
		
		// Read search query
		queryBuf := make([]byte, 256)
		n, _ := s.Read(queryBuf)
		query := string(queryBuf[:n])
		
		fmt.Printf("Searching for: %s\n", query)
		
		// Respond with result if found
		if result, ok := myDB[query]; ok {
			s.Write([]byte(result))
		} else {
			s.Write([]byte("Not found!"))
		}
	})

	// 5. Input loop for searching other nodes
	fmt.Println("\nEnter a Peer ID and a keyword to search (e.g. [PeerID] golang)")
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		input := strings.Split(scanner.Text(), " ")
		if len(input) < 2 {
			fmt.Println("Usage: [PeerID] [Keyword]")
			continue
		}

		targetID, _ := peer.Decode(input[0])
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
