package utils

import "time"

type Config struct {
	Port           string        `yaml:"port" validate:"required,numeric"`
	Backends       []string      `yaml:"backends" validate:"required,dive,required"`
	HealthInterval time.Duration `yaml:"healthInterval" validate:"gt=0"`
	Algorithm      string        `yaml:"algorithm" validate:"oneof=roundRobin round_robin leastconn least_conn"`
}
