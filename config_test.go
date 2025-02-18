package praetorian_test

import (
	"errors"
	"testing"

	"github.com/karlbateman/praetorian"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr error
	}{
		{
			name:    "empty config",
			config:  ``,
			wantErr: praetorian.ErrEnvConfigEmpty,
		},
		{
			name:    "invalid config",
			config:  `<></>`,
			wantErr: praetorian.ErrEnvConfigInvalid,
		},
		{
			name:    "invalid root key",
			config:  `{"activeKeyId": "1", "rootKeys": {"1": "123"}}`,
			wantErr: praetorian.ErrInvalidRootKey,
		},
		{
			name:    "invalid root key length",
			config:  `{"activeKeyId": "1", "rootKeys": {"1": "BPK//lj6hlpjuA5gPZo19OIjUDgnIQ=="}}`,
			wantErr: praetorian.ErrInvalidRootKeyLength,
		},
		{
			name:    "root key not found",
			config:  `{"activeKeyId": "2", "rootKeys": {"1": "kSRFQxepULO9UC5SL5pA/mXjbI1GXu9ha2T0yPr3scU="}}`,
			wantErr: praetorian.ErrActiveRootKeyNotFound,
		},
		{
			name:    "valid config",
			config:  `{"activeKeyId": "1", "rootKeys": {"1": "kSRFQxepULO9UC5SL5pA/mXjbI1GXu9ha2T0yPr3scU="}}`,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(praetorian.EnvKey, tt.config)
			_, err := praetorian.NewConfig()
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("NewConfig() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
