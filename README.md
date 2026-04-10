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
    "easyp2p"
)

func main() {
    ctx := context.Background()
    
    // Create a node with default settings
    node, err := easyp2p.NewNode(ctx, easyp2p.DefaultConfig())
    if err != nil {
        panic(err)
    }
    defer node.Close()

    println("My Peer ID:", node.ID().String())
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

## Peer Discovery 🌍
- **Local Network**: Peers on the same Wi-Fi will find each other automatically via mDNS.
- **Internet**: The library uses public "Bootstrap Servers" to help your node join the global P2P network. It then uses the Kademlia DHT to find other peers across the globe.

## Credits & Disclosure 🤖

This library was **vibe-coded** with the help of Gemini CLI. It was built to make `libp2p` accessible for beginners and personal projects. While it is fully functional and tested, it is intended as a learning tool and a foundation for personal decentralized experiments.

## Examples 📁
Check out the `examples/` directory for full, runnable implementations:
- `examples/chat`: A simple IRC-style command line chat.
- `examples/vpn`: A basic stream-based data transfer example.
- `examples/database`: A request-response pattern for network searching.
