package job

import (
	"context"
	"sync"
	"time"

	"github.com/leshless/pet/cub/internal/config"
	"github.com/leshless/pet/cub/internal/telemetry"
)

type Port interface {
	Run(ctx context.Context) error
}

// @PublicPointerInstance
type port struct {
	telemetry.Telemetry
	jobsByName       map[string]Job
	jobsConfigByName map[string]config.Job
}

var _ Port = (*port)(nil)

func (p *port) Run(ctx context.Context) error {
	p.Logger.Info(ctx, "starting job port")

	var wg sync.WaitGroup

	for jobName, job := range p.jobsByName {
		// Config should always be present as we validated it during port initialization
		jobConfig := p.jobsConfigByName[jobName]

		wg.Add(1)
		go p.jobExecutor(ctx, job, jobConfig)
	}

	wg.Wait()

	return nil
}

func (p *port) jobExecutor(ctx context.Context, job Job, jobConfig config.Job) {
	ctx = telemetry.ContextWith(ctx, telemetry.JobName(jobConfig.Name))

	p.Logger.Info(ctx, "starting job executor")

	runJob := func() {
		ctx, cancel := context.WithTimeout(ctx, jobConfig.Timeout)

		// TODO: This may potentially lead to huge burst of goroutines if job action wont properly handle context timeout
		// We can take it for now, but some additional logic or goroutine limiting mechanism may be added later
		go func() {
			defer cancel()

			ts := time.Now()

			p.Logger.Info(ctx, "executing job")

			err := job.Exec(ctx)

			latency := time.Since(ts)

			ctx := telemetry.ContextWith(ctx, telemetry.Any("latency", latency))

			p.Registry.Counter(ctx, telemetry.ExecutedJobsTotal, telemetry.Successful(err == nil)).Inc()
			p.Registry.Summary(ctx, telemetry.JobDurationSeconds, telemetry.Successful(err == nil)).Observe(latency.Seconds())

			if err != nil {
				p.Logger.Error(
					ctx,
					"job executed with error",
					telemetry.Error(err),
				)

				return
			}

			p.Logger.Info(ctx, "job executed successfully")
		}()
	}

	if jobConfig.RunOnStartup {
		go runJob()
	}

	ticker := time.NewTicker(jobConfig.Period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			runJob()
		case <-ctx.Done():
			p.Logger.Warn(ctx, "stopping job executor")
			return
		}
	}
}
