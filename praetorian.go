// Copyright © 2025 Karl Bateman. All Rights Reserved. Use of this software is
// governed by a BSD-style license that can be found in the LICENSE file.

// Package praetorian implements a minimal key-wrapping service: it seals
// ("wraps") and opens ("unwraps") arbitrary JSON payloads under root keys
// that are held only in memory, and exposes both operations over HTTP.
//
// Root key material is loaded once at startup, from the PRAETORIAN_CONFIG
// environment variable (see NewConfig), and never touches disk. Each root
// key is 32 bytes of AES-256 key material; wrapping seals the request body
// with AES-256-GCM under a randomly generated nonce (see Keystore and
// RootKey), and unwrapping authenticates and opens a previously wrapped
// token, rejecting it if the ciphertext has been tampered with.
//
// A wrap response reports the identifier of the root key actually used to
// seal the data. A caller passes that same identifier back to unwrap, so
// data wrapped under a since-superseded key remains decryptable for as long
// as that key's material is still present in the configuration, even after
// ActiveKeyID has moved on to a newer key.
//
// The HTTP interface (see NewServer) exposes exactly two endpoints, POST
// /wrap and POST /unwrap. It performs no authentication of its own and is
// intended to run behind a trusted network boundary or an authenticating
// proxy.
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
