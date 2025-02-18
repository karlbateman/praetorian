package praetorian

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// ErrorResponse is returned from the HTTP server when an error occurs.
type ErrorResponse struct {
	Message string `json:"message"`
}

type server struct {
	*http.Server
	keys     KeyFinder
	mux      *http.ServeMux
	Shutdown func(context.Context) error
}

// NewServer allows wrapping and unwrapping to occur over a HTTP interface.
func NewServer(keys KeyFinder) *server {
	addr := fmt.Sprintf(":%s", port())
	mux := http.NewServeMux()

	srv := &server{
		keys: keys,
		mux:  mux,
	}
	srv.Routes()

	srv.Server = &http.Server{
		Addr:    addr,
		Handler: NewLogger(mux),
	}
	srv.Shutdown = srv.Server.Shutdown

	return srv
}

// Routes sets up HTTP endpoints and configures the respective handlers.
func (s *server) Routes() {
	s.mux.HandleFunc("/wrap", HandleWrap(ActiveKeyID, s.keys))
	s.mux.HandleFunc("/unwrap", HandleUnwrap(s.keys))
}

// Start launches the server which listens for HTTP requests.
func (s *server) Start() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Printf("listening on port %s...\n", port())
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("server error:", err)
		}
	}()

	<-stop
	log.Println("performing graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Println("forced shutdown:", err)
		return err
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
	encoder.Encode(data)
}
