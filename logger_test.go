package praetorian_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/karlbateman/praetorian"
)

func TestNewLogger(t *testing.T) {
	var buff bytes.Buffer
	log.SetOutput(&buff)
	defer log.SetOutput(nil)

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	logger := praetorian.NewLogger(mockHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "127.0.0.1:3000"
	rec := httptest.NewRecorder()

	logger.ServeHTTP(rec, req)

	time.Sleep(10 * time.Millisecond)
	out := buff.String()

	wantSubstring := "GET /test 418"
	if !strings.Contains(out, wantSubstring) {
		t.Errorf("NewLogger() log = %q, wantSubstring = %q", out, wantSubstring)
	}

	wantIP := "from 127.0.0.1:3000"
	if !strings.Contains(out, wantIP) {
		t.Errorf("NewLogger() log = %q, wantIP = %q", out, wantIP)
	}
}
