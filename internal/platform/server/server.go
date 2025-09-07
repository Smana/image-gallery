package server

import (
	"net/http"
	"time"
)

func New(port string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}