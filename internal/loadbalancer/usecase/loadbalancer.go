package usecase

import (
	"GoRelay/internal/models"
	"GoRelay/pkg/http_errors"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"sync/atomic"
	"time"
)

const (
	RoundRobin       string = "roundrobin"
	LeastConnections string = "leastconnections"
)

type LoadBalancerUseCase struct {
	Pool      *models.ServerPool
	algorithm string
	health    HealthChecker
	proxy     *httputil.ReverseProxy
	transport *http.Transport
}

func NewLoadBalancerUseCase(pool *models.ServerPool, algorithm string, health HealthChecker, transport *http.Transport) *LoadBalancerUseCase {
	uc := &LoadBalancerUseCase{
		Pool:      pool,
		algorithm: strings.ToLower(algorithm),
		health:    health,
		transport: transport,
	}
	uc.proxy = &httputil.ReverseProxy{
		/* Director: A ReverseProxy function that rewrites req.URL to route to a
		backend selected by SelectBackend(). It’s low-level, called implicitly by ReverseProxy. */

		// Sets req.URL.Scheme and req.URL.Host to route to the chosen backend.
		Director: func(req *http.Request) {
			backend, ok := req.Context().Value("backend").(*models.Backend)
			if !ok || backend == nil {
				// Fallback (shouldn’t happen with proper Proxy call)
				return
			}
			req.URL.Scheme = backend.URL.Scheme
			req.URL.Host = backend.URL.Host
			req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			w.WriteHeader(http.StatusBadGateway)
		},
		Transport: transport,
	}
	return uc
}

type statusRecorder struct {
	statusCode int
	http.ResponseWriter
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

/*Proxy Forwards the request to director via ServeHTTP and tracks the response using recorder */
func (uc *LoadBalancerUseCase) Proxy(req *http.Request, w http.ResponseWriter, backend *models.Backend) error {
	if backend == nil {
		return http_errors.ErrNoHealthyBackend
	}
	ctx := context.WithValue(req.Context(), "backend", backend)
	*req = *req.WithContext(ctx)
	recorder := &statusRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	uc.proxy.ServeHTTP(recorder, req)
	if recorder.statusCode < 200 || recorder.statusCode >= 300 {
		return fmt.Errorf("backend returned non-2xx status: %d", recorder.statusCode)
	}
	return nil
}

func (uc *LoadBalancerUseCase) roundRobinSelection() *models.Backend {
	healthyBackends := uc.Pool.GetBackends()
	if len(healthyBackends) == 0 {
		return nil
	}
	index := atomic.LoadUint64(&uc.Pool.CurrentIndex)
	atomic.AddUint64(&uc.Pool.CurrentIndex, 1)
	return healthyBackends[index%uint64(len(healthyBackends))]
}

func (uc *LoadBalancerUseCase) leastConnsSelection() *models.Backend {
	min := math.MaxInt64
	var chosen *models.Backend
	for _, b := range uc.Pool.Backends {
		if b.IsAlive() && b.GetActiveConnections() < min {
			min = b.GetActiveConnections()
			chosen = b
		}
	}
	return chosen
}

func (uc *LoadBalancerUseCase) SelectBackend() *models.Backend {
	switch strings.ToLower(uc.algorithm) {
	case RoundRobin:
		return uc.roundRobinSelection()
	case LeastConnections:
		return uc.leastConnsSelection()
	default:
		return uc.roundRobinSelection()
	}
}

/*
	HandleRequest is the use case’s main method for processing requests, called by the HTTP handler.

It manages retries, connection tracking, and passive health checks, invoking Proxy to forward requests via the single ReverseProxy
*/
func (uc *LoadBalancerUseCase) HandleRequest(req *http.Request, w http.ResponseWriter) error {
	for range 3 { // Retry up to 3 times
		backend := uc.SelectBackend()
		if backend == nil {
			w.WriteHeader(http.StatusBadGateway)
			return http_errors.ErrNoHealthyBackend
		}
		fmt.Println(backend.URL)
		backend.IncrementConnections()

		// Use a new ResponseRecorder for each retry
		tempRecorder := httptest.NewRecorder()
		err := uc.Proxy(req, tempRecorder, backend)
		backend.DecrementConnections()
		if err == nil {
			// Copy successful response to original ResponseWriter
			w.WriteHeader(tempRecorder.Code)
			io.Copy(w, tempRecorder.Body)
			return nil
		}
		backend.SetAlive(false)
	}
	w.WriteHeader(http.StatusBadGateway)
	return http_errors.ErrNoHealthyBackend
}

func (uc *LoadBalancerUseCase) StartHealthChecksWithContext(ctx context.Context, interval time.Duration) {
	for _, b := range uc.Pool.Backends {
		go func(backend *models.Backend) {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					alive := uc.health.CheckHealth(backend)
					backend.SetAlive(alive)
				case <-ctx.Done():
					return
				}
			}
		}(b)
	}
}

func (uc *LoadBalancerUseCase) GetHealthyBackends() int {
	return len(uc.Pool.GetBackends())
}
