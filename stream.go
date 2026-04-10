package easyp2p

import (
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// Stream is a wrapper around network.Stream with simplified methods.
type Stream struct {
	network.Stream
}

// HandleProtocol registers a handler for a specific protocol.
func (n *Node) HandleProtocol(pid string, handler func(s *Stream)) {
	n.Host.SetStreamHandler(protocol.ID(pid), func(s network.Stream) {
		handler(&Stream{s})
	})
}

// ConnectTo opens a stream to a peer for a specific protocol.
func (n *Node) ConnectTo(peerID peer.ID, pid string) (*Stream, error) {
	s, err := n.Host.NewStream(n.ctx, peerID, protocol.ID(pid))
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}
	return &Stream{s}, nil
}

// ReadFull reads data from the stream into the provided buffer.
func (s *Stream) ReadFull(buf []byte) error {
	_, err := io.ReadFull(s.Stream, buf)
	return err
}

// WriteAll writes data from the provided buffer to the stream.
func (s *Stream) WriteAll(buf []byte) error {
	_, err := s.Stream.Write(buf)
	return err
}

// Close closes the stream.
func (s *Stream) Close() error {
	return s.Stream.Close()
}
