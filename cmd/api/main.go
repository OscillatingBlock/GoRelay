package main

import (
	handler "GoRelay/internal/loadbalancer/delivery"
	"GoRelay/internal/loadbalancer/repository"
	"GoRelay/internal/loadbalancer/usecase"
	"GoRelay/internal/models"
	"GoRelay/internal/server"
	"GoRelay/pkg/logger"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log := logger.NewLogger()
	cfg_repo := repository.NewConfigRepository("configs/config.yaml", log)

	cfg, err := cfg_repo.Load()
	if err != nil {
		log.Error("Error while loading config ", "error", err)
		os.Exit(1)
	}

	health_repo := repository.NewHealthRepository(log)

	pool := models.NewServerPool()
	for _, b := range cfg.Backends {
		backend, err := models.NewBackend(b)
		if err != nil {
			log.Error("Error while getting new backend", "url", b)
			os.Exit(1)
		}
		pool.AddBackend(backend)
	}

	transport := &http.Transport{}
	uc := usecase.NewLoadBalancerUseCase(pool, cfg.Algorithm, health_repo, transport)

	go uc.StartHealthChecksWithContext(context.Background(), cfg.HealthInterval)

	h := handler.NewHandler(uc, log)
	route_cfg := handler.NewRouteConfig(h)
	srv := server.NewServer(route_cfg, log)

	go func() {
		if err := srv.Start(cfg.Port); err != nil && err != http.ErrServerClosed {
			log.Error("skill issue", "error", err)
			os.Exit(1)
		}
	}()
	log.Info("server started", "port", cfg.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", "error", err)
	}

	log.Info("server exited gracefully")
}
