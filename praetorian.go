package praetorian

import "errors"

const (
	ActiveKeyID   = "active"
	EnvKey        = "PRAETORIAN_CONFIG"
	RootKeyLength = 32
)

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

type RootKey interface {
	ID() string
	Decrypt(data []byte) ([]byte, error)
	Encrypt(data []byte) ([]byte, error)
}

// KeyFinder retrieves root keys from the an underlying keystore.
type KeyFinder interface {
	Find(id string) (RootKey, error)
}
