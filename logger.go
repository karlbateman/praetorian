package praetorian

import (
	"log"
	"net/http"
	"time"
)

type LoggerResponse struct {
	http.ResponseWriter
	statusCode int
}

func (lr *LoggerResponse) WriteHeader(code int) {
	lr.statusCode = code
	lr.ResponseWriter.WriteHeader(code)
}

func NewLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lr := &LoggerResponse{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lr, r)
		log.Printf("%s %s %d %s from %s\n", r.Method, r.URL.Path, lr.statusCode, time.Since(start), r.RemoteAddr)
	})
}
