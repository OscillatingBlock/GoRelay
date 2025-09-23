package handler

import "net/http"

type RouteConfig struct {
	mux     *http.ServeMux
	handler *Handler
}

func NewRouteConfig(handler *Handler) *RouteConfig {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.ProxyHandler)
	mux.HandleFunc("/health", handler.HealthHandler)
	return &RouteConfig{
		mux:     mux,
		handler: handler,
	}
}

func (rc *RouteConfig) GetMux() *http.ServeMux {
	return rc.mux
}
