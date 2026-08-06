// Copyright © 2025 Karl Bateman. All Rights Reserved. Use of this software is
// governed by a BSD-style license that can be found in the LICENSE file.

// Keystore is the in-memory KeyFinder used at runtime, and key is its
// RootKey implementation, performing AES-256-GCM encryption and decryption.

package praetorian

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"sync"
)

// Keystore is an in-memory KeyFinder backed by root keys loaded from a
// Config.
type Keystore struct {
	keys sync.Map
}

// NewKeystore initializes a new Keystore from the provided config and returns it.
func NewKeystore(cfg *Config) (*Keystore, error) {
	ks := &Keystore{}
	for id, val := range cfg.RootKeys {
		k, err := newKey(id, val)
		if err != nil {
			return nil, err
		}
		if id == cfg.ActiveKeyID {
			ks.keys.Store(ActiveKeyID, k)
		}
		ks.keys.Store(id, k)
	}
	return ks, nil
}

// Find a root key with the given identifier.
func (ks *Keystore) Find(id string) (RootKey, error) {
	k, ok := ks.keys.Load(id)
	if !ok {
		return nil, ErrRootKeyNotFound
	}
	rk, ok := k.(*key)
	if !ok {
		return nil, errors.New("keystore: corrupted entry")
	}
	return rk, nil
}

type key struct {
	id    string
	value []byte
	aead  cipher.AEAD
}

// newKey builds the AEAD once, so the AES key schedule is not recomputed on
// every wrap and unwrap.
func newKey(id string, value []byte) (*key, error) {
	block, err := aes.NewCipher(value)
	if err != nil {
		return nil, ErrNewCipherBlock
	}
	aead, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, ErrNewGCMWithRandomNonce
	}
	return &key{id: id, value: value, aead: aead}, nil
}

// ID is a getter which returns the keys unique identifier.
func (k *key) ID() string {
	return k.id
}

// Encrypt the given data using the current root key.
func (k *key) Encrypt(d []byte) ([]byte, error) {
	return k.aead.Seal(nil, nil, d, nil), nil
}

// Decrypt the given data using the current root key.
func (k *key) Decrypt(d []byte) ([]byte, error) {
	pt, err := k.aead.Open(nil, nil, d, nil)
	if err != nil {
		return nil, ErrGCMOpen
	}
	return pt, nil
}
