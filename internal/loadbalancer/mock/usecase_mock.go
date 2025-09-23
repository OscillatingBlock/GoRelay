package mock

import (
	"GoRelay/internal/models"
	"net/http"
)

/* used to facilitate isolated unit testing by simulating the behavior of dependencies without relying on their real implementations. */

type LoadBalancerUseCaseMock struct {
	SelectBackendFunc      func() *models.Backend
	HandleRequestFunc      func(*http.Request, http.ResponseWriter) error
	GetHealthyBackendsFunc func() int
}

func (m *LoadBalancerUseCaseMock) HandleRequest(req *http.Request, w http.ResponseWriter) error {
	return m.HandleRequestFunc(req, w)
}

func (m *LoadBalancerUseCaseMock) GetHealthyBackends() int {
	return m.GetHealthyBackendsFunc()
}
