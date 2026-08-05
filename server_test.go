// Copyright © 2025 Karl Bateman. All Rights Reserved. Use of this software is
// governed by a BSD-style license that can be found in the LICENSE file.
package praetorian_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/karlbateman/praetorian"
)

func TestNewServer(t *testing.T) {
	t.Setenv(praetorian.EnvKey, testConfig)
	t.Setenv("PORT", "8080")

	cfg, err := praetorian.NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() failed to create config: %v", err)
	}
	ks, err := praetorian.NewKeystore(cfg)
	if err != nil {
		t.Fatalf("NewKeystore() failed to create keystore: %v", err)
	}
	srv := praetorian.NewServer(ks)

	wantAddr := ":8080"
	if srv.Server.Addr != wantAddr {
		t.Errorf("Server.Addr got = %q, wantAddr = %q", srv.Addr, wantAddr)
	}
}

func TestServer_StartGracefulShutdown(t *testing.T) {
	t.Setenv(praetorian.EnvKey, testConfig)
	t.Setenv("PORT", "8080")

	var buff bytes.Buffer
	log.SetOutput(&buff)
	defer log.SetOutput(nil)

	cfg, err := praetorian.NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() failed to create config: %v", err)
	}
	ks, err := praetorian.NewKeystore(cfg)
	if err != nil {
		t.Fatalf("NewKeystore() failed to create keystore: %v", err)
	}
	srv := praetorian.NewServer(ks)

	go func() {
		_ = srv.Start()
	}()

	time.Sleep(500 * time.Millisecond)
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("os.FindProcess() failed to return the current process: %v", err)
	}

	p.Signal(os.Interrupt)
	time.Sleep(500 * time.Millisecond)
	out := buff.String()

	wantShutdown := "performing graceful shutdown"
	if !strings.Contains(out, wantShutdown) {
		t.Errorf("Server.Start() log = %q, wantShutdown = %q", out, wantShutdown)
	}

	wantSuccess := "server shutdown successful"
	if !strings.Contains(out, wantSuccess) {
		t.Errorf("Server.Start() log = %q, wantSuccess = %q", out, wantSuccess)
	}
}

func TestServer_StartForcedShutdown(t *testing.T) {
	t.Setenv(praetorian.EnvKey, testConfig)
	t.Setenv("PORT", "8080")

	// Start() logs from a background goroutine regardless of this test's
	// assertions; give it a sink so it doesn't write to a stale output left
	// behind by a previous test's cleanup.
	log.SetOutput(io.Discard)
	defer log.SetOutput(nil)

	cfg, err := praetorian.NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() failed to create config: %v", err)
	}
	ks, err := praetorian.NewKeystore(cfg)
	if err != nil {
		t.Fatalf("NewKeystore() failed to create keystore: %v", err)
	}

	srv := praetorian.NewServer(ks)
	srv.Shutdown = func(ctx context.Context) error {
		return fmt.Errorf("mock forced shutdown error")
	}

	startErr := make(chan error, 1)
	go func() {
		startErr <- srv.Start()
	}()

	time.Sleep(500 * time.Millisecond)
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("os.FindProcess() failed to return the current process: %v", err)
	}
	p.Signal(os.Interrupt)

	err = <-startErr
	wantSubstring := "forced shutdown"
	if err == nil || !strings.Contains(err.Error(), wantSubstring) {
		t.Errorf("Server.Start() error = %v, want substring %q", err, wantSubstring)
	}
}
