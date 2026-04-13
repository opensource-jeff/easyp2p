# easyp2p Wiki 📖

Welcome to the `easyp2p` wiki! This guide provides in-depth documentation on how the library works and how to build powerful P2P applications with it.

## Table of Contents 📑

### 1. [Getting Started](Configuration.md)
Learn how to configure your node, handle ports, and manage identity persistence.

### 2. [Discovery & Networking](Discovery.md)
Understand how `easyp2p` finds other nodes locally (mDNS) and globally (DHT), and how NAT traversal works.

### 3. [Communication: Protocols & Streams](Protocols-and-Streams.md)
Learn how to define your own protocols, establish direct streams between peers, and use PubSub for broadcasting.

### 4. [Core Concepts](../CONCEPTS.md)
A high-level overview of the "magic" behind P2P networks.

---

## Architecture at a Glance 🏗️

`easyp2p` is built on top of [libp2p](https://libp2p.io/), the modular network stack used by IPFS and Ethereum 2.0. It simplifies the most common tasks:

- **Transport**: Automatically uses TCP and QUIC.
- **Security**: All traffic is encrypted using Noise or TLS 1.3.
- **Multiplexing**: Many streams can run over a single connection.
- **Routing**: Uses Kademlia DHT for finding peers and services.

## Use Cases 🚀
- **Chat Applications**: Using PubSub for rooms and Streams for private messaging.
- **Decentralized Databases**: Using request-response streams for querying remote nodes.
- **P2P VPNs**: Tunneling traffic through encrypted streams.
- **File Sharing**: Establishing direct streams for high-speed data transfer.
