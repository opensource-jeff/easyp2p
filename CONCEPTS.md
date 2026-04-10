# Libp2p Concepts Guide 🧠

This guide explains the "magic" behind `libp2p` and `easyp2p` in simple terms.

---

## 1. Building Protocols 🗺️
Think of a **Protocol ID** like a path in a URL (e.g., `/google.com/mail` vs `/google.com/drive`). 

Even though you have one connection between two peers, you can run many different "apps" over that same connection by using different Protocol IDs.

### How to design them:
- **Be Specific**: Always version your protocols (e.g., `/my-irc/1.0.0`).
- **Handlers**: Use `node.HandleProtocol` to tell your node what to do when someone talks to it using that specific ID.
- **Example**: 
    - IRC: `/easyp2p/chat/1.0`
    - VPN: `/easyp2p/vpn/1.0`
    - Database: `/easyp2p/db/1.0`

---

## 2. Concurrent Connections 🏎️
P2P apps are naturally "busy." You might be talking to 50 peers at once. If you're not careful, one slow peer can freeze your whole app.

### My Top Advice for Concurrency:
1.  **Goroutines are your friends**: In `easyp2p`, I've already set up `OnMessage` to run in a goroutine. This means your app keeps running while it waits for a message.
2.  **Use Contexts for Timeouts**: Never wait forever for a peer. 
    ```go
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    stream, err := node.ConnectTo(ctx, peerID, protocol)
    ```
3.  **The "Peer Manager" Pattern**: Create a simple map to keep track of who you are connected to. Use a `sync.RWMutex` to make it safe for multiple goroutines to read and write to that map.

---

## 3. DHT Discovery (The "Global Phonebook") 🌍
How do you find a peer in Tokyo from your house in London? You use the **Distributed Hash Table (DHT)**.

### The Analogy: "Friend of a Friend"
Imagine you are looking for a rare book. 
1. You ask your 5 closest friends (Bootstrap Nodes).
2. They don't have it, but they know 5 other people who might.
3. You ask those people, and eventually, someone says, "Oh, I know exactly who has that book!"

### Key DHT Concepts:
- **Bootstrap Nodes**: These are the "well-known" people everyone knows. They help you join the network.
- **Routing Table**: Your node keeps a list of peers it knows about. It's like a mini-address book that gets smarter the longer your node is online.
- **FindPeer**: Asking the network, "Where is Peer X?"
- **Provide/FindProviders**: Telling the network, "I have the Database Service," or asking "Who has the Database Service?"

---

## 4. Discovery: mDNS vs DHT 🔍
- **mDNS (Local)**: Like shouting in a room. Everyone in the room (your Wi-Fi) hears you. It's **instant** but only works locally.
- **DHT (Global)**: Like putting an ad in a global newspaper. It takes a **few seconds** to propagate, but it works anywhere in the world.

`easyp2p` uses **both** automatically so your app just works!
