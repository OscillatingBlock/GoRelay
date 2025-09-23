package usecase

import (
	"GoRelay/internal/loadbalancer/mock"
	"GoRelay/internal/models"
	"GoRelay/pkg/http_errors"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSelectBackend(t *testing.T) {
	//mock healthChecker
	healthChecker := &mock.HealthRepositoryMock{
		CheckHealthFunc: func(b *models.Backend) bool {
			return b.IsAlive()
		},
	}

	//mock transport
	transport := &http.Transport{}

	// Helper to create a backend
	newBackend := func(rawURL string, alive bool, conns int) *models.Backend {
		b, err := models.NewBackend(rawURL)
		if err != nil {
			t.Fatalf("Failed to create backend: %v", err)
		}
		b.SetAlive(alive)
		for i := 0; i < conns; i++ {
			b.IncrementConnections()
		}
		return b
	}

	tests := []struct {
		name      string
		algorithm string
		backends  []*models.Backend
		expected  *url.URL // Expected backend URL or nil
	}{
		{
			name:      "RoundRobin_SingleHealthy",
			algorithm: RoundRobin,
			backends: []*models.Backend{
				newBackend("http://localhost:5001", true, 1),
				newBackend("http://localhost:5002", false, 2),
			},
			expected: mustParseURL(t, "http://localhost:5001"),
		},
		{
			name:      "RoundRobin_MultipleHealthy",
			algorithm: RoundRobin,
			backends: []*models.Backend{
				newBackend("http://localhost:5001", true, 1),
				newBackend("http://localhost:5002", true, 2),
			},
			expected: mustParseURL(t, "http://localhost:5001"), // First call picks index 0
		},
		{
			name:      "RoundRobin_NoHealthy",
			algorithm: RoundRobin,
			backends: []*models.Backend{
				newBackend("http://localhost:5001", false, 1),
				newBackend("http://localhost:5002", false, 2),
			},
			expected: nil,
		},
		{
			name:      "LeastConnections_SingleHealthy",
			algorithm: LeastConnections,
			backends: []*models.Backend{
				newBackend("http://localhost:5001", true, 1),
				newBackend("http://localhost:5002", false, 0),
			},
			expected: mustParseURL(t, "http://localhost:5001"),
		},
		{
			name:      "LeastConnections_MultipleHealthy",
			algorithm: LeastConnections,
			backends: []*models.Backend{
				newBackend("http://localhost:5001", true, 2),
				newBackend("http://localhost:5002", true, 1),
			},
			expected: mustParseURL(t, "http://localhost:5002"), // Fewest connections
		},
		{
			name:      "LeastConnections_NoHealthy",
			algorithm: LeastConnections,
			backends: []*models.Backend{
				newBackend("http://localhost:5001", false, 1),
				newBackend("http://localhost:5002", false, 0),
			},
			expected: nil,
		},
		{
			name:      "EmptyPool",
			algorithm: RoundRobin,
			backends:  []*models.Backend{},
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Create mock pool
			pool := models.NewServerPool()
			for _, b := range tt.backends {
				pool.AddBackend(b)
			}

			//Initialise Use Case
			uc := NewLoadBalancerUseCase(pool, tt.algorithm, healthChecker, transport)
			selectedBackend := uc.SelectBackend()

			if tt.expected == nil {
				assert.Nil(t, selectedBackend, "Expected no backend selected")
			} else {
				assert.NotNil(t, selectedBackend, "Expected a backend to be selected")
				assert.Equal(t, tt.expected.String(), selectedBackend.URL.String(), "Selected backend URL should match expected")
			}

			// For RoundRobin, test second call to verify cycling
			if tt.algorithm == RoundRobin && len(pool.GetBackends()) > 1 {
				second := uc.SelectBackend()
				// Calculate expected second backend (next healthy one)
				healthyBackends := pool.GetBackends()
				var expectedSecond *url.URL
				for i, b := range healthyBackends {
					if b.URL.String() == selectedBackend.URL.String() {
						// Next healthy backend (cyclic)
						nextIndex := (i + 1) % len(healthyBackends)
						expectedSecond = healthyBackends[nextIndex].URL
						break
					}
				}
				assert.NotNil(t, second, "Second call should select a backend")
				assert.Equal(t, expectedSecond.String(), second.URL.String(), "RoundRobin second call should cycle to next healthy backend")
			}

		})
	}

}

// Helper to parse URL or fail test
func mustParseURL(t *testing.T, rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("Failed to parse URL %s: %v", rawURL, err)
	}
	return u
}

func TestHandleRequests(t *testing.T) {
	tests := []struct {
		name         string
		backends     []*models.Backend
		healthCheck  func(b *models.Backend) bool
		mockServer   func() *httptest.Server // Keep signature as func() *httptest.Server
		expectError  bool
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success_SingleHealthyBackend",
			backends: func() []*models.Backend {
				b, _ := models.NewBackend("http://localhost:5001")
				b.SetAlive(true)
				return []*models.Backend{b}
			}(),
			healthCheck: func(b *models.Backend) bool { return b.IsAlive() },
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("Success"))
				}))
			},
			expectError:  false,
			expectedCode: http.StatusOK,
			expectedBody: "Success",
		},
		{
			name: "NoHealthyBackends",
			backends: func() []*models.Backend {
				b, _ := models.NewBackend("http://localhost:5001")
				b.SetAlive(false)
				return []*models.Backend{b}
			}(),
			healthCheck: func(b *models.Backend) bool { return b.IsAlive() },
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			expectError:  true,
			expectedCode: http.StatusBadGateway,
		},
		{
			name:        "EmptyPool",
			backends:    []*models.Backend{},
			healthCheck: func(b *models.Backend) bool { return false },
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
			},
			expectError:  true,
			expectedCode: http.StatusBadGateway,
		},
		{
			name: "RetryOnFailure",
			backends: func() []*models.Backend {
				b1, _ := models.NewBackend("http://localhost:5001")
				b2, _ := models.NewBackend("http://localhost:5002")
				b1.SetAlive(true)
				b2.SetAlive(true)
				return []*models.Backend{b1, b2}
			}(),
			healthCheck: func(b *models.Backend) bool { return b.IsAlive() },
			mockServer: func() *httptest.Server {
				// Capture backend URLs as constants for the closure
				firstBackend := "http://localhost:5001"
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.String() == firstBackend {
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte("Retry Success"))
					}
				}))
			},
			expectError:  false,
			expectedCode: http.StatusOK,
			expectedBody: "Retry Success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := tt.mockServer()
			defer mockServer.Close()

			for _, b := range tt.backends {
				b.URL, _ = url.Parse(mockServer.URL)
			}

			pool := models.NewServerPool()
			for _, b := range tt.backends {
				pool.AddBackend(b)
			}

			healthChecker := &mock.HealthRepositoryMock{
				CheckHealthFunc: tt.healthCheck,
			}

			transport := &http.Transport{}
			uc := NewLoadBalancerUseCase(pool, "roundrobin", healthChecker, transport)

			// Test request
			//NewRequest creates a new incoming server request
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			err := uc.HandleRequest(req, w)

			// Assertions
			if tt.expectError {
				assert.Error(t, err, "Expected an error")
				assert.Equal(t, http_errors.ErrNoHealthyBackend, err, "Expected ErrNoHealthyBackend")
			} else {
				assert.NoError(t, err, "Expected no error")
				assert.Equal(t, tt.expectedCode, w.Code, "Unexpected status code")
				if tt.expectedBody != "" {
					assert.Equal(t, tt.expectedBody, w.Body.String(), "Unexpected response body")
				}
			}

			// Verify connection tracking
			for _, b := range tt.backends {
				assert.Equal(t, 0, b.GetActiveConnections(), "Active connections should be zero after request")
			}
		})
	}
}

func TestNoHealthyBackends(t *testing.T) {
	healthChecker := &mock.HealthRepositoryMock{
		CheckHealthFunc: func(b *models.Backend) bool { return b.IsAlive() },
	}
	pool := models.NewServerPool()
	transport := &http.Transport{}

	uc := NewLoadBalancerUseCase(pool, "roundrobin", healthChecker, transport)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	_ = uc.HandleRequest(req, w)

	assert.Equal(t, 502, w.Code, "Expected response 502")
}

func TestHealthChecks(t *testing.T) {
	healthChecker := &mock.HealthRepositoryMock{
		CheckHealthFunc: func(b *models.Backend) bool {
			return b.URL.String() == "http://localhost:5001"
		},
	}
	backend, _ := models.NewBackend("http://localhost:5001")
	backend.SetAlive(false)
	backend2, _ := models.NewBackend("http://localhost:5003")
	backend2.SetAlive(true)
	pool := models.NewServerPool()
	pool.AddBackend(backend)
	pool.AddBackend(backend2)
	transport := &http.Transport{}

	uc := NewLoadBalancerUseCase(pool, "roundrobin", healthChecker, transport)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	go uc.StartHealthChecksWithContext(ctx, 100*time.Millisecond)

	<-ctx.Done() // Wait for health checks to complete

	assert.True(t, backend.IsAlive(), "Backend 5001 should be marked healthy")
	assert.False(t, backend2.IsAlive(), "Backend 5003 should be marked unhealthy")
}
