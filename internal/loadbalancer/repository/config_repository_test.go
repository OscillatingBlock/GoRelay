package repository

import (
	"GoRelay/pkg/logger"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	logger := logger.NewLogger()
	var b []string
	var expected_time time.Duration
	t.Run("valid config", func(t *testing.T) {
		cfg_repo := NewConfigRepository("validconfig.yaml", logger)
		cfg, err := cfg_repo.Load()
		assert.Nil(t, err, "expected no errors for valid config")
		assert.NotNil(t, cfg.Port, "expected not nil port")
		assert.NotNil(t, cfg.Backends, "expected not nil backends")
		assert.IsType(t, cfg.Backends, b, "expected backend type []string")
		assert.IsType(t, cfg.HealthInterval, expected_time, "expected HealthInterval type time.Duration")
	})

	t.Run("invalid yaml", func(t *testing.T) {
		cfg_repo := NewConfigRepository("invalidconfig.yaml", logger)
		cfg, err := cfg_repo.Load()
		assert.NotNil(t, err, "expected errors for invalid yaml")
		assert.Nil(t, cfg, "expected empty cfg")
	})
}
