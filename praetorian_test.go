package praetorian_test

import (
	"errors"
	"io"
	"strings"
	"time"

	"github.com/karlbateman/praetorian"
)

const (
	testConfig = `{"activeKeyId": "1", "rootKeys": {"1": "kSRFQxepULO9UC5SL5pA/mXjbI1GXu9ha2T0yPr3scU="}}`
)

type MockReader struct {
	Data  []byte
	Err   error
	Delay time.Duration
}

func (m *MockReader) Read(p []byte) (n int, err error) {
	if m.Delay > 0 {
		time.Sleep(m.Delay)
	}
	if m.Err != nil {
		return 0, m.Err
	}
	n = copy(p, m.Data)
	m.Data = m.Data[n:]
	if len(m.Data) == 0 {
		return n, io.EOF
	}
	return n, nil
}

type MockKeystore struct{}

func (m *MockKeystore) Find(id string) (praetorian.RootKey, error) {
	if id == "missing" {
		return nil, errors.New("root key not found")
	}
	return &MockKey{}, nil
}

type MockKey struct{}

func (k *MockKey) ID() string {
	return "1"
}

func (k *MockKey) Encrypt(data []byte) ([]byte, error) {
	if strings.Contains(string(data), "error") {
		return nil, errors.New("encryption failed")
	}
	return []byte(`encrypted message`), nil
}

func (k *MockKey) Decrypt(data []byte) ([]byte, error) {
	if strings.Contains(string(data), "open") {
		return nil, praetorian.ErrGCMOpen
	}
	if strings.Contains(string(data), "error") {
		return nil, errors.New("decryption failed")
	}
	return []byte(`{"value": "decrypted message"}`), nil
}
