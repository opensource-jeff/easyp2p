package easyp2p

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// Node represents a simplified libp2p node.
type Node struct {
	Host    host.Host
	DHT     *dht.IpfsDHT
	PubSub  *pubsub.PubSub
	ctx     context.Context
	cancel  context.CancelFunc
	
	mu      sync.Mutex
	topics  map[string]*Topic
}

// Config holds the configuration for a Node.
type Config struct {
	ListenPort     int
	BootstrapPeers []string
	EnableMDNS     bool
	KeyPath        string // Path to save/load the node's private key
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	var bootstrapPeers []string
	for _, p := range dht.DefaultBootstrapPeers {
		bootstrapPeers = append(bootstrapPeers, p.String())
	}

	return Config{
		ListenPort:     0, // random port
		BootstrapPeers: bootstrapPeers,
		EnableMDNS:     true,
		KeyPath:        "", // disabled by default
	}
}

// NewNode creates and starts a new libp2p node.
func NewNode(ctx context.Context, cfg Config) (*Node, error) {
	ctx, cancel := context.WithCancel(ctx)

	// Define listen addresses
	var listenAddrs []string
	if cfg.ListenPort == 0 {
		listenAddrs = []string{
			"/ip4/0.0.0.0/tcp/0",
			"/ip4/0.0.0.0/udp/0/quic-v1",
		}
	} else {
		listenAddrs = []string{
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cfg.ListenPort),
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", cfg.ListenPort),
		}
	}

	// Prepare options
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(listenAddrs...),
		libp2p.NATPortMap(),
	}

	// Load or generate identity if KeyPath is provided
	if cfg.KeyPath != "" {
		priv, err := loadOrGenerateKey(cfg.KeyPath)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to load identity: %w", err)
		}
		opts = append(opts, libp2p.Identity(priv))
	}

	// Initialize libp2p host
	h, err := libp2p.New(opts...)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	// Initialize PubSub (GossipSub)
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		h.Close()
		cancel()
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}

	// Initialize Kademlia DHT in client mode (can be changed to server mode if needed)
	kdht, err := dht.New(ctx, h, dht.Mode(dht.ModeAuto))
	if err != nil {
		h.Close()
		cancel()
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	node := &Node{
		Host:   h,
		DHT:    kdht,
		PubSub: ps,
		ctx:    ctx,
		cancel: cancel,
		topics: make(map[string]*Topic),
	}

	// Start bootstrapping
	if len(cfg.BootstrapPeers) > 0 {
		go node.bootstrap(cfg.BootstrapPeers)
	}

	// Start mDNS if enabled
	if cfg.EnableMDNS {
		if err := node.setupMDNS(); err != nil {
			// Log error but continue
			fmt.Printf("Warning: failed to setup mDNS: %v\n", err)
		}
	}

	return node, nil
}

// ID returns the peer ID of the node.
func (n *Node) ID() peer.ID {
	return n.Host.ID()
}

// Addrs returns the multiaddresses the node is listening on.
func (n *Node) Addrs() []multiaddr.Multiaddr {
	return n.Host.Addrs()
}

// Close shuts down the node.
func (n *Node) Close() error {
	n.cancel()
	return n.Host.Close()
}

// loadOrGenerateKey reads a private key from disk or generates a new one.
func loadOrGenerateKey(path string) (crypto.PrivKey, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Generate a new Ed25519 key pair
		priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
		if err != nil {
			return nil, err
		}

		// Marshal the private key to bytes
		keyBytes, err := crypto.MarshalPrivateKey(priv)
		if err != nil {
			return nil, err
		}

		// Save to disk with restrictive permissions (0600)
		if err := os.WriteFile(path, keyBytes, 0600); err != nil {
			return nil, err
		}

		return priv, nil
	}

	// Load existing key from disk
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal the bytes back into a PrivKey object
	return crypto.UnmarshalPrivateKey(keyBytes)
}
