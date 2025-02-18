package praetorian

import (
	"crypto/aes"
	"crypto/cipher"
	"sync"
)

type keystore struct {
	sync.Map
}

// NewKeyset initializes a new Keyring from the provided config and returns it.
func NewKeystore(cfg *config) (KeyFinder, error) {
	ks := &keystore{}
	for id, val := range cfg.RootKeys {
		if id == cfg.ActiveKeyID {
			ks.Store("active", &key{id, val})
		}
		ks.Store(id, &key{id, val})
	}
	return ks, nil
}

// Find a root key with the given identifier.
func (ks *keystore) Find(id string) (RootKey, error) {
	if k, ok := ks.Load(id); ok {
		return k.(*key), nil
	}
	return nil, ErrRootKeyNotFound
}

type key struct {
	id    string
	value []byte
}

// ID is a getter which returns the keys unique identifier.
func (k *key) ID() string {
	return k.id
}

// Encrypt the given data using the current root key.
func (k *key) Encrypt(d []byte) ([]byte, error) {
	block, err := aes.NewCipher(k.value)
	if err != nil {
		return nil, ErrNewCipherBlock
	}
	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, ErrNewGCMWithRandomNonce
	}
	return gcm.Seal(nil, nil, d, nil), nil
}

// Decrypt the given data using the current root key.
func (k *key) Decrypt(d []byte) ([]byte, error) {
	block, err := aes.NewCipher(k.value)
	if err != nil {
		return nil, ErrNewCipherBlock
	}
	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, ErrNewGCMWithRandomNonce
	}
	ci, err := gcm.Open(nil, nil, d, nil)
	if err != nil {
		return nil, ErrGCMOpen
	}
	return ci, nil
}
