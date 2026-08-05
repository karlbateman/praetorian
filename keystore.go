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
		if id == cfg.ActiveKeyID {
			ks.keys.Store(ActiveKeyID, &key{id, val})
		}
		ks.keys.Store(id, &key{id, val})
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
}

// ID is a getter which returns the keys unique identifier.
func (k *key) ID() string {
	return k.id
}

// Encrypt the given data using the current root key.
func (k *key) Encrypt(d []byte) ([]byte, error) {
	gcm, err := k.gcm()
	if err != nil {
		return nil, err
	}
	return gcm.Seal(nil, nil, d, nil), nil
}

// Decrypt the given data using the current root key.
func (k *key) Decrypt(d []byte) ([]byte, error) {
	gcm, err := k.gcm()
	if err != nil {
		return nil, err
	}
	ci, err := gcm.Open(nil, nil, d, nil)
	if err != nil {
		return nil, ErrGCMOpen
	}
	return ci, nil
}

// gcm builds the AEAD cipher used to encrypt and decrypt with this key.
func (k *key) gcm() (cipher.AEAD, error) {
	block, err := aes.NewCipher(k.value)
	if err != nil {
		return nil, ErrNewCipherBlock
	}
	gcm, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, ErrNewGCMWithRandomNonce
	}
	return gcm, nil
}
