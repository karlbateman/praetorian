package praetorian

import (
	"encoding/base64"
	"encoding/json"
	"os"
)

type config struct {
	ActiveKeyID string
	RootKeys    map[string][]byte
}

// NewConfig returns a key configuration from the environment.
func NewConfig() (*config, error) {
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

	c := &config{
		ActiveKeyID: env.ActiveKeyID,
		RootKeys:    make(map[string][]byte),
	}

	if _, ok := env.RootKeys[env.ActiveKeyID]; ok {
		for i, m := range env.RootKeys {
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
	return nil, ErrActiveRootKeyNotFound
}
