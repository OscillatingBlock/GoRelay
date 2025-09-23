package server

import (
	handler "GoRelay/internal/loadbalancer/delivery"
	"GoRelay/pkg/logger"
	"context"
	"net/http"
)

type Server struct {
	srv    *http.Server
	routes *handler.RouteConfig
	logger *logger.Logger
}

func NewServer(routes *handler.RouteConfig, logger *logger.Logger) *Server {
	srv := &http.Server{
		Handler: routes.GetMux(),
	}
	return &Server{
		srv:    srv,
		routes: routes,
		logger: logger,
	}
}

func (s *Server) Start(port string) error {
	s.srv.Addr = ":" + port
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
