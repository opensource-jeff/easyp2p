# Configuration 🛠️

`easyp2p` is designed to be "Easy" by providing sensible defaults. You can use `DefaultConfig()` and then modify only the fields you need.

## The Config Struct

```go
type Config struct {
	ListenPort     int      // The port to listen on. 0 means random (default).
	BootstrapPeers []string // List of multiaddresses for bootstrap nodes.
	EnableMDNS     bool     // Whether to enable local mDNS discovery.
	IdentityPath   string   // Path to save/load the node's private key.
	PeerCachePath  string   // Path to save/load known peers.
	Persist        bool     // Whether to persist identity and peer cache.
}
```

### 1. `ListenPort`
By default, this is `0`, which means `libp2p` will pick a random available port. This is usually the best choice for client applications or nodes behind NAT. If you are running a server with a static IP, you might want to set this to a fixed port (e.g., `4001`).

### 2. `BootstrapPeers`
`easyp2p` comes with a list of default bootstrap peers from Protocol Labs. These help you join the global P2P network. You can provide your own list if you are running a private network.

### 3. `EnableMDNS`
Enabled by default (`true`). This allows nodes on the same local network (Wi-Fi/LAN) to find each other instantly without needing a DHT or internet connection.

### 4. `IdentityPath` & `PeerCachePath`
These define where your node's private key and discovered peers are saved. 
- Default Identity Path: `~/.config/easyp2p/identity.key`
- Default Peer Cache Path: `~/.config/easyp2p/peers.json`

### 5. `Persist` 🆕
Added in v1.1.0. This boolean controls whether your node's identity and peer cache are saved to disk.
- **`Persist: true` (Default)**: Your `PeerID` remains the same across restarts. Discovered peers are remembered.
- **`Persist: false`**: A new `PeerID` is generated every time. No data is written to disk. This is useful for demos, tests, or ephemeral client nodes.

## Usage Example

```go
cfg := easyp2p.DefaultConfig()
cfg.Persist = false
cfg.ListenPort = 5001

node, err := easyp2p.NewNode(ctx, cfg)
```
