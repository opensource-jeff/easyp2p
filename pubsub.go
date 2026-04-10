package easyp2p

import (
	"context"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// Topic represents a PubSub topic.
type Topic struct {
	topic *pubsub.Topic
	sub   *pubsub.Subscription
	node  *Node
	ctx   context.Context
}

// JoinTopic joins a PubSub topic.
func (n *Node) JoinTopic(topicName string) (*Topic, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if t, ok := n.topics[topicName]; ok {
		return t, nil
	}

	t, err := n.PubSub.Join(topicName)
	if err != nil {
		return nil, fmt.Errorf("failed to join topic: %w", err)
	}

	sub, err := t.Subscribe()
	if err != nil {
		t.Close()
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	topic := &Topic{
		topic: t,
		sub:   sub,
		node:  n,
		ctx:   n.ctx,
	}

	n.topics[topicName] = topic
	return topic, nil
}

// Publish broadcasts a message to the topic.
func (t *Topic) Publish(message string) error {
	return t.topic.Publish(t.ctx, []byte(message))
}

// OnMessage registers a callback for messages on this topic.
func (t *Topic) OnMessage(handler func(msg string, sender string)) {
	go func() {
		for {
			msg, err := t.sub.Next(t.ctx)
			if err != nil {
				return
			}
			// Don't handle our own messages
			if msg.ReceivedFrom == t.node.Host.ID() {
				continue
			}
			handler(string(msg.Data), msg.ReceivedFrom.String())
		}
	}()
}

// Close leaves the topic.
func (t *Topic) Close() error {
	t.node.mu.Lock()
	delete(t.node.topics, t.topic.String())
	t.node.mu.Unlock()

	t.sub.Cancel()
	return t.topic.Close()
}
