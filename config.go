// Copyright © 2025 Karl Bateman. All Rights Reserved. Use of this software is
// governed by a BSD-style license that can be found in the LICENSE file.

// Config and NewConfig decode root key material from the environment into
// the form the rest of the package operates on.

package praetorian

import (
	"encoding/base64"
	"encoding/json"
	"os"
)

// Config is a key configuration decoded from the environment.
type Config struct {
	ActiveKeyID string
	RootKeys    map[string][]byte
}

// NewConfig returns a key configuration from the environment.
func NewConfig() (*Config, error) {
	val := os.Getenv(EnvKey)
	if val == "" {
		return nil, ErrEnvConfigEmpty
	}

	// represents the JSON structure set in the environment.
	var env struct {
		ActiveKeyID string            `json:"activeKeyId"`
		RootKeys    map[string]string `json:"rootKeys"`
	}

	if err := json.Unmarshal([]byte(val), &env); err != nil {
		return nil, ErrEnvConfigInvalid
	}

	if _, ok := env.RootKeys[env.ActiveKeyID]; !ok {
		return nil, ErrActiveRootKeyNotFound
	}

	c := &Config{
		ActiveKeyID: env.ActiveKeyID,
		RootKeys:    make(map[string][]byte),
	}
	for i, m := range env.RootKeys {
		// ActiveKeyID is used as an alias for whichever key is active, so a
		// non-active key claiming that id would collide in the keystore. Map
		// iteration order is randomised, making the winner vary per process.
		if i == ActiveKeyID && i != env.ActiveKeyID {
			return nil, ErrReservedRootKeyID
		}
		k, err := base64.StdEncoding.DecodeString(m)
		if err != nil {
			return nil, ErrInvalidRootKey
		}
		if len(k) != RootKeyLength {
			return nil, ErrInvalidRootKeyLength
		}
		c.RootKeys[i] = k
	}
	return c, nil
}
