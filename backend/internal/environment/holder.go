package environment

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
)

type Holder interface {
	Environment() *Environment
}

// @PublicPointerInstance
type holder struct {
	environment *Environment
}

var _ Holder = (*holder)(nil)

func InitHolder() (*holder, error) {
	var environment Environment

	err := env.Parse(&environment)
	if err != nil {
		return nil, fmt.Errorf("Getting variables from the environment: %w", err)
	}

	if err := validator.New().Struct(environment); err != nil {
		return nil, fmt.Errorf("Environment is invalid: %w", err)
	}

	return NewHolder(&environment), nil
}

func (h *holder) Environment() *Environment {
	return h.environment
}
