# easyp2p 🚀

`easyp2p` is a beginner-friendly Go library that takes the complexity out of building peer-to-peer applications using `libp2p`. Whether you're building a chat server, a VPN, or a decentralized database, `easyp2p` provides simple abstractions to get you started in minutes.

## Features 🌟
- **Zero Configuration**: Sensible defaults for TCP, QUIC, and NAT traversal.
- **Auto Discovery**: Locally via mDNS and globally via Kademlia DHT.
- **Bootstrap Support**: Seamlessly join the network using well-known bootstrap nodes.
- **PubSub (GossipSub)**: Easy message broadcasting for chat and IRC.
- **Direct Streams**: Raw data transfer for custom protocols like VPNs and DB queries.

## Quick Start 🛠️

### 1. Installation
```bash
go get github.com/opensource-jeff/easyp2p
```

### 2. Basic Node Setup
```go
package main

import (
    "context"
    "time"
    "fmt"
    "easyp2p"
)

func main() {
    ctx := context.Background()
    
    // Create a node with default settings (automatically handles identity and peer cache)
    node := easyp2p.Must(easyp2p.NewNode(ctx, easyp2p.DefaultConfig()))
    defer node.Close()

    // Listen for NAT status changes
    node.OnNATStatusChange(func(status easyp2p.NATStatus) {
        fmt.Printf("NAT Status changed: %s\n", status)
    })

    // Wait for the network to be ready (at least 1 non-bootstrap peer)
    fmt.Println("Connecting to peers...")
    if err := node.WaitForNetwork(ctx, 30*time.Second); err != nil {
        fmt.Println("Error:", err)
    }

    // Display node information
    node.PrintDescribe()
}
```

## Use Cases 📖

### Decentralized IRC (PubSub)
Join a room and broadcast messages to all connected peers.

```go
topic, _ := node.JoinTopic("global-chat")

// Receive messages
topic.OnMessage(func(msg string, sender string) {
    fmt.Printf("[%s]: %s\n", sender, msg)
})

// Send a message
topic.Publish("Hello everyone!")
```

### Peer-to-Peer VPN (Streams)
Establish direct point-to-point connections for raw data transfer.

```go
const myProtocol = "/my-app/vpn/1.0"

// Handle incoming data
node.HandleProtocol(myProtocol, func(s *easyp2p.Stream) {
    buf := make([]byte, 1024)
    n, _ := s.Read(buf)
    fmt.Println("Received data:", string(buf[:n]))
})

// Connect to a peer
stream, _ := node.ConnectTo(peerID, myProtocol)
stream.Write([]byte("Confidential VPN traffic"))
```

### Decentralized Database (Request-Response)
Implement custom search logic over the network.

```go
const dbProtocol = "/my-app/db/1.0"

// Search responder
node.HandleProtocol(dbProtocol, func(s *easyp2p.Stream) {
    query, _ := s.Read(...)
    result := searchLocalDB(query)
    s.Write([]byte(result))
})

// Query another node
stream, _ := node.ConnectTo(peerID, dbProtocol)
stream.Write([]byte("search-query"))
```

## Security & Networking 🔐

### Encryption & Authentication
`easyp2p` uses the industry-standard **libp2p** security stack:
- **End-to-End Encryption**: All traffic is encrypted using **Noise** or **TLS 1.3** by default.
- **Identity (Ed25519)**: Every node is uniquely identified by an Ed25519 private key. This key is used during the cryptographic handshake to verify the `PeerID` of the other node, preventing impersonation.
- **Identity Persistence**: By default, your node's private key is saved to `~/.config/easyp2p/identity.key`, so your `PeerID` stays the same every time you restart your application. You can disable this by setting `Persist: false` in your `Config`.

### Peer Discovery & Connectivity
`easyp2p` uses a multi-layered approach to help you find other peers:
- **Local Discovery (mDNS)**: Peers on the same Wi-Fi or local network will find each other automatically without any configuration.
- **Global Discovery (Bootstrap Servers)**: On the public internet, your node first connects to a set of stable **Bootstrap Servers** (provided by Protocol Labs). These servers act as "meeting points" to help your node join the global P2P network.
- **Kademlia DHT**: Once connected to a bootstrap node, your node uses the **Distributed Hash Table (DHT)** to find other peers running your application anywhere in the world.
- **NAT Traversal**: `easyp2p` automatically attempts to punch through firewalls and routers using **UPnP** and **NAT-PMP**. If a direct connection is impossible, it will automatically use **Circuit Relays** to ensure you can still communicate.

## Credits & Disclosure 🤖

This library was **vibe-coded** with the help of Gemini CLI. It was built to make `libp2p` accessible for beginners and personal projects. While it is fully functional and tested, it is intended as a learning tool and a foundation for personal decentralized experiments.

## Examples 📁
Check out the `examples/` directory for full, runnable implementations:
- `examples/chat`: A simple IRC-style command line chat.
- `examples/vpn`: A basic stream-based data transfer example.
- `examples/database`: A request-response pattern for network searching.

## Documentation & Wiki 📖
For more in-depth guides, check out our [Wiki](docs/WIKI.md):
- [Getting Started & Configuration](docs/Configuration.md)
- [Discovery & Networking](docs/Discovery.md)
- [Protocols & Streams](docs/Protocols-and-Streams.md)
- [Core Concepts](CONCEPTS.md)
