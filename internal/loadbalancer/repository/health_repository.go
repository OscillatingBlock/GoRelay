package repository

import (
	"GoRelay/internal/models"
	"GoRelay/pkg/logger"
	"net/http"
	"time"
)

type HealthRepository struct {
	client *http.Client
	logger *logger.Logger
}

func NewHealthRepository(logger *logger.Logger) *HealthRepository {
	var httpClient = &http.Client{
		Timeout: 5 * time.Second,
	}
	return &HealthRepository{
		logger: logger,
		client: httpClient,
	}
}

func (r *HealthRepository) CheckHealth(backend *models.Backend) bool {
	resp, err := r.client.Head(backend.URL.String())
	if err != nil {
		r.logger.Error("error sending HEAD request to backend", "url", backend.URL.String(), "error", err)
		return false
	}
	defer resp.Body.Close()
	isHealthy := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !isHealthy {
		r.logger.Warn("backend unhealthy", "url", backend.URL.String(), "status", resp.StatusCode)
	}
	return isHealthy
}
