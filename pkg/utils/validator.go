package utils

import "github.com/go-playground/validator"

func ValidateConfig(cfg *Config) error {
	validate := validator.New()
	err := validate.Struct(cfg)
	return err
}
