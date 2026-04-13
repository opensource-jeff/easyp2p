# Discovery: Finding Other Peers 🔍

`easyp2p` automatically helps you find other nodes on the network. It uses a hybrid discovery approach to work locally and globally.

## How it Works

### 1. Local Discovery (mDNS) 🏠
If you run two `easyp2p` nodes on the same Wi-Fi, they will find each other almost instantly.
- **Protocol**: mDNS (Multicast DNS).
- **Pros**: Zero configuration, very fast, works without internet.
- **Cons**: Only works within the same local network.

### 2. Global Discovery (DHT) 🌍
To find peers on the public internet, `easyp2p` uses the **Kademlia DHT**.
1. **Bootstrap**: First, your node connects to well-known **Bootstrap Nodes** provided by Protocol Labs.
2. **Join the Network**: Your node announces its existence to these bootstrap nodes and asks for other peers.
3. **Peer Routing**: Over time, your node builds a "routing table" and knows more and more peers directly.

## Using Peer Information

You can listen for new peers found by `easyp2p`:

```go
node.OnPeerFound(func(info peer.AddrInfo) {
    fmt.Printf("New peer found: %s\n", info.ID)
})

node.OnPeerLost(func(info peer.AddrInfo) {
    fmt.Printf("Peer disconnected: %s\n", info.ID)
})
```

## NAT Traversal & Reachability 🧱
Many peers are behind home routers (NAT). `easyp2p` uses several techniques to help nodes connect:
- **UPnP / NAT-PMP**: Automatically asks your router to open a port.
- **Circuit Relays**: If two nodes cannot connect directly, they will automatically use a third peer (a "Relay") to communicate.

You can monitor your node's reachability status:
```go
node.OnNATStatusChange(func(status easyp2p.NATStatus) {
    fmt.Printf("My network status: %s\n", status)
})
```
- `unknown`: Still checking.
- `open`: Directly reachable from the internet.
- `relayed`: Reachable only via a third party.
- `unreachable`: Cannot be reached.
