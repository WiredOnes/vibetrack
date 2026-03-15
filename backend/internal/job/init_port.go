package job

import (
	"context"
	"errors"

	"github.com/leshless/golibrary/set"
	"github.com/leshless/golibrary/sets"
	"github.com/leshless/golibrary/xmaps"
	"github.com/leshless/pet/cub/internal/config"
	"github.com/leshless/pet/cub/internal/telemetry"
)

func InitPort(
	logger telemetry.Logger,
	tel telemetry.Telemetry,
	configHolder config.Holder,
	updateHealthStatusJob UpdateHealthStatus,
) (*port, error) {
	logger.Info(context.Background(), "initializing job port")

	jobsByName := make(map[string]Job)
	jobsByName[UpdateHealthStatusJobName] = updateHealthStatusJob

	jobNames := set.FromSlice(xmaps.Keys(jobsByName))

	if len(jobNames) != len(jobsByName) {
		logger.Error(context.Background(), "duplicate job names found in source code")
		return nil, errors.New("duplicated job name in source code")
	}

	jobConfigs := configHolder.Config().Jobs

	jobConfigsByName := make(map[string]config.Job)
	for _, jobConfig := range jobConfigs {
		jobConfigsByName[jobConfig.Name] = jobConfig
	}

	jobConfigNames := set.FromSlice(xmaps.Keys(jobConfigsByName))

	if len(jobConfigNames) != len(jobConfigs) {
		logger.Error(context.Background(), "duplicate job names found in config")
		return nil, errors.New("duplicated job name in config")
	}

	if len(sets.SymDiff(jobConfigNames, jobNames)) != 0 {
		logger.Error(context.Background(), "job names in config and source code do not match")
		return nil, errors.New("job names in config and source code do not match")
	}

	port := NewPort(
		tel,
		jobsByName,
		jobConfigsByName,
	)

	logger.Info(context.Background(), "job port sucessfully initialized")

	return port, nil
}
