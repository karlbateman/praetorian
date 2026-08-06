// Copyright © 2025 Karl Bateman. All Rights Reserved. Use of this software is
// governed by a BSD-style license that can be found in the LICENSE file.

// The /wrap HTTP handler and its WrapResponse payload.

package praetorian

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// WrapResponse is returned from the HTTP server after a successful wrap
// operation.
type WrapResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

// HandleWrap encrypts the POSTed request body under activeKey and returns
// the resulting token.
func HandleWrap(activeKey string, keys KeyFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			jsonResponse(w, http.StatusNotFound, &ErrorResponse{
				Message: "Not Found",
			})
			return
		}

		maxBytes := int64(1 << 20) // 1MB limit
		b, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBytes))
		if err != nil {
			jsonResponse(w, http.StatusBadRequest, &ErrorResponse{
				Message: "failed to read request body",
			})
			return
		}

		if !json.Valid(b) {
			jsonResponse(w, http.StatusBadRequest, &ErrorResponse{
				Message: "invalid JSON",
			})
			return
		}

		key, err := keys.Find(activeKey)
		if err != nil {
			log.Printf("wrap: find active key: %v", err)
			jsonResponse(w, http.StatusInternalServerError, &ErrorResponse{
				Message: "unable to wrap data",
			})
			return
		}

		enc, err := key.Encrypt(b)
		if err != nil {
			log.Printf("wrap: encrypt: %v", err)
			jsonResponse(w, http.StatusInternalServerError, &ErrorResponse{
				Message: "unable to wrap data",
			})
			return
		}

		token := base64.StdEncoding.EncodeToString(enc)
		jsonResponse(w, http.StatusCreated, &WrapResponse{
			ID:    key.ID(),
			Token: token,
		})
	}
}
