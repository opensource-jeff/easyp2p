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

	// 1. Initialize the node
	node := easyp2p.Must(easyp2p.NewNode(ctx, easyp2p.DefaultConfig()))
	defer node.Close()

	node.PrintDescribe()

	// 2. Define protocol for VPN traffic
	const vpnProtocol = "/easyp2p/vpn/1.0"

	// 3. Set handler for incoming traffic
	node.HandleProtocol(vpnProtocol, func(s *easyp2p.Stream) {
		defer s.Close()
		fmt.Printf("\n--- Incoming traffic from %s ---\n", s.Conn().RemotePeer().String()[:8])
		
		buf := make([]byte, 1024)
		for {
			n, err := s.Read(buf)
			if err != nil {
				fmt.Printf("Connection closed by peer\n")
				break
			}
			fmt.Printf("Received %d bytes: %s\n", n, string(buf[:n]))
		}
	})

	// 4. Dial a peer for VPN connection (if provided as argument)
	if len(os.Args) > 1 {
		targetID, err := peer.Decode(os.Args[1])
		if err != nil {
			panic(fmt.Sprintf("Invalid peer ID: %v", err))
		}

		fmt.Printf("Connecting to %s for VPN...\n", targetID.String()[:8])
		
		// Use a timeout for the connection
		tctx, tcancel := context.WithTimeout(ctx, 30*time.Second)
		defer tcancel()

		stream, err := node.ConnectTo(targetID, vpnProtocol)
		if err != nil {
			fmt.Printf("Failed to connect: %v\n", err)
		} else {
			defer stream.Close()
			fmt.Println("VPN Connected! Sending heartbeat data...")
			stream.Write([]byte("HEARTBEAT - VPN is active!"))
		}
		_ = tctx // just to use it if I want to pass it to ConnectTo, but ConnectTo uses node.ctx
	}

	fmt.Println("\nWaiting for connections...")
	fmt.Println("Run this script with a Peer ID as an argument to connect to someone.")
	
	// Keep node running
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "/quit" {
			break
		}
	}
}
