# Protocols & Streams 🗺️

In `easyp2p`, you communicate by defining **Protocols**. A protocol is a specific "app" or "service" running on your node.

## Why Protocols?
Using protocols allows your node to do many things at once:
- Handle chat messages (`/chat/1.0`)
- Respond to database queries (`/db/1.0`)
- Facilitate a VPN tunnel (`/vpn/1.0`)

All of these can run over the same physical connection between two peers.

## Setting Up a Protocol Handler
Tell your node what to do when someone connects using a specific protocol:

```go
const myProtocol = "/myapp/v1"

node.HandleProtocol(myProtocol, func(s *easyp2p.Stream) {
    defer s.Close()
    
    // Read from the stream
    buf := make([]byte, 1024)
    n, err := s.Read(buf)
    if err != nil {
        return
    }
    
    fmt.Printf("Received: %s\n", string(buf[:n]))
    
    // Respond back
    s.Write([]byte("Message received!"))
})
```

## Connecting to a Peer
To talk to another node, you must know their **Peer ID**:

```go
peerID, _ := peer.Decode("Qm...")
stream, err := node.ConnectTo(peerID, myProtocol)
if err != nil {
    // Handle error
}

// Write to the peer
stream.Write([]byte("Hello!"))

// Read response
result := make([]byte, 256)
n, _ := stream.Read(result)
```

## PubSub (GossipSub) 🗣️
If you want to broadcast messages to *everyone* in a group (like a chat room), use **Topics**.

```go
topic, _ := node.JoinTopic("general-chat")

// Receive broadcasts
topic.OnMessage(func(msg string, sender string) {
    fmt.Printf("[%s]: %s\n", sender, msg)
})

// Send broadcast
topic.Publish("Hello everyone!")
```
`easyp2p` uses **GossipSub**, which is a highly efficient peer-to-peer messaging protocol.
