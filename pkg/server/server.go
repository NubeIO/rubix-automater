package server

import (
	"context"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// HTTPServer represents an HTTP server.
type HTTPServer struct {
	srv    http.Server
	logger *logrus.Logger
}

// NewHTTPServer creates and returns a new HTTPServer instance.
func NewHTTPServer(srv http.Server, logger *logrus.Logger) *HTTPServer {
	return &HTTPServer{
		srv:    srv,
		logger: logger,
	}
}

// Serve starts the server.
func (s *HTTPServer) Serve() {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil {
			s.logger.Printf("failed to serve HTTP: %s", err)
		}
	}()
}

// GracefullyStop gracefully stops the server.
func (s *HTTPServer) GracefullyStop() {
	s.logger.Println("server exiting...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.Fatal("failed to properly shutdown the server:", err)
	}
}
