package usecase

import (
	"GoRelay/internal/models"
	"GoRelay/pkg/utils"
	"net/http"
)

type HealthChecker interface {
	CheckHealth(backend *models.Backend) bool
}

type ConfigLoader interface {
	Load() (*utils.Config, error)
}

type RequestProxy interface {
	HandleRequest(req *http.Request, w http.ResponseWriter) error
	GetHealthyBackends() int
}
