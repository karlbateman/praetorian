package praetorian_test

import (
	"bytes"
	"context"
	"fmt"
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
		t.Errorf("NewConfig() failed to create config: %v", err)
	}
	ks, err := praetorian.NewKeystore(cfg)
	if err != nil {
		t.Errorf("NewKeystore() failed to create keystore: %v", err)
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
		t.Errorf("NewConfig() failed to create config: %v", err)
	}
	ks, err := praetorian.NewKeystore(cfg)
	if err != nil {
		t.Errorf("NewKeystore() failed to create keystore: %v", err)
	}
	srv := praetorian.NewServer(ks)

	go func() {
		_ = srv.Start()
	}()

	time.Sleep(500 * time.Millisecond)
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Errorf("os.FindProcess() failed to return the current process: %v", err)
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

	var buff bytes.Buffer
	log.SetOutput(&buff)
	defer log.SetOutput(nil)

	cfg, err := praetorian.NewConfig()
	if err != nil {
		t.Errorf("NewConfig() failed to create config: %v", err)
	}
	ks, err := praetorian.NewKeystore(cfg)
	if err != nil {
		t.Errorf("NewKeystore() failed to create keystore: %v", err)
	}

	srv := praetorian.NewServer(ks)
	srv.Shutdown = func(ctx context.Context) error {
		return fmt.Errorf("mock forced shutdown error")
	}

	go func() {
		_ = srv.Start()
	}()

	time.Sleep(500 * time.Millisecond)
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Errorf("os.FindProcess() failed to return the current process: %v", err)
	}

	p.Signal(os.Interrupt)
	time.Sleep(500 * time.Millisecond)
	out := buff.String()

	wantShutdown := "forced shutdown"
	if !strings.Contains(out, wantShutdown) {
		t.Errorf("Server.Start() log = %q, wantShutdown = %q", out, wantShutdown)
	}
}
