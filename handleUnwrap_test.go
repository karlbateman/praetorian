// Copyright © 2025 Karl Bateman. All Rights Reserved. Use of this software is
// governed by a BSD-style license that can be found in the LICENSE file.
package praetorian_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/karlbateman/praetorian"
)

func TestHandleUnwrap(t *testing.T) {
	tests := []struct {
		name       string
		body       io.Reader
		method     string
		wantCType  string
		wantResult string
		wantStatus int
	}{
		{
			name:       "success",
			body:       strings.NewReader(`{}`),
			method:     http.MethodPost,
			wantCType:  "application/json;charset=utf-8",
			wantStatus: http.StatusOK,
			wantResult: `{"value":"decrypted message"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ks := &MockKeystore{}
			handler := praetorian.HandleUnwrap(ks)

			req := httptest.NewRequest(tt.method, "/unwrap", tt.body)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("HandleUnwrap() status = %d, wantStatus = %d", rec.Code, tt.wantStatus)
			}

			cType := rec.Header().Get("Content-Type")
			if cType != tt.wantCType {
				t.Errorf("HandleUnwrap() cType = %q, wantCType = %q", cType, tt.wantCType)
			}

			body, err := io.ReadAll(rec.Body)
			if err != nil {
				t.Fatalf("HandleUnwrap() failed to read response body: %v", err)
			}

			got := strings.TrimSpace(string(body))
			want := strings.TrimSpace(tt.wantResult)
			if got != want {
				t.Errorf("HandleUnwrap() got = %v, wantResult = %v", got, want)
			}
		})
	}
}

// BenchmarkHandleUnwrap guards against regressions in the per-request cost
// of the /unwrap handler's happy path: decoding the request, decrypting,
// and encoding the response.
func BenchmarkHandleUnwrap(b *testing.B) {
	ks := &MockKeystore{}
	handler := praetorian.HandleUnwrap(ks)
	body := []byte(`{}`)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/unwrap", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func TestHandleUnwrap_Errors(t *testing.T) {
	tests := []struct {
		name        string
		body        io.Reader
		method      string
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "decryption failure",
			body:        strings.NewReader(`{"id": "1", "token": "ZXJyb3I="}`),
			method:      http.MethodPost,
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "decryption failed",
		},
		{
			name:        "invalid JSON body",
			body:        strings.NewReader("{invalid}"),
			method:      http.MethodPost,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "invalid JSON",
		},
		{
			name:        "unsupported HTTP method",
			body:        strings.NewReader("{}"),
			method:      http.MethodGet,
			wantStatus:  http.StatusNotFound,
			wantMessage: "Not Found",
		},
		{
			name:        "empty body",
			body:        http.NoBody,
			method:      http.MethodPost,
			wantStatus:  http.StatusBadRequest,
			wantMessage: "invalid JSON",
		},
		{
			name:        "missing root key",
			body:        strings.NewReader(`{"id": "missing", "token": "ZW5jcnlwdGVkIG1lc3NhZ2U="}`),
			method:      http.MethodPost,
			wantStatus:  http.StatusNotFound,
			wantMessage: "root key not found",
		},
		{
			name:        "invalid encrypted data",
			body:        strings.NewReader(`{"id": "1", "token": "b3Blbgo="}`),
			method:      http.MethodPost,
			wantStatus:  http.StatusUnprocessableEntity,
			wantMessage: "data authentication failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ks := &MockKeystore{}
			handler := praetorian.HandleUnwrap(ks)

			req := httptest.NewRequest(tt.method, "/unwrap", tt.body)
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
