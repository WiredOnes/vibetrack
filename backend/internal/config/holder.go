package config

import (
	"fmt"
	iofs "io/fs"

	"github.com/go-playground/validator/v10"
	"go.yaml.in/yaml/v3"
)

const (
	configFilePath = "etc/cub/config.yml"
)

type Holder interface {
	Config() *Config
}

// @PublicPointerInstance
type holder struct {
	config *Config
}

var _ Holder = (*holder)(nil)

func InitHolder(fs iofs.FS) (*holder, error) {
	data, err := iofs.ReadFile(fs, configFilePath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling config file: %w", err)
	}

	if err := validator.New().Struct(config); err != nil {
		return nil, fmt.Errorf("Config is invalid: %w", err)
	}

	return NewHolder(&config), nil
}

func (h *holder) Config() *Config {
	return h.config
}
