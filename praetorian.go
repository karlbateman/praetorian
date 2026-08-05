// Copyright © 2025 Karl Bateman. All Rights Reserved. Use of this software is
// governed by a BSD-style license that can be found in the LICENSE file.
package praetorian

import "errors"

// ActiveKeyID is the identifier under which the currently active root key is
// stored, in addition to its own configured identifier.
const ActiveKeyID = "active"

// EnvKey is the name of the environment variable holding the JSON-encoded
// key configuration.
const EnvKey = "PRAETORIAN_CONFIG"

// RootKeyLength is the required length, in bytes, of a decoded root key.
const RootKeyLength = 32

// Sentinel errors returned by this package's configuration, keystore, and
// encryption operations.
var (
	ErrActiveRootKeyNotFound = errors.New("active key does not exist in root keys")
	ErrEnvConfigEmpty        = errors.New("env config not set or empty")
	ErrEnvConfigInvalid      = errors.New("unable to parse config data")
	ErrInvalidRootKey        = errors.New("unable to decode root key")
	ErrInvalidRootKeyLength  = errors.New("root key length must be 32 bytes")
	ErrRootKeyNotFound       = errors.New("root key not found")
	ErrNewCipherBlock        = errors.New("unable to create AES-256 cipher block")
	ErrNewGCMWithRandomNonce = errors.New("unable to create cipher with Galois-Counter-Mode")
	ErrGCMOpen               = errors.New("unable to read encrypted data")
)

// RootKey is used to perform encryption and decryption of inputs.
//
// To reduce the risk of error or misuse, the underlying AES-256 algorithm uses
// a managed nonce value. Therefore the RootKey MUST NOT be used to encrypt
// more than 2^32 messages, to limit the risk of random nonce collisions to
// negligible levels.
type RootKey interface {
	ID() string
	Decrypt(data []byte) ([]byte, error)
	Encrypt(data []byte) ([]byte, error)
}

// KeyFinder retrieves root keys from an underlying keystore.
type KeyFinder interface {
	Find(id string) (RootKey, error)
}
