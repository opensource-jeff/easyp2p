package identity

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Identity holds the node's private key and peer ID.
type Identity struct {
	PrivKey crypto.PrivKey
	PeerID  peer.ID
}

// Generate creates a new Ed25519 key pair without saving it to disk.
func Generate() (*Identity, error) {
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	id, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("failed to derive peer ID: %w", err)
	}

	return &Identity{
		PrivKey: priv,
		PeerID:  id,
	}, nil
}

// LoadOrCreate loads an Ed25519 private key from the given path or creates a new one.
// If the file does not exist, it generates a new key pair and saves it to the path with 0600 permissions.
func LoadOrCreate(path string) (*Identity, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Ensure the directory exists
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory for identity: %w", err)
		}

		// Generate a new Ed25519 key pair
		priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
		if err != nil {
			return nil, fmt.Errorf("failed to generate key pair: %w", err)
		}

		// Marshal the private key to bytes
		keyBytes, err := crypto.MarshalPrivateKey(priv)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal private key: %w", err)
		}

		// Save to disk with restrictive permissions (0600)
		if err := os.WriteFile(path, keyBytes, 0600); err != nil {
			return nil, fmt.Errorf("failed to save identity to disk: %w", err)
		}

		// Derive Peer ID
		id, err := peer.IDFromPrivateKey(priv)
		if err != nil {
			return nil, fmt.Errorf("failed to derive peer ID: %w", err)
		}

		return &Identity{
			PrivKey: priv,
			PeerID:  id,
		}, nil
	}

	// Load existing key from disk
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read identity file: %w", err)
	}

	// Unmarshal the bytes back into a PrivKey object
	priv, err := crypto.UnmarshalPrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal private key: %w", err)
	}

	// Derive Peer ID
	id, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("failed to derive peer ID: %w", err)
	}

	return &Identity{
		PrivKey: priv,
		PeerID:  id,
	}, nil
}
