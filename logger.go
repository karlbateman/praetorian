// Copyright © 2025 Karl Bateman. All Rights Reserved. Use of this software is
// governed by a BSD-style license that can be found in the LICENSE file.

// Request-logging middleware shared by every route registered on Server.

package praetorian

import (
	"log"
	"net/http"
	"time"
)

// loggerResponse captures the status code written by the wrapped handler so
// it can be included in the request log line.
type loggerResponse struct {
	http.ResponseWriter
	statusCode int
}

func (lr *loggerResponse) WriteHeader(code int) {
	lr.statusCode = code
	lr.ResponseWriter.WriteHeader(code)
}

// NewLogger wraps next with a handler that logs each request's method, path,
// status code, duration, and remote address.
func NewLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lr := &loggerResponse{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lr, r)
		log.Printf("%s %s %d %s from %s\n", r.Method, r.URL.Path, lr.statusCode, time.Since(start), r.RemoteAddr)
	})
}
