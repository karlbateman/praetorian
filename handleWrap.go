package praetorian

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
)

type WrapResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

func HandleWrap(activeKey string, keys KeyFinder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			maxBytes := int64(1 << 20) // 1MB limit
			b, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBytes))
			if err != nil {
				jsonResponse(w, http.StatusBadRequest, &ErrorResponse{
					Message: "failed to read request body",
				})
				return
			}
			defer r.Body.Close()

			if !json.Valid(b) {
				jsonResponse(w, http.StatusBadRequest, &ErrorResponse{
					Message: "invalid JSON",
				})
				return
			}

			key, err := keys.Find(activeKey)
			if err != nil {
				jsonResponse(w, http.StatusNotFound, &ErrorResponse{
					Message: err.Error(),
				})
				return
			}

			enc, err := key.Encrypt(b)
			if err != nil {
				jsonResponse(w, http.StatusInternalServerError, &ErrorResponse{
					Message: err.Error(),
				})
				return
			}

			token := base64.StdEncoding.EncodeToString(enc)
			jsonResponse(w, http.StatusCreated, &WrapResponse{
				ID:    key.ID(),
				Token: token,
			})
		default:
			jsonResponse(w, http.StatusNotFound, &ErrorResponse{
				Message: "Not Found",
			})
		}
	}
}
