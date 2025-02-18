// Copyright © 2025 Karl Bateman. All Rights Reserved. Use of this software is
// governed by a BSD-style license that can be found in the LICENSE file.
package praetorian

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
)

func HandleUnwrap(keys KeyFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var b WrapResponse
			if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
				jsonResponse(w, http.StatusBadRequest, &ErrorResponse{
					Message: "invalid JSON",
				})
				return
			}

			key, err := keys.Find(b.ID)
			if err != nil {
				jsonResponse(w, http.StatusNotFound, &ErrorResponse{
					Message: err.Error(),
				})
				return
			}

			token, err := base64.StdEncoding.DecodeString(b.Token)
			if err != nil {
				jsonResponse(w, http.StatusBadRequest, &ErrorResponse{
					Message: err.Error(),
				})
				return
			}

			dec, err := key.Decrypt(token)
			if err != nil {
				if errors.Is(err, ErrGCMOpen) {
					jsonResponse(w, http.StatusUnprocessableEntity, &ErrorResponse{
						Message: "data authentication failed",
					})
					return
				}
				jsonResponse(w, http.StatusInternalServerError, &ErrorResponse{
					Message: err.Error(),
				})
				return
			}

			jsonResponse(w, http.StatusOK, json.RawMessage(dec))
		default:
			jsonResponse(w, http.StatusNotFound, &ErrorResponse{
				Message: "Not Found",
			})
		}
	}
}
