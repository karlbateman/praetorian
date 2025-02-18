package praetorian_test

import (
	"errors"
	"testing"

	"github.com/karlbateman/praetorian"
)

func TestKeystore_Find(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		id      string
		wantErr error
	}{
		{
			name:    "key does not exist",
			config:  `{"activeKeyId": "1", "rootKeys": {"1": "kSRFQxepULO9UC5SL5pA/mXjbI1GXu9ha2T0yPr3scU="}}`,
			id:      "2",
			wantErr: praetorian.ErrRootKeyNotFound,
		},
		{
			name:    "key exists",
			config:  `{"activeKeyId": "1", "rootKeys": {"1": "kSRFQxepULO9UC5SL5pA/mXjbI1GXu9ha2T0yPr3scU="}}`,
			id:      "1",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(praetorian.EnvKey, tt.config)
			cfg, err := praetorian.NewConfig()
			if err != nil {
				t.Errorf("NewConfig() failed to create config: %v", err)
			}
			ks, err := praetorian.NewKeystore(cfg)
			if err != nil {
				t.Errorf("NewKeystore() failed to create keystore: %v", err)
			}

			_, err = ks.Find(tt.id)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("Keystore.Find() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestKey_Encrypt(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		{
			name:    "encrypt succeeds",
			data:    []byte("a secret never to be told"),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(praetorian.EnvKey, testConfig)
			cfg, err := praetorian.NewConfig()
			if err != nil {
				t.Errorf("NewConfig() failed to create config: %v", err)
			}
			ks, err := praetorian.NewKeystore(cfg)
			if err != nil {
				t.Errorf("NewKeystore() failed to create keystore: %v", err)
			}

			k, err := ks.Find("1")
			if err != nil {
				t.Errorf("Keystore.Find() failed to return key: %v", err)
			}

			result, err := k.Encrypt(tt.data)
			if tt.wantErr != nil && errors.Is(err, tt.wantErr) {
				t.Errorf("Key.Encrypt() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if len(result) == 0 {
				t.Errorf("Key.Encrypt() did not return an encrypted result")
			}
		})
	}
}

func TestKey_Decrypt(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		{
			name:    "decrypt succeeds",
			data:    []byte("a secret never to be told"),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(praetorian.EnvKey, `{"activeKeyId": "1", "rootKeys": {"1": "kSRFQxepULO9UC5SL5pA/mXjbI1GXu9ha2T0yPr3scU="}}`)
			cfg, err := praetorian.NewConfig()
			if err != nil {
				t.Errorf("NewConfig() failed to create config: %v", err)
			}
			ks, err := praetorian.NewKeystore(cfg)
			if err != nil {
				t.Errorf("NewKeystore() failed to create keystore: %v", err)
			}

			k, err := ks.Find("1")
			if err != nil {
				t.Errorf("Keystore.Find() failed to return key: %v", err)
			}

			enc, err := k.Encrypt(tt.data)
			if err != nil {
				t.Errorf("Key.Encrypt() failed to encrypt data: %v", err)
			}
			if len(enc) == 0 {
				t.Errorf("Key.Encrypt() did not return an encrypted result")
			}

			result, err := k.Decrypt(enc)
			if tt.wantErr != nil && errors.Is(err, tt.wantErr) {
				t.Errorf("Key.Decrypt() error = %v, wantErr = %v", err, tt.wantErr)
			}

			got := string(result)
			wantData := string(tt.data)
			if got != wantData {
				t.Errorf("Key.Decrypt() got = %v, wantData = %v", got, wantData)
			}
		})
	}
}
