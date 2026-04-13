package easyp2p

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/opensource-jeff/easyp2p/internal/identity"
	"github.com/opensource-jeff/easyp2p/internal/peercache"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// NATStatus represents the status of NAT reachability.
type NATStatus string

const (
	NATStatusUnknown     NATStatus = "checking"
	NATStatusOpen        NATStatus = "open"
	NATStatusRelayed     NATStatus = "relayed"
	NATStatusUnreachable NATStatus = "unreachable"
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

	peerCache      *peercache.PeerCache
	bootstrapPeers []peer.ID

	natMu         sync.RWMutex
	natCallbacks  []func(NATStatus)
	lastNATStatus NATStatus

	peerMu             sync.RWMutex
	peerFoundCallbacks []func(peer.AddrInfo)
	peerLostCallbacks  []func(peer.AddrInfo)
}

// Config holds the configuration for a Node.
type Config struct {
	ListenPort     int
	BootstrapPeers []string
	EnableMDNS     bool
	IdentityPath   string // Path to save/load the node's private key
	PeerCachePath  string // Path to save/load known peers
	Persist        bool   // Whether to persist identity and peer cache
}

// DefaultConfig returns a sensible default configuration.
func DefaultConfig() Config {
	var bootstrapPeers []string
	for _, p := range dht.DefaultBootstrapPeers {
		bootstrapPeers = append(bootstrapPeers, p.String())
	}

	configDir, _ := os.UserConfigDir()

	return Config{
		ListenPort:     0, // random port
		BootstrapPeers: bootstrapPeers,
		EnableMDNS:     true,
		IdentityPath:   filepath.Join(configDir, "easyp2p", "identity.key"),
		PeerCachePath:  filepath.Join(configDir, "easyp2p", "peers.json"),
		Persist:        true,
	}
}

// Must is a helper that panics if err is not nil, otherwise it returns the node.
func Must(node *Node, err error) *Node {
	if err != nil {
		panic(fmt.Sprintf("easyp2p: failed to create node: %v", err))
	}
	return node
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
		libp2p.EnableNATService(),
	}

	// Load or generate identity
	var id *identity.Identity
	var err error
	if cfg.Persist && cfg.IdentityPath != "" {
		id, err = identity.LoadOrCreate(cfg.IdentityPath)
	} else {
		id, err = identity.Generate()
	}

	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to handle identity: %w", err)
	}
	opts = append(opts, libp2p.Identity(id.PrivKey))

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

	// Initialize Kademlia DHT in server mode so we can be discovered by others
	kdht, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		h.Close()
		cancel()
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	var bootstrapIDs []peer.ID
	for _, p := range cfg.BootstrapPeers {
		addr, err := multiaddr.NewMultiaddr(p)
		if err != nil {
			continue
		}
		info, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			continue
		}
		bootstrapIDs = append(bootstrapIDs, info.ID)
	}

	pc := peercache.New(cfg.PeerCachePath)

	node := &Node{
		Host:           h,
		DHT:            kdht,
		PubSub:         ps,
		ctx:            ctx,
		cancel:         cancel,
		topics:         make(map[string]*Topic),
		peerCache:      pc,
		bootstrapPeers: bootstrapIDs,
		lastNATStatus:  NATStatusUnknown,
	}

	// Load and connect to cached peers
	if cfg.Persist {
		cachedPeers, err := pc.Load()
		if err == nil && len(cachedPeers) > 0 {
			for _, p := range cachedPeers {
				go func(p peer.AddrInfo) {
					tctx, tcancel := context.WithTimeout(ctx, 5*time.Second)
					defer tcancel()
					if err := h.Connect(tctx, p); err != nil {
						// Silent failure for cached peers
					}
				}(p)
			}
		}
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

	// Start background peer cache saving
	if cfg.Persist {
		go node.backgroundPeerCache()
	}

	// Start NAT status polling
	go node.backgroundNATStatus()

	// Register network notifier
	h.Network().Notify(&notifier{node: node})

	return node, nil
}

// OnPeerFound registers a callback to be fired when a new peer connects.
func (n *Node) OnPeerFound(fn func(peer.AddrInfo)) {
	n.peerMu.Lock()
	defer n.peerMu.Unlock()
	n.peerFoundCallbacks = append(n.peerFoundCallbacks, fn)
}

// OnPeerLost registers a callback to be fired when a peer disconnects.
func (n *Node) OnPeerLost(fn func(peer.AddrInfo)) {
	n.peerMu.Lock()
	defer n.peerMu.Unlock()
	n.peerLostCallbacks = append(n.peerLostCallbacks, fn)
}

type notifier struct {
	node *Node
}

func (nt *notifier) Listen(network.Network, multiaddr.Multiaddr)      {}
func (nt *notifier) ListenClose(network.Network, multiaddr.Multiaddr) {}

func (nt *notifier) Connected(net network.Network, conn network.Conn) {
	peerID := conn.RemotePeer()
	info := nt.node.Host.Peerstore().PeerInfo(peerID)

	nt.node.peerMu.RLock()
	callbacks := make([]func(peer.AddrInfo), len(nt.node.peerFoundCallbacks))
	copy(callbacks, nt.node.peerFoundCallbacks)
	nt.node.peerMu.RUnlock()

	for _, fn := range callbacks {
		go fn(info)
	}
}

func (nt *notifier) Disconnected(net network.Network, conn network.Conn) {
	peerID := conn.RemotePeer()
	// Check if this was the last connection to this peer
	if net.Connectedness(peerID) == network.Connected {
		return
	}

	info := nt.node.Host.Peerstore().PeerInfo(peerID)

	nt.node.peerMu.RLock()
	callbacks := make([]func(peer.AddrInfo), len(nt.node.peerLostCallbacks))
	copy(callbacks, nt.node.peerLostCallbacks)
	nt.node.peerMu.RUnlock()

	for _, fn := range callbacks {
		go fn(info)
	}
}

// OnNATStatusChange registers a callback to be fired when the NAT status changes.
func (n *Node) OnNATStatusChange(fn func(NATStatus)) {
	n.natMu.Lock()
	defer n.natMu.Unlock()
	n.natCallbacks = append(n.natCallbacks, fn)
}

func (n *Node) backgroundNATStatus() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Listen for reachability events to update current status
	sub, err := n.Host.EventBus().Subscribe(new(event.EvtLocalReachabilityChanged))
	if err != nil {
		fmt.Printf("Warning: failed to subscribe to reachability events: %v\n", err)
	}
	defer func() {
		if sub != nil {
			sub.Close()
		}
	}()

	currentReach := network.ReachabilityUnknown

	for {
		select {
		case <-n.ctx.Done():
			return
		case e := <-sub.Out():
			evt := e.(event.EvtLocalReachabilityChanged)
			currentReach = evt.Reachability
		case <-ticker.C:
			// Check reachability
			status := n.mapReachability(currentReach)
			
			// Check if we have any p2p-circuit addresses to see if it's relayed
			if status == NATStatusUnreachable {
				for _, addr := range n.Host.Addrs() {
					for _, part := range addr.Protocols() {
						if part.Code == multiaddr.P_CIRCUIT {
							status = NATStatusRelayed
							break
						}
					}
					if status == NATStatusRelayed {
						break
					}
				}
			}

			n.natMu.Lock()
			if status != n.lastNATStatus {
				n.lastNATStatus = status
				callbacks := make([]func(NATStatus), len(n.natCallbacks))
				copy(callbacks, n.natCallbacks)
				n.natMu.Unlock()

				for _, fn := range callbacks {
					go fn(status)
				}
			} else {
				n.natMu.Unlock()
			}
		}
	}
}

func (n *Node) mapReachability(reach network.Reachability) NATStatus {
	switch reach {
	case network.ReachabilityUnknown:
		return NATStatusUnknown
	case network.ReachabilityPublic:
		return NATStatusOpen
	case network.ReachabilityPrivate:
		return NATStatusUnreachable
	default:
		return NATStatusUnknown
	}
}

func (n *Node) backgroundPeerCache() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			// Final save on shutdown
			n.savePeers()
			return
		case <-ticker.C:
			n.savePeers()
		}
	}
}

func (n *Node) savePeers() {
	peerstore := n.Host.Peerstore()
	var peers []peer.AddrInfo
	for _, p := range peerstore.Peers() {
		// Only save peers we have addresses for and that are not our own
		if p == n.Host.ID() {
			continue
		}
		info := peerstore.PeerInfo(p)
		if len(info.Addrs) > 0 {
			peers = append(peers, info)
		}
	}
	_ = n.peerCache.Save(peers)
}

// ID returns the peer ID of the node.
func (n *Node) ID() peer.ID {
	return n.Host.ID()
}

// Addrs returns the multiaddresses the node is listening on.
func (n *Node) Addrs() []multiaddr.Multiaddr {
	return n.Host.Addrs()
}

// WaitForNetwork blocks until the node has at least 1 connected peer that is NOT a bootstrap node.
// It polls every 500ms and returns an error if the timeout is reached or context is cancelled.
func (n *Node) WaitForNetwork(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		// Check for timeout
		if time.Now().After(deadline) {
			return fmt.Errorf("network not ready: could not find peers within %s. Check your internet connection or bootstrap node addresses.", timeout)
		}

		// Check if at least 1 connected peer is NOT a bootstrap node
		connectedPeers := n.Host.Network().Peers()
		for _, p := range connectedPeers {
			isBootstrap := false
			for _, b := range n.bootstrapPeers {
				if p == b {
					isBootstrap = true
					break
				}
			}

			if !isBootstrap {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-n.ctx.Done():
			return n.ctx.Err()
		case <-ticker.C:
			// continue polling
		}
	}
}

// Close shuts down the node.
func (n *Node) Close() error {
	n.cancel()
	return n.Host.Close()
}

// Describe returns a formatted multi-line string containing node information.
func (n *Node) Describe() string {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.natMu.RLock()
	nat := n.lastNATStatus
	n.natMu.RUnlock()

	out := fmt.Sprintf("Peer ID: %s\n", n.Host.ID())
	out += "Listen Addresses:\n"
	for _, addr := range n.Host.Addrs() {
		out += fmt.Sprintf("  - %s\n", addr)
	}

	connectedPeers := n.Host.Network().Peers()
	out += fmt.Sprintf("Connected Peers: %d\n", len(connectedPeers))
	out += fmt.Sprintf("NAT Status: %s\n", nat)

	if len(n.topics) > 0 {
		out += "Topics:\n"
		for name, topic := range n.topics {
			peers := topic.topic.ListPeers()
			out += fmt.Sprintf("  - %s (%d subscribers)\n", name, len(peers))
		}
	} else {
		out += "Topics: None\n"
	}

	return out
}

// PrintDescribe prints the node information to stdout.
func (n *Node) PrintDescribe() {
	fmt.Println(n.Describe())
}
