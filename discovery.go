package easyp2p

import (
	"context"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
)

// bootstrap attempts to connect to a list of bootstrap nodes.
func (n *Node) bootstrap(peers []string) {
	var wg sync.WaitGroup
	for _, paddr := range peers {
		ma, err := multiaddr.NewMultiaddr(paddr)
		if err != nil {
			fmt.Printf("Error: invalid bootstrap address %s: %v\n", paddr, err)
			continue
		}

		pi, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			fmt.Printf("Error: invalid bootstrap address %s: %v\n", paddr, err)
			continue
		}

		wg.Add(1)
		go func(pi *peer.AddrInfo) {
			defer wg.Done()
			if err := n.Host.Connect(n.ctx, *pi); err != nil {
				// Success or failure, bootstrapping is best effort
			}
		}(pi)
	}
	wg.Wait()

	// After connecting to bootstrap nodes, start DHT routing
	if err := n.DHT.Bootstrap(n.ctx); err != nil {
		fmt.Printf("Warning: DHT bootstrap failed: %v\n", err)
	}
}

// setupMDNS initializes mDNS for local network discovery.
func (n *Node) setupMDNS() error {
	ser := mdns.NewMdnsService(n.Host, "easyp2p-discovery", &mdnsHandler{node: n})
	return ser.Start()
}

type mdnsHandler struct {
	node *Node
}

// HandlePeerFound is called by mDNS when a peer is discovered locally.
func (h *mdnsHandler) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == h.node.Host.ID() {
		return
	}
	if err := h.node.Host.Connect(context.Background(), pi); err != nil {
		// Log connection failure? 
	}
}
