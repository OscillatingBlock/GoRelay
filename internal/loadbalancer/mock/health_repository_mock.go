package mock

import "GoRelay/internal/models"

type HealthRepositoryMock struct {
	CheckHealthFunc func(*models.Backend) bool
}

func (m *HealthRepositoryMock) CheckHealth(backend *models.Backend) bool {
	return m.CheckHealthFunc(backend)
}
