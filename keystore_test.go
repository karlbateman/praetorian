// Copyright © 2025 Karl Bateman. All Rights Reserved. Use of this software is
// governed by a BSD-style license that can be found in the LICENSE file.
package praetorian_test

import (
	"errors"
	"testing"

	"github.com/karlbateman/praetorian"
)

func TestNewKeystore(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *praetorian.Config
		wantErr error
	}{
		{
			name: "invalid key material",
			cfg: &praetorian.Config{
				ActiveKeyID: "1",
				RootKeys:    map[string][]byte{"1": []byte("too short")},
			},
			wantErr: praetorian.ErrNewCipherBlock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := praetorian.NewKeystore(tt.cfg)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewKeystore() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

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
				t.Fatalf("NewConfig() failed to create config: %v", err)
			}
			ks, err := praetorian.NewKeystore(cfg)
			if err != nil {
				t.Fatalf("NewKeystore() failed to create keystore: %v", err)
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
				t.Fatalf("NewConfig() failed to create config: %v", err)
			}
			ks, err := praetorian.NewKeystore(cfg)
			if err != nil {
				t.Fatalf("NewKeystore() failed to create keystore: %v", err)
			}

			k, err := ks.Find("1")
			if err != nil {
				t.Fatalf("Keystore.Find() failed to return key: %v", err)
			}

			result, err := k.Encrypt(tt.data)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("Key.Encrypt() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if len(result) == 0 {
				t.Errorf("Key.Encrypt() did not return an encrypted result")
			}
		})
	}
}

// FuzzKeyDecrypt exercises the one boundary in this package that handles
// fully attacker-controlled bytes: the token an /unwrap caller supplies is
// base64-decoded and passed straight to Decrypt. The goal is not a specific
// return value but that malformed or malicious ciphertext is always
// reported as an error and never panics.
func FuzzKeyDecrypt(f *testing.F) {
	f.Setenv(praetorian.EnvKey, testConfig)
	cfg, err := praetorian.NewConfig()
	if err != nil {
		f.Fatalf("NewConfig() failed to create config: %v", err)
	}
	ks, err := praetorian.NewKeystore(cfg)
	if err != nil {
		f.Fatalf("NewKeystore() failed to create keystore: %v", err)
	}
	k, err := ks.Find("1")
	if err != nil {
		f.Fatalf("Keystore.Find() failed to return key: %v", err)
	}

	f.Add([]byte(""))
	f.Add([]byte("not encrypted data"))

	enc, err := k.Encrypt([]byte("a secret never to be told"))
	if err != nil {
		f.Fatalf("Key.Encrypt() failed to encrypt seed data: %v", err)
	}
	f.Add(enc)

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = k.Decrypt(data)
	})
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
				t.Fatalf("NewConfig() failed to create config: %v", err)
			}
			ks, err := praetorian.NewKeystore(cfg)
			if err != nil {
				t.Fatalf("NewKeystore() failed to create keystore: %v", err)
			}

			k, err := ks.Find("1")
			if err != nil {
				t.Fatalf("Keystore.Find() failed to return key: %v", err)
			}

			enc, err := k.Encrypt(tt.data)
			if err != nil {
				t.Fatalf("Key.Encrypt() failed to encrypt data: %v", err)
			}
			if len(enc) == 0 {
				t.Errorf("Key.Encrypt() did not return an encrypted result")
			}

			result, err := k.Decrypt(enc)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
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

// BenchmarkKey_Encrypt guards against regressions in the per-call cost of
// Encrypt now that the AEAD is built once in NewKeystore rather than on
// every call; b.N should not scale with AES key-schedule setup cost.
func BenchmarkKey_Encrypt(b *testing.B) {
	b.Setenv(praetorian.EnvKey, testConfig)
	cfg, err := praetorian.NewConfig()
	if err != nil {
		b.Fatalf("NewConfig() failed to create config: %v", err)
	}
	ks, err := praetorian.NewKeystore(cfg)
	if err != nil {
		b.Fatalf("NewKeystore() failed to create keystore: %v", err)
	}
	k, err := ks.Find("1")
	if err != nil {
		b.Fatalf("Keystore.Find() failed to return key: %v", err)
	}
	data := []byte("a secret never to be told")

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := k.Encrypt(data); err != nil {
			b.Fatalf("Key.Encrypt() failed: %v", err)
		}
	}
}

// BenchmarkKey_Decrypt mirrors BenchmarkKey_Encrypt for the decrypt path.
func BenchmarkKey_Decrypt(b *testing.B) {
	b.Setenv(praetorian.EnvKey, testConfig)
	cfg, err := praetorian.NewConfig()
	if err != nil {
		b.Fatalf("NewConfig() failed to create config: %v", err)
	}
	ks, err := praetorian.NewKeystore(cfg)
	if err != nil {
		b.Fatalf("NewKeystore() failed to create keystore: %v", err)
	}
	k, err := ks.Find("1")
	if err != nil {
		b.Fatalf("Keystore.Find() failed to return key: %v", err)
	}
	enc, err := k.Encrypt([]byte("a secret never to be told"))
	if err != nil {
		b.Fatalf("Key.Encrypt() failed to encrypt seed data: %v", err)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := k.Decrypt(enc); err != nil {
			b.Fatalf("Key.Decrypt() failed: %v", err)
		}
	}
}
