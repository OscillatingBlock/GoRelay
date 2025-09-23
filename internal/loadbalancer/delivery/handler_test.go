package handler

import (
	"GoRelay/internal/loadbalancer/mock"
	"GoRelay/internal/models"
	"GoRelay/pkg/http_errors"
	"GoRelay/pkg/logger"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProxyHandler(t *testing.T) {
	t.Run("Healthy Backend 200", func(t *testing.T) {
		uc := &mock.LoadBalancerUseCaseMock{
			HandleRequestFunc: func(req *http.Request, w http.ResponseWriter) error {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Success"))
				return nil
			},
			GetHealthyBackendsFunc: func() int { return 3 },
		}
		handler := &Handler{uc: uc, logger: logger.NewLogger()}

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.ProxyHandler(w, req)

		expectedCode := 200
		assert.Equal(t, expectedCode, w.Code, "expected same response code")
		assert.Equal(t, "Success", w.Body.String(), "expected response body same as sent by server")
	})

	t.Run("no backends", func(t *testing.T) {
		uc := &mock.LoadBalancerUseCaseMock{
			HandleRequestFunc: func(req *http.Request, w http.ResponseWriter) error {
				return http_errors.ErrNoHealthyBackend
			},
			GetHealthyBackendsFunc: func() int { return 0 },
		}
		handler := &Handler{uc: uc, logger: logger.NewLogger()}

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.ProxyHandler(w, req)

		expectedCode := 502
		assert.Equal(t, expectedCode, w.Code, "expected same respone code")
	})

	t.Run("retry logic", func(t *testing.T) {
		// Create backends for state tracking
		backend, _ := models.NewBackend("http://localhost:5001")
		backend.SetAlive(true)
		backend2, _ := models.NewBackend("http://localhost:5002")
		backend2.SetAlive(true)

		// Mock use case
		uc := &mock.LoadBalancerUseCaseMock{
			HandleRequestFunc: func(req *http.Request, w http.ResponseWriter) error {
				// Simulate retry logic: first backend fails, second succeeds
				backend.SetAlive(false) // Mark first backend unhealthy
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Success"))
				return nil
			},
			GetHealthyBackendsFunc: func() int {
				// Return number of healthy backends (after first is marked unhealthy)
				healthy := 0
				if backend.IsAlive() {
					healthy++
				}
				if backend2.IsAlive() {
					healthy++
				}
				return healthy
			},
		}

		handler := &Handler{uc: uc, logger: logger.NewLogger()}

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.ProxyHandler(w, req)

		assert.Equal(t, "Success", w.Body.String(), "Expected response body")
		assert.False(t, backend.IsAlive(), "First backend should be marked unhealthy")
		assert.True(t, backend2.IsAlive(), "Second backend should remain healthy")
		assert.Equal(t, 0, backend.GetActiveConnections(), "First backend should have no active connections")
		assert.Equal(t, 0, backend2.GetActiveConnections(), "Second backend should have no active connections")
		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200")
	})
}

func TestHealthHandler(t *testing.T) {
	logger := logger.NewLogger()

	t.Run("HealthyBackends", func(t *testing.T) {
		uc := &mock.LoadBalancerUseCaseMock{
			HandleRequestFunc: func(r *http.Request, w http.ResponseWriter) error {
				return nil
			},
			GetHealthyBackendsFunc: func() int { return 4 },
		}
		handler := &Handler{uc: uc, logger: logger}

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		handler.HealthHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status OK")
		assert.Equal(t, "4", w.Body.String(), "Expected one healthy backend")
	})

	t.Run("EmptyPool", func(t *testing.T) {
		uc := &mock.LoadBalancerUseCaseMock{
			HandleRequestFunc:      func(r *http.Request, w http.ResponseWriter) error { return nil },
			GetHealthyBackendsFunc: func() int { return 0 },
		}
		handler := &Handler{uc: uc, logger: logger}

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		handler.HealthHandler(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code, "Expected status Service Unavailable")
		assert.Equal(t, "0", w.Body.String(), "Expected zero healthy backends")
	})

}
