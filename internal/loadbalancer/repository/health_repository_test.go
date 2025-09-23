package repository

import (
	"GoRelay/internal/models"
	"GoRelay/pkg/logger"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthRepository_CheckHealth(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler func(w http.ResponseWriter, r *http.Request)
		expectHealthy bool
	}{
		{
			name: "HealthyBackend_200",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectHealthy: true,
		},
		{
			name: "UnhealthyBackend_500",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectHealthy: false,
		},
		{
			name: "NetworkError",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				// Simulate network error by closing connection
				if hj, ok := w.(http.Hijacker); ok {
					conn, _, _ := hj.Hijack()
					conn.Close()
				}
			},
			expectHealthy: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverHandler))
			defer server.Close()

			backend, err := models.NewBackend(server.URL)
			assert.NoError(t, err, "Failed to create backend")

			logger := logger.NewLogger()
			repo := NewHealthRepository(logger)

			isHealthy := repo.CheckHealth(backend)

			assert.Equal(t, tt.expectHealthy, isHealthy, "Unexpected health status")
		})
	}
}
