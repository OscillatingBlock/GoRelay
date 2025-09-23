package handler

import (
	"GoRelay/internal/loadbalancer/usecase"
	"GoRelay/pkg/logger"
	"fmt"
	"net/http"
)

type Handler struct {
	uc     usecase.RequestProxy
	logger *logger.Logger
}

func NewHandler(uc *usecase.LoadBalancerUseCase, logger *logger.Logger) *Handler {
	return &Handler{
		uc:     uc,
		logger: logger,
	}
}

func (h *Handler) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	err := h.uc.HandleRequest(r, w)
	if err != nil {
		h.logger.Error("Error while serving Request", "error", err)
		w.WriteHeader(http.StatusBadGateway)
	}
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	healthyCount := h.uc.GetHealthyBackends()
	if healthyCount == 0 {
		h.logger.Warn("No healthy backends available")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "0")
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%d", healthyCount)
}
