package praetorian_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/karlbateman/praetorian"
)

func TestHandleWrap(t *testing.T) {
	tests := []struct {
		name       string
		body       io.Reader
		method     string
		wantToken  string
		wantStatus int
	}{
		{
			name:       "success",
			body:       strings.NewReader(`{"value": "keep it secret, keep it safe"}`),
			method:     http.MethodPost,
			wantToken:  base64.StdEncoding.EncodeToString([]byte("encrypted message")),
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ks := &MockKeystore{}
			handler := praetorian.HandleWrap(praetorian.ActiveKeyID, ks)

			req := httptest.NewRequest(tt.method, "/wrap", tt.body)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("HandleWrap() status = %d, wantStatus = %d", rec.Code, tt.wantStatus)
			}

			var res praetorian.WrapResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
				t.Fatalf("HandleWrap() failed to parse response: %v", err)
			}

			wantID := "1"
			if res.ID != wantID {
				t.Errorf("HandleWrap() key = %q, wantID = %q", res.ID, wantID)
			}
			if res.Token != tt.wantToken {
				t.Errorf("HandleWrap() token = %q, wantToken = %q", res.Token, tt.wantToken)
			}
		})
	}
}

func TestHandleWrap_Errors(t *testing.T) {
	tests := []struct {
		name        string
		activeKey   string
		body        io.Reader
		method      string
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "encryption failure",
			activeKey:   praetorian.ActiveKeyID,
			body:        strings.NewReader(`{"message": "error"}`),
			method:      http.MethodPost,
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "encryption failed",
		},
		{
			name:        "invalid JSON body",
			activeKey:   praetorian.ActiveKeyID,
			body:        strings.NewReader("{invalid}"),
			method:      http.MethodPost,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "invalid JSON",
		},
		{
			name:        "unsupported HTTP method",
			activeKey:   praetorian.ActiveKeyID,
			body:        strings.NewReader("{}"),
			method:      http.MethodGet,
			wantStatus:  http.StatusNotFound,
			wantMessage: "Not Found",
		},
		{
			name:        "empty body",
			activeKey:   praetorian.ActiveKeyID,
			body:        http.NoBody,
			method:      http.MethodPost,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "invalid JSON",
		},
		{
			name:        "body too large",
			activeKey:   praetorian.ActiveKeyID,
			body:        bytes.NewReader(make([]byte, 2<<20)),
			method:      http.MethodPost,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "failed to read request body",
		},
		{
			name:        "client disconnected",
			activeKey:   praetorian.ActiveKeyID,
			body:        &MockReader{Err: io.ErrUnexpectedEOF},
			method:      http.MethodPost,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "failed to read request body",
		},
		{
			name:        "read timeout",
			activeKey:   praetorian.ActiveKeyID,
			body:        &MockReader{Delay: 2 * time.Second, Err: io.ErrUnexpectedEOF},
			method:      http.MethodPost,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "failed to read request body",
		},
		{
			name:        "active key not found",
			activeKey:   "missing",
			body:        strings.NewReader(`{}`),
			method:      http.MethodPost,
			wantStatus:  http.StatusNotFound,
			wantMessage: "root key not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ks := &MockKeystore{}
			handler := praetorian.HandleWrap(tt.activeKey, ks)

			req := httptest.NewRequest(tt.method, "/wrap", tt.body)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("HandleWrap() status = %d, wantStatus = %d", rec.Code, tt.wantStatus)
			}

			var res praetorian.ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
				t.Fatalf("HandleWrap() failed to parse response: %v", err)
			}

			if res.Message != tt.wantMessage {
				t.Errorf("HandleWrap() got = %q, wantMessage = %q", res.Message, tt.wantMessage)
			}
		})
	}
}
