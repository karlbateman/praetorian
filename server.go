// Copyright © 2025 Karl Bateman. All Rights Reserved. Use of this software is
// governed by a BSD-style license that can be found in the LICENSE file.

// Server wires the wrap and unwrap handlers onto an HTTP server and manages
// its start-up and graceful-shutdown lifecycle.

package praetorian

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ErrorResponse is returned from the HTTP server when an error occurs.
type ErrorResponse struct {
	Message string `json:"message"`
}

// Server serves the wrap and unwrap operations over an HTTP interface.
type Server struct {
	*http.Server
	keys     KeyFinder
	mux      *http.ServeMux
	Shutdown func(context.Context) error
}

// NewServer allows wrapping and unwrapping to occur over a HTTP interface.
func NewServer(keys KeyFinder) *Server {
	addr := fmt.Sprintf(":%s", port())
	mux := http.NewServeMux()

	srv := &Server{
		keys: keys,
		mux:  mux,
	}
	srv.routes()

	srv.Server = &http.Server{
		Addr:    addr,
		Handler: NewLogger(mux),
	}
	srv.Shutdown = srv.Server.Shutdown

	return srv
}

// routes sets up HTTP endpoints and configures the respective handlers.
func (s *Server) routes() {
	s.mux.HandleFunc("/wrap", HandleWrap(ActiveKeyID, s.keys))
	s.mux.HandleFunc("/unwrap", HandleUnwrap(s.keys))
}

// Start launches the server which listens for HTTP requests, and blocks
// until it either fails to bind or receives a shutdown signal.
func (s *Server) Start() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	listenErr := make(chan error, 1)
	go func() {
		log.Printf("listening on %s...\n", s.Server.Addr)
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			listenErr <- err
			return
		}
		listenErr <- nil
	}()

	select {
	case err := <-listenErr:
		if err != nil {
			return fmt.Errorf("listen: %w", err)
		}
		return nil
	case <-stop:
	}

	log.Println("performing graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		return fmt.Errorf("forced shutdown: %w", err)
	}

	log.Println("server shutdown successful")
	return nil
}

func port() string {
	val := os.Getenv("PORT")
	if val == "" {
		val = "3000"
	}
	return val
}

func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(data); err != nil {
		// The status and headers are already written, so this can only be logged.
		log.Println("failed to encode JSON response:", err)
	}
}
