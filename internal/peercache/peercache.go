package peercache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p/core/peer"
)

// PeerCache handles saving and loading of known peers.
type PeerCache struct {
	Path string
}

// New creates a new PeerCache with the given path.
func New(path string) *PeerCache {
	return &PeerCache{Path: path}
}

// Save marshals peers to JSON and writes to disk at the cache path with permissions 0600.
func (pc *PeerCache) Save(peers []peer.AddrInfo) error {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(pc.Path), 0755); err != nil {
		return fmt.Errorf("failed to create directory for peer cache: %w", err)
	}

	data, err := json.Marshal(peers)
	if err != nil {
		return fmt.Errorf("failed to marshal peers: %w", err)
	}

	if err := os.WriteFile(pc.Path, data, 0600); err != nil {
		return fmt.Errorf("failed to save peer cache to disk: %w", err)
	}

	return nil
}

// Load reads and unmarshals from disk. If the file does not exist, return nil, nil.
func (pc *PeerCache) Load() ([]peer.AddrInfo, error) {
	if _, err := os.Stat(pc.Path); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(pc.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read peer cache file: %w", err)
	}

	var peers []peer.AddrInfo
	if err := json.Unmarshal(data, &peers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal peer cache: %w", err)
	}

	return peers, nil
}
