package easyp2p

import (
	"context"
	"testing"
	"time"
)

func TestNodeDiscoveryAndPubSub(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. Create two nodes
	cfg1 := DefaultConfig()
	cfg1.EnableMDNS = true
	cfg1.BootstrapPeers = nil // No internet needed for local test
	node1, err := NewNode(ctx, cfg1)
	if err != nil {
		t.Fatal(err)
	}
	defer node1.Close()

	cfg2 := DefaultConfig()
	cfg2.EnableMDNS = true
	cfg2.BootstrapPeers = nil
	node2, err := NewNode(ctx, cfg2)
	if err != nil {
		t.Fatal(err)
	}
	defer node2.Close()

	// 2. Join same topic
	topic1, err := node1.JoinTopic("test-topic")
	if err != nil {
		t.Fatal(err)
	}
	topic2, err := node2.JoinTopic("test-topic")
	if err != nil {
		t.Fatal(err)
	}

	// 3. Setup message receiver on node 2
	received := make(chan string, 1)
	topic2.OnMessage(func(msg string, sender string) {
		received <- msg
	})

	// 4. Wait for mDNS discovery (can take a few seconds)
	// In a real test we'd check peerstore, but for simplicity we'll just try publishing
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	testMsg := "hello from node 1"
	for {
		select {
		case <-ctx.Done():
			t.Fatal("timed out waiting for message")
		case <-ticker.C:
			topic1.Publish(testMsg)
		case msg := <-received:
			if msg != testMsg {
				t.Fatalf("expected %s, got %s", testMsg, msg)
			}
			return // Success!
		}
	}
}
